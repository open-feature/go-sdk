package multi

import (
	"errors"

	of "github.com/open-feature/go-sdk/openfeature"
)

// newFirstSuccessStrategy returns a [StrategyFn] that returns the result of the First [of.FeatureProvider] whose response
// is not an error. The definition of "first" depends on the configured run-mode. With sequential execution, it's the first provider in order. With parallel, it's the first to return a result.
func newFirstSuccessStrategy() StrategyFn[FlagTypes] {
	return firstSuccessStrategyFn[FlagTypes]()
}

func firstSuccessStrategyFn[T FlagTypes]() StrategyFn[T] {
	return func(resolutions ResolutionIterator[T], defaultValue T, _ FallbackEvaluator[T]) *of.GenericResolutionDetail[T] {
		resolutionErrors := make([]error, 0)
		for name, resolution := range resolutions {
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
