package multiprovider

import (
	"context"
	"errors"
	reflect "reflect"
	"slices"
	"strings"

	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/sync/errgroup"
)

type Comparator func(values []any) bool

// NewComparisonStrategy creates a new instance of ComparisonStrategy. The fallback provider specified is called when
// there is a comparison failure -- prior to returning a default result. The Comparator parameter is optional and nil
// can be passed as long as ObjectEvaluation is never called. Unless the `alwaysUseCustom` parameter is true the default
// comparisons built into Go will be used. The custom Comparator will only be used for ObjectEvaluation. However, if the
// parameter is set to true the custom Comparator will always be used regardless of evaluation type. If ObjectEvaluation
// is called without setting a Comparator and the returned object is not `comparable` then the a panic will occur. A
// panic will always occur if the Comparator is nil, but `alwaysUseCustom` is true.
func NewComparisonStrategy(providers []*NamedProvider, fallbackProvider of.FeatureProvider, comparator Comparator) StrategyFn[FlagTypes] {
	return evaluateComparison[FlagTypes](providers, fallbackProvider, comparator)
}

func defaultComparator(values []any) bool {
	if len(values) == 0 {
		return false
	}
	current := values[0]

	switch current.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, uintptr, float32, float64, string, bool:
		for i, v := range values {
			if i == 0 {
				continue
			}
			if v != current {
				return false
			}
		}
		return true
	default:
		t := reflect.TypeOf(current)
		if t.Comparable() {
			set := map[any]struct{}{}
			for _, v := range values {
				set[v] = struct{}{}
			}

			return len(set) == 1
		}
		return false
	}
}

func comparisonResolutionError(metadata of.FlagMetadata) of.ResolutionError {
	if isDefault, err := metadata.GetBool(MetadataIsDefaultValue); err != nil || !isDefault {
		return of.ResolutionError{}
	}

	if notFound, err := metadata.GetString(MetadataSuccessfulProviderName); err == nil && notFound == "none" {
		return of.NewFlagNotFoundResolutionError("not found in any providers")
	}

	if evalErr, err := metadata.GetString(MetadataEvaluationError); err == nil && evalErr != "" {
		return of.NewGeneralResolutionError(evalErr)
	}

	return of.NewGeneralResolutionError("comparison failure")
}

