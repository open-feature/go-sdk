package multiprovider

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
	"time"
)

type firstSuccessStrategy struct {
	providers []*NamedProvider
	timeout   time.Duration
}

var _ Strategy = (*firstSuccessStrategy)(nil)

// NewFirstSuccessStrategy Creates a new firstSuccessStrategy instance as a Strategy
func NewFirstSuccessStrategy(providers []*NamedProvider, timeout time.Duration) Strategy {
	return &firstSuccessStrategy{providers: providers, timeout: timeout}
}

func (f *firstSuccessStrategy) Name() EvaluationStrategy {
	return StrategyFirstSuccess
}

func (f *firstSuccessStrategy) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
	res := evaluateFirstSuccess[bool](ctx, f.providers, flag, of.Boolean, defaultValue, evalCtx, f.timeout)
	return of.BoolResolutionDetail{
		Value:                    res.Value.(bool),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstSuccessStrategy) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
	res := evaluateFirstSuccess[string](ctx, f.providers, flag, of.String, defaultValue, evalCtx, f.timeout)
	return of.StringResolutionDetail{
		Value:                    res.Value.(string),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstSuccessStrategy) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
	res := evaluateFirstSuccess[float64](ctx, f.providers, flag, of.Float, defaultValue, evalCtx, f.timeout)
	return of.FloatResolutionDetail{
		Value:                    res.Value.(float64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstSuccessStrategy) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
	res := evaluateFirstSuccess[int64](ctx, f.providers, flag, of.Int, defaultValue, evalCtx, f.timeout)
	return of.IntResolutionDetail{
		Value:                    res.Value.(int64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

func (f *firstSuccessStrategy) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	return evaluateFirstSuccess[any](ctx, f.providers, flag, of.Object, defaultValue, evalCtx, f.timeout)
}

func evaluateFirstSuccess[R any](ctx context.Context, providers []*NamedProvider, flag string, flagType of.Type, defaultVal R, evalCtx of.FlattenedContext, timeout time.Duration) of.InterfaceResolutionDetail {
	metadata := make(of.FlagMetadata)
	metadata[MetadataStrategyUsed] = StrategyFirstSuccess
	errChan := make(chan ProviderError, len(providers))
	notFoundChan := make(chan any)

	type namedResolution struct {
		of.InterfaceResolutionDetail
		name string
	}
	finishChan := make(chan *namedResolution, len(providers))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for _, provider := range providers {
		go func(c context.Context, p *NamedProvider) {
			resultChan := make(chan *namedResolution)
			go func() {
				r := evaluate[R](ctx, p, flag, flagType, defaultVal, evalCtx)
				resultChan <- &namedResolution{r, p.Name}
			}()

			select {
			case <-c.Done():
				return
			case r := <-resultChan:
				if r.Error() != nil && r.ResolutionDetail().ErrorCode == of.FlagNotFoundCode {
					notFoundChan <- struct{}{}
					return
				} else if r.Error() != nil {
					errChan <- ProviderError{
						Err:          r.ResolutionError,
						ProviderName: p.Name,
					}
					return
				}
				finishChan <- r
			}

		}(ctx, provider)
	}

	errs := make([]ProviderError, 0, len(providers))
	notFoundCount := 0
	for {
		if len(errs) == len(providers) {
			err := NewAggregateError(errs)
			return BuildDefaultResult[R](StrategyFirstSuccess, defaultVal, err)
		}

		select {
		case result := <-finishChan:
			resolution := result.InterfaceResolutionDetail
			metadata[MetadataSuccessfulProviderName] = result.name
			cancel()
			resolution.FlagMetadata = mergeFlagMeta(resolution.FlagMetadata, metadata)
			return resolution
		case err := <-errChan:
			errs = append(errs, err)
		case <-notFoundChan:
			notFoundCount += 1
			if notFoundCount == len(providers) {
				return BuildDefaultResult[R](StrategyFirstSuccess, defaultVal, nil)
			}
		case <-ctx.Done():
			err := ctx.Err()
			if len(errs) > 0 {
				err = NewAggregateError(errs)
			}
			return BuildDefaultResult[R](StrategyFirstSuccess, defaultVal, err)
		}
	}
}
