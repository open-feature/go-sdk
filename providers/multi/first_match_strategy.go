package multi

import (
	of "go.openfeature.dev/openfeature/v2"
)

// newFirstMatchStrategy returns a [StrategyFn] that returns the result of the first [of.FeatureProvider] whose response is
// not [of.FlagNotFoundCode]. The definition of "first" depends on the configured run-mode. With sequential execution, it's the first provider in order. With parallel, it's the first to return a result.
func newFirstMatchStrategy() StrategyFn[of.FlagTypes] {
	return firstMatchStrategyFn[of.FlagTypes]()
}

func firstMatchStrategyFn[T of.FlagTypes]() StrategyFn[T] {
	return func(resolutions ResolutionIterator[T], defaultValue T, _ FallbackEvaluator[T]) *of.GenericResolutionDetail[T] {
		for providerName, resolution := range resolutions {
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
			resolution.FlagMetadata = setFlagMetadata(StrategyFirstMatch, providerName, resolution.FlagMetadata)
			return resolution
		}

		return BuildDefaultResult(StrategyFirstMatch, defaultValue, nil)
	}
}
