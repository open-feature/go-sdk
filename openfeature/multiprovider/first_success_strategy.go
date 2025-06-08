package multiprovider

import (
	"context"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
)

func NewFirstSuccessStrategy(providers []*NamedProvider, timeout time.Duration) StrategyFn[FlagTypes] {
	return firstSuccessStrategyFn[FlagTypes](providers, timeout)
}

func firstSuccessStrategyFn[T FlagTypes](providers []*NamedProvider, timeout time.Duration) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, evalCtx of.FlattenedContext) of.GeneralResolutionDetail[T] {
		metadata := make(of.FlagMetadata)
		metadata[MetadataStrategyUsed] = StrategyFirstSuccess
		errChan := make(chan ProviderError, len(providers))
		notFoundChan := make(chan any)

		type namedResolution struct {
			res  of.GeneralResolutionDetail[T]
			name string
		}
		finishChan := make(chan *namedResolution, len(providers))
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		for _, provider := range providers {
			go func(c context.Context, p *NamedProvider) {
				resultChan := make(chan *namedResolution)
				go func() {
					r := evaluate(ctx, p, flag, defaultValue, evalCtx)
					resultChan <- &namedResolution{r, p.Name}
				}()

				select {
				case <-c.Done():
					return
				case r := <-resultChan:
					if r.res.Error() != nil && r.res.ResolutionDetail().ErrorCode == of.FlagNotFoundCode {
						notFoundChan <- struct{}{}
						return
					} else if r.res.Error() != nil {
						errChan <- ProviderError{
							Err:          r.res.ResolutionError,
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
				return BuildDefaultResult(StrategyFirstSuccess, defaultValue, err)
			}

			select {
			case result := <-finishChan:
				resolution := result.res
				metadata[MetadataSuccessfulProviderName] = result.name
				cancel()
				resolution.FlagMetadata = mergeFlagMeta(resolution.FlagMetadata, metadata)
				return resolution
			case err := <-errChan:
				errs = append(errs, err)
			case <-notFoundChan:
				notFoundCount += 1
				if notFoundCount == len(providers) {
					return BuildDefaultResult(StrategyFirstSuccess, defaultValue, nil)
				}
			case <-ctx.Done():
				err := ctx.Err()
				if len(errs) > 0 {
					err = NewAggregateError(errs)
				}
				return BuildDefaultResult(StrategyFirstSuccess, defaultValue, err)
			}
		}
	}
}
