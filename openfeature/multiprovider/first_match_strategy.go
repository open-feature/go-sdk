package multiprovider

import (
	"context"

	of "github.com/open-feature/go-sdk/openfeature"
)

func NewFirstMatchStrategy(providers []*NamedProvider) StrategyFn[FlagTypes] {
	return firstMatchStrategyFn[FlagTypes](providers)
}

func firstMatchStrategyFn[T FlagTypes](providers []*NamedProvider) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) GeneralResolutionDetail[T] {
		for _, provider := range providers {
			resolution := evaluate(ctx, provider, flag, defaultValue, flatCtx)
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
