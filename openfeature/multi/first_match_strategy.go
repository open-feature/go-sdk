package multi

import (
	"context"

	of "github.com/open-feature/go-sdk/openfeature"
)

// newFirstMatchStrategy returns a [StrategyFn] that returns the result of the first [of.FeatureProvider] whose response is
// not [of.FlagNotFoundCode]. This is executed sequentially, and not in parallel.
func newFirstMatchStrategy(providers []*NamedProvider) StrategyFn[FlagTypes] {
	return firstMatchStrategyFn[FlagTypes](providers)
}

func firstMatchStrategyFn[T FlagTypes](providers []*NamedProvider) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
		for _, provider := range providers {
			resolution := Evaluate(ctx, provider, flag, defaultValue, flatCtx)
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

		return BuildDefaultResult(StrategyFirstMatch, defaultValue, nil)
	}
}
