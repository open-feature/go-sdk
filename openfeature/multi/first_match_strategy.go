package multi

import (
	"context"

	of "github.com/open-feature/go-sdk/openfeature"
)

// newFirstMatchStrategy returns a [StrategyFn] that returns the result of the first [of.FeatureProvider] whose response is
// not [of.FlagNotFoundCode]. The definition of "first" depends on the configured run-mode. With sequential execution, it's the first provider in order. With parallel, it's the first to return a result.
func newFirstMatchStrategy(providers []NamedProvider, runMode runModeFn[FlagTypes]) StrategyFn[FlagTypes] {
	return firstMatchStrategyFn(providers, runMode)
}

func firstMatchStrategyFn[T FlagTypes](providers []NamedProvider, runMode runModeFn[T]) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
		for providerName, resolution := range runMode(ctx, providers, flag, defaultValue, flatCtx) {
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
