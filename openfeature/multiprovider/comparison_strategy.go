package multiprovider

import (
	"cmp"
	"context"
	"errors"
	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/sync/errgroup"
	"reflect"
	"slices"
	"strings"
)

type (
	ComparisonStrategy struct {
		providers        []*NamedProvider
		fallbackProvider of.FeatureProvider
		customComparator Comparator
		alwaysUseCustom  bool
	}

	Comparator func(values []any) bool
)

var _ Strategy = (*ComparisonStrategy)(nil)

func NewComparisonStrategy(providers []*NamedProvider, fallbackProvider of.FeatureProvider, comparator Comparator, alwaysUseCustom bool) *ComparisonStrategy {
	return &ComparisonStrategy{
		providers:        providers,
		fallbackProvider: fallbackProvider,
		customComparator: comparator,
		alwaysUseCustom:  alwaysUseCustom,
	}
}

func (c ComparisonStrategy) Name() EvaluationStrategy {
	return StrategyComparison
}

func (c ComparisonStrategy) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
	compFunc := func(values []any) bool {
		current := values[0].(bool)
		match := true
		for i, v := range values {
			if i == 0 {
				continue
			}
			if current != v.(bool) {
				match = false
				break
			}
		}

		return match
	}
	result := evaluateComparison[bool](ctx, c.providers, flag, of.Boolean, defaultValue, evalCtx, compFunc, c.fallbackProvider)
	return of.BoolResolutionDetail{
		Value:                    result.Value.(bool),
		ProviderResolutionDetail: result.ProviderResolutionDetail,
	}
}

func buildComparator[R cmp.Ordered]() Comparator {
	return func(values []any) bool {
		v := make([]R, 0, len(values))
		for _, val := range values {
			v = append(v, val.(R))
		}
		slices.Sort(v)
		v = slices.Compact(v)
		return len(v) == 1
	}
}

func (c ComparisonStrategy) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
	compFunc := buildComparator[string]()
	result := evaluateComparison[string](ctx, c.providers, flag, of.String, defaultValue, evalCtx, compFunc, c.fallbackProvider)
	return of.StringResolutionDetail{
		Value:                    result.Value.(string),
		ProviderResolutionDetail: result.ProviderResolutionDetail,
	}
}

func (c ComparisonStrategy) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
	compFunc := buildComparator[float64]()
	result := evaluateComparison[float64](ctx, c.providers, flag, of.Float, defaultValue, evalCtx, compFunc, c.fallbackProvider)
	return of.FloatResolutionDetail{
		Value:                    result.Value.(float64),
		ProviderResolutionDetail: result.ProviderResolutionDetail,
	}
}

func (c ComparisonStrategy) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
	compFunc := buildComparator[int64]()
	result := evaluateComparison[int64](ctx, c.providers, flag, of.Int, defaultValue, evalCtx, compFunc, c.fallbackProvider)
	return of.IntResolutionDetail{
		Value:                    result.Value.(int64),
		ProviderResolutionDetail: result.ProviderResolutionDetail,
	}
}

