package multi

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"

	of "github.com/open-feature/go-sdk/openfeature"
)

// ErrAggregationNotAllowed is an error returned if [of.FeatureProvider.ObjectEvaluation] is called using the [StrategyComparison]
// strategy without a custom [Comparator] function configured when response objects are not comparable.
var ErrAggregationNotAllowed = errors.New(errAggregationNotAllowedText)

// Comparator is used to compare the results of [of.FeatureProvider.ObjectEvaluation].
// This is required if returned results are not comparable.
type Comparator func(values []any) bool

// newComparisonStrategy returns a [StrategyComparison] strategy function. The fallback provider specified is called when
// there is a comparison failure -- prior to returning a default result. The [Comparator] parameter is optional and nil
// can be passed as long as ObjectEvaluation is never called with objects that are not comparable. The custom [Comparator]
// will only be used for [of.FeatureProvider.ObjectEvaluation] if set. If [of.FeatureProvider.ObjectEvaluation] is
// called without setting a [Comparator], and the returned object(s) are not comparable, then an error will occur.
func newComparisonStrategy(providers []NamedProvider, fallbackProvider of.FeatureProvider, comparator Comparator, runMode runModeFn[FlagTypes]) StrategyFn[FlagTypes] {
	return evaluateComparison(providers, fallbackProvider, comparator, runMode)
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
		if current == nil {
			return false // nilable values are not comparable
		}
		t := reflect.TypeOf(current)
		if t.Comparable() {
			set := map[any]struct{}{}
			for _, v := range values {
				if v == nil {
					return false // nil is not comparable
				}
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

func evaluateComparison[T FlagTypes](providers []NamedProvider, fallbackProvider of.FeatureProvider, comparator Comparator, runMode runModeFn[T]) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, evalCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
		if comparator == nil {
			comparator = defaultComparator
			switch any(defaultValue).(type) {
			case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, uintptr, float32, float64, string, bool:
				break
			default:
				t := reflect.TypeOf(defaultValue)
				if !t.Comparable() {
					// Impossible to evaluate strategy with expected result type
					defaultResult := BuildDefaultResult(StrategyComparison, defaultValue, ErrAggregationNotAllowed)
					defaultResult.FlagMetadata[MetadataFallbackUsed] = false
					defaultResult.FlagMetadata[MetadataIsDefaultValue] = true
					return defaultResult
				}
			}
		}

		// Short circuit if there's only one provider as no comparison nor workers are needed
		if len(providers) == 1 {
			result := Evaluate(ctx, providers[0], flag, defaultValue, evalCtx)
			metadata := setFlagMetadata(StrategyComparison, providers[0].Name(), make(of.FlagMetadata))
			metadata[MetadataFallbackUsed] = false
			result.FlagMetadata = mergeFlagMeta(result.FlagMetadata, metadata)
			return result
		}

		type namedResult struct {
			name string
			res  *of.GenericResolutionDetail[T]
		}

		results := make([]namedResult, 0, len(providers))
		resultValues := make([]T, 0, len(providers))
		notFoundCount := 0
		for name, result := range runMode(ctx, providers, flag, defaultValue, evalCtx) {
			notFound := result.ResolutionDetail().ErrorCode == of.FlagNotFoundCode
			if !notFound && result.Error() != nil {
				result := BuildDefaultResult(StrategyComparison, defaultValue, result.Error())
				result.FlagMetadata[MetadataFallbackUsed] = false
				result.FlagMetadata[MetadataIsDefaultValue] = true
				result.FlagMetadata[MetadataEvaluationError] = result.Error()
				result.ResolutionError = comparisonResolutionError(result.FlagMetadata)
				return result
			}
			if !notFound {
				r := namedResult{
					name: name,
					res:  &result,
				}
				results = append(results, r)
				resultValues = append(resultValues, r.res.Value)
			} else {
				notFoundCount++
			}
		}

		if notFoundCount == len(providers) {
			result := BuildDefaultResult(StrategyComparison, defaultValue, nil)
			result.FlagMetadata[MetadataFallbackUsed] = false
			result.FlagMetadata[MetadataIsDefaultValue] = true
			result.ResolutionError = comparisonResolutionError(result.FlagMetadata)
			return result
		}

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
			metadata[MetadataSuccessfulProviderNames] = strings.Join(success, ", ")
			// Unique values only
			slices.Sort(variants)
			variants = slices.Compact(variants)
			var variantResults string
			if len(variants) == 1 {
				variantResults = variants[0]
			} else {
				variantResults = strings.Join(variants, ", ")
			}
			return of.GenericResolutionDetail[T]{
				Value: resultValues[0], // All values should be equal
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:       ReasonAggregated,
					Variant:      variantResults,
					FlagMetadata: metadata,
				},
			}
		}

		if fallbackProvider != nil {
			fallbackResult := Evaluate(
				ctx,
				&namedProvider{name: "fallback", FeatureProvider: fallbackProvider},
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
		defaultResult.FlagMetadata = mergeFlagMeta(defaultResult.FlagMetadata, metadata)
		defaultResult.FlagMetadata[MetadataFallbackUsed] = false
		defaultResult.FlagMetadata[MetadataIsDefaultValue] = true
		return defaultResult
	}
}
