package multi

import (
	"context"
	"errors"

	of "github.com/open-feature/go-sdk/openfeature"
)

// newFirstSuccessStrategy returns a [StrategyFn] that returns the result of the First [of.FeatureProvider] whose response
// is not an error.
func newFirstSuccessStrategy(providers []NamedProvider, runMode runModeFn[FlagTypes]) StrategyFn[FlagTypes] {
	return firstSuccessStrategyFn(providers, runMode)
}

func firstSuccessStrategyFn[T FlagTypes](providers []NamedProvider, runMode runModeFn[T]) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
		resolutionErrors := make([]error, 0, len(providers))
		for name, resolution := range runMode(ctx, providers, flag, defaultValue, flatCtx) {
			if resolution.Error() != nil {
				resolutionErrors = append(resolutionErrors, resolution.Error())
				continue
			}
			resolution.FlagMetadata = setFlagMetadata(StrategyFirstSuccess, name, resolution.FlagMetadata)
			return resolution
		}
		return BuildDefaultResult(StrategyFirstSuccess, defaultValue, errors.Join(resolutionErrors...))
	}
}