func (c ComparisonStrategy) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	var compFunc Comparator
	switch defaultValue.(type) {
	case int8:
		compFunc = func(values []any) bool {
			v := make([]int8, 0, len(values))
			for _, val := range values {
				v = append(v, val.(int8))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case int16:
		compFunc = func(values []any) bool {
			v := make([]int16, 0, len(values))
			for _, val := range values {
				v = append(v, val.(int16))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case int32:
		compFunc = func(values []any) bool {
			v := make([]int32, 0, len(values))
			for _, val := range values {
				v = append(v, val.(int32))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case int64:
		compFunc = func(values []any) bool {
			v := make([]int64, 0, len(values))
			for _, val := range values {
				v = append(v, val.(int64))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case int:
		compFunc = func(values []any) bool {
			v := make([]int, 0, len(values))
			for _, val := range values {
				v = append(v, val.(int))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uint8:
		compFunc = func(values []any) bool {
			v := make([]uint8, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uint8))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uint16:
		compFunc = func(values []any) bool {
			v := make([]uint16, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uint16))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uint32:
		compFunc = func(values []any) bool {
			v := make([]uint32, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uint32))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uint64:
		compFunc = func(values []any) bool {
			v := make([]uint64, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uint64))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uint:
		compFunc = func(values []any) bool {
			v := make([]uint, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uint))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case uintptr:
		compFunc = func(values []any) bool {
			v := make([]uintptr, 0, len(values))
			for _, val := range values {
				v = append(v, val.(uintptr))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case float32:
		compFunc = func(values []any) bool {
			v := make([]float32, 0, len(values))
			for _, val := range values {
				v = append(v, val.(float32))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case float64:
		compFunc = func(values []any) bool {
			v := make([]float64, 0, len(values))
			for _, val := range values {
				v = append(v, val.(float64))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	case string:
		compFunc = func(values []any) bool {
			v := make([]string, 0, len(values))
			for _, val := range values {
				v = append(v, val.(string))
			}
			slices.Sort(v)
			v = slices.Compact(v)
			return len(v) == 1
		}
	default:
		t := reflect.TypeOf(defaultValue)
		if t.Comparable() && !c.alwaysUseCustom {
			compFunc = func(vals []any) bool {
				set := map[any]any{}
				for _, v := range vals {
					set[v] = 0
				}

				return len(set) == 1
			}
		} else if c.customComparator == nil && c.alwaysUseCustom {
			panic(nil) // runtime panic -- this state should be impossible
		} else if c.customComparator != nil {
			compFunc = c.customComparator
		} else {
			// Impossible to evaluate strategy with expected result type
			defaultResult := BuildDefaultResult[any](StrategyComparison, defaultValue, errors.New(ErrAggregationNotAllowedText))
			defaultResult.FlagMetadata[MetadataFallbackUsed] = false
			defaultResult.FlagMetadata[MetadataIsDefaultValue] = true
			return defaultResult
		}
	}

	return evaluateComparison[any](ctx, c.providers, flag, of.Object, defaultValue, evalCtx, compFunc, c.fallbackProvider)
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

func evaluateComparison[R any](ctx context.Context, providers []*NamedProvider, flag string, flagType of.Type, defaultVal R, evalCtx of.FlattenedContext, comp Comparator, fallbackProvider of.FeatureProvider) of.InterfaceResolutionDetail {
	// Short circuit if there's only one provider as no comparison nor workers are needed
	if len(providers) == 1 {
		result := evaluate(ctx, providers[0], flag, flagType, defaultVal, evalCtx)
		metadata := setFlagMetadata(StrategyComparison, providers[0].Name, make(of.FlagMetadata))
		metadata[MetadataFallbackUsed] = false
		result.FlagMetadata = mergeFlagMeta(result.FlagMetadata, metadata)
		return result
	}

	type namedResult struct {
		name string
		*of.InterfaceResolutionDetail
	}

	resultChan := make(chan *namedResult, len(providers))
	notFoundChan := make(chan interface{})
	errGrp, ctx := errgroup.WithContext(ctx)
	for _, provider := range providers {
		closedProvider := provider
		errGrp.Go(func() error {
			localChan := make(chan *namedResult)

			go func(c context.Context, p *NamedProvider) {
				result := evaluate(ctx, p, flag, flagType, defaultVal, evalCtx)
				localChan <- &namedResult{
					name:                      p.Name,
					InterfaceResolutionDetail: &result,
				}
			}(ctx, closedProvider)

			select {
			case r := <-localChan:
				notFound := r.ResolutionDetail().ErrorCode == of.FlagNotFoundCode
				if !notFound && r.Error() != nil {
					return &ProviderError{
						ProviderName: r.name,
						Err:          r.Error(),
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
	resultValues := make([]R, 0, len(providers))
	notFoundCount := 0
	for {
		select {
		case <-ctx.Done():
			// Error occurred
			result := BuildDefaultResult(StrategyComparison, defaultVal, ctx.Err())
			result.FlagMetadata[MetadataFallbackUsed] = false
			result.FlagMetadata[MetadataIsDefaultValue] = true
			result.FlagMetadata[MetadataEvaluationError] = ctx.Err().Error()
			result.ResolutionError = comparisonResolutionError(result.FlagMetadata)
			return result
		case r := <-resultChan:
			results = append(results, *r)
			resultValues = append(resultValues, r.Value.(R))
			if (len(results) + notFoundCount) == len(providers) {
				// All results accounted for
				goto continueComparison
			}
		case <-notFoundChan:
			notFoundCount += 1
			if notFoundCount == len(providers) {
				result := BuildDefaultResult(StrategyComparison, defaultVal, nil)
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
		metadata[r.name] = r.FlagMetadata
	}
	resultsForComparison := make([]any, 0, len(resultValues))
	for _, r := range resultValues {
		resultsForComparison = append(resultsForComparison, r)
	}
	if comp(resultsForComparison) {
		metadata[MetadataFallbackUsed] = false
		metadata[MetadataIsDefaultValue] = false
		metadata[MetadataComparisonDisagreeingProviders] = []string{}
		success := make([]string, 0, len(providers))
		variants := make([]string, 0, len(providers))
		// Gather metadata from provider results
		for _, r := range results {
			metadata[r.name] = r.FlagMetadata
			success = append(success, r.name)
			variants = append(variants, r.Variant)
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
		return of.InterfaceResolutionDetail{
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
			flagType,
			defaultVal,
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

	defaultResult := BuildDefaultResult[R](StrategyComparison, defaultVal, errors.New("no fallback provider configured"))
	mergeFlagMeta(defaultResult.FlagMetadata, metadata)
	defaultResult.FlagMetadata[MetadataFallbackUsed] = false
	defaultResult.FlagMetadata[MetadataIsDefaultValue] = true

	return defaultResult
}
