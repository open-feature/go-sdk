package multiprovider

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
)

// firstMatchStrategy
type firstMatchStrategy struct {
	providers []*NamedProvider
}

var _ Strategy = (*firstMatchStrategy)(nil)

// NewFirstMatchStrategy Creates a new firstMatchStrategy instance as a Strategy. This strategy will execute providers
// sequentially in the order they are provided until one returns a non-error result. If no providers return a successful
// result then the default value will be used.
func NewFirstMatchStrategy(providers []*NamedProvider) Strategy {
	return &firstMatchStrategy{providers: providers}
}

func (f *firstMatchStrategy) Name() EvaluationStrategy {
	return StrategyFirstMatch
}

func (f *firstMatchStrategy) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
	res := evaluateFirstMatch[bool](ctx, f.providers, flag, of.Boolean, defaultValue, evalCtx)
	return of.BoolResolutionDetail{
		Value:                    res.Value.(bool),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstMatchStrategy) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
	res := evaluateFirstMatch[string](ctx, f.providers, flag, of.String, defaultValue, evalCtx)
	return of.StringResolutionDetail{
		Value:                    res.Value.(string),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstMatchStrategy) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
	res := evaluateFirstMatch[float64](ctx, f.providers, flag, of.Float, defaultValue, evalCtx)
	return of.FloatResolutionDetail{
		Value:                    res.Value.(float64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstMatchStrategy) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
	res := evaluateFirstMatch[int64](ctx, f.providers, flag, of.Int, defaultValue, evalCtx)
	return of.IntResolutionDetail{
		Value:                    res.Value.(int64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstMatchStrategy) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	return evaluateFirstMatch[any](ctx, f.providers, flag, of.Object, defaultValue, evalCtx)
}

// evaluateFirstMatch execute the first match strategy
func evaluateFirstMatch[R any](ctx context.Context, providers []*NamedProvider, flag string, flagType of.Type, defaultVal R, flatCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	for _, provider := range providers {
		resolution := evaluate[R](ctx, provider, flag, flagType, defaultVal, flatCtx)
		if resolution.Error() != nil && resolution.ResolutionDetail().ErrorCode == of.FlagNotFoundCode {
			continue
		}

		if resolution.Error() != nil {
			resolution.FlagMetadata = mergeFlagMeta(resolution.FlagMetadata, of.FlagMetadata{
				MetadataSuccessfulProviderName: "none",
				MetadataStrategyUsed:           StrategyFirstMatch,
			})
			// Stop evaluation if an error occurs
			return resolution
		}

		// success!
		resolution.FlagMetadata = setFlagMetadata(StrategyFirstMatch, provider.Name, resolution.FlagMetadata)
		return resolution
	}

	return BuildDefaultResult[R](StrategyFirstMatch, defaultVal, nil)
}