func evaluateComparison[T FlagTypes](providers []*NamedProvider, fallbackProvider of.FeatureProvider, comparator Comparator) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, evalCtx of.FlattenedContext) GeneralResolutionDetail[T] {
		if comparator == nil {
			comparator = defaultComparator
			switch any(defaultValue).(type) {
			case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, uintptr, float32, float64, string, bool:
				break
			default:
				t := reflect.TypeOf(defaultValue)
				if !t.Comparable() {
					// Impossible to evaluate strategy with expected result type
					defaultResult := BuildDefaultResult(StrategyComparison, defaultValue, errors.New(ErrAggregationNotAllowedText))
					defaultResult.FlagMetadata[MetadataFallbackUsed] = false
					defaultResult.FlagMetadata[MetadataIsDefaultValue] = true
					return defaultResult
				}
			}
		}

		// Short circuit if there's only one provider as no comparison nor workers are needed
		if len(providers) == 1 {
			result := evaluate(ctx, providers[0], flag, defaultValue, evalCtx)
			metadata := setFlagMetadata(StrategyComparison, providers[0].Name, make(of.FlagMetadata))
			metadata[MetadataFallbackUsed] = false
			result.FlagMetadata = mergeFlagMeta(result.FlagMetadata, metadata)
			return result
		}

		type namedResult struct {
			name string
			res  *GeneralResolutionDetail[T]
		}

		resultChan := make(chan *namedResult, len(providers))
		notFoundChan := make(chan any)
		errGrp, ctx := errgroup.WithContext(ctx)
		for _, provider := range providers {
			closedProvider := provider
			errGrp.Go(func() error {
				localChan := make(chan *namedResult)

				go func(c context.Context, p *NamedProvider) {
					result := evaluate(ctx, p, flag, defaultValue, evalCtx)
					localChan <- &namedResult{
						name: p.Name,
						res:  &result,
					}
				}(ctx, closedProvider)

				select {
				case r := <-localChan:
					notFound := r.res.ResolutionDetail().ErrorCode == of.FlagNotFoundCode
					if !notFound && r.res.Error() != nil {
						return &ProviderError{
							ProviderName: r.name,
							Err:          r.res.Error(),
						}
					}
					if !notFound {
						resultChan <- r
					} else {
						notFoundChan <- struct{}{}
					}
					return nil
				case <-ctx.Done():
					return nil
				}
			})
		}

		results := make([]namedResult, 0, len(providers))
		resultValues := make([]T, 0, len(providers))
		notFoundCount := 0
		for {
			select {
			case <-ctx.Done():
				// Error occurred
				result := BuildDefaultResult(StrategyComparison, defaultValue, ctx.Err())
				result.FlagMetadata[MetadataFallbackUsed] = false
				result.FlagMetadata[MetadataIsDefaultValue] = true
				result.FlagMetadata[MetadataEvaluationError] = ctx.Err().Error()
				result.ResolutionError = comparisonResolutionError(result.FlagMetadata)
				return result
			case r := <-resultChan:
				results = append(results, *r)
				resultValues = append(resultValues, r.res.Value)
				if (len(results) + notFoundCount) == len(providers) {
					// All results accounted for
					goto continueComparison
				}
			case <-notFoundChan:
				notFoundCount += 1
				if notFoundCount == len(providers) {
					result := BuildDefaultResult(StrategyComparison, defaultValue, nil)
					result.FlagMetadata[MetadataFallbackUsed] = false
					result.FlagMetadata[MetadataIsDefaultValue] = true
					result.ResolutionError = comparisonResolutionError(result.FlagMetadata)
					return result
				}
				if (len(results) + notFoundCount) == len(providers) {
					// All results accounted for
					goto continueComparison
				}
			}
		}
	continueComparison:
		// Evaluate Results Are Equal
		metadata := make(of.FlagMetadata)
		metadata[MetadataStrategyUsed] = StrategyComparison
		// Build Aggregate metadata key'd by their names of all Providers
		for _, r := range results {
			metadata[r.name] = r.res.FlagMetadata
		}
		resultsForComparison := make([]any, 0, len(resultValues))
		for _, r := range resultValues {
			resultsForComparison = append(resultsForComparison, r)
		}
		if comparator(resultsForComparison) {
			metadata[MetadataFallbackUsed] = false
			metadata[MetadataIsDefaultValue] = false
			metadata[MetadataComparisonDisagreeingProviders] = []string{}
			success := make([]string, 0, len(providers))
			variants := make([]string, 0, len(providers))
			// Gather metadata from provider results
			for _, r := range results {
				metadata[r.name] = r.res.FlagMetadata
				success = append(success, r.name)
				variants = append(variants, r.res.Variant)
			}
			// maintain stable order of metadata results
			slices.Sort(success)
			metadata[MetadataSuccessfulProviderName+"s"] = strings.Join(success, ", ")
			// Unique values only
			slices.Sort(variants)
			variants = slices.Compact(variants)
			var variantResults string
			if len(variants) == 1 {
				variantResults = variants[0]
			} else {
				variantResults = strings.Join(variants, ", ")
			}
			return GeneralResolutionDetail[T]{
				Value: resultValues[0], // All values should be equal
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:       ReasonAggregated,
					Variant:      variantResults,
					FlagMetadata: metadata,
				},
			}
		}

		if fallbackProvider != nil {
			fallbackResult := evaluate(
				ctx,
				&NamedProvider{Name: "fallback", FeatureProvider: fallbackProvider},
				flag,
				defaultValue,
				evalCtx,
			)
			fallbackResult.FlagMetadata = mergeFlagMeta(fallbackResult.FlagMetadata, metadata)
			fallbackResult.FlagMetadata[MetadataFallbackUsed] = true
			fallbackResult.FlagMetadata[MetadataIsDefaultValue] = false
			fallbackResult.FlagMetadata[MetadataSuccessfulProviderName] = "fallback"
			fallbackResult.FlagMetadata[MetadataStrategyUsed] = StrategyComparison
			fallbackResult.Reason = ReasonAggregatedFallback
			return fallbackResult
		}

		defaultResult := BuildDefaultResult(StrategyComparison, defaultValue, errors.New("no fallback provider configured"))
		mergeFlagMeta(defaultResult.FlagMetadata, metadata)
		defaultResult.FlagMetadata[MetadataFallbackUsed] = false
		defaultResult.FlagMetadata[MetadataIsDefaultValue] = true

		return defaultResult
	}
}
