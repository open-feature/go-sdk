package multiprovider

import (
	"context"
	"errors"

	of "github.com/open-feature/go-sdk/openfeature"
)

func NewFirstSuccessStrategy(providers []*NamedProvider) StrategyFn[FlagTypes] {
	return firstSuccessStrategyFn[FlagTypes](providers)
}

func firstSuccessStrategyFn[T FlagTypes](providers []*NamedProvider) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) GeneralResolutionDetail[T] {
		resolutionErrors := make([]error, 0, len(providers))
		for _, provider := range providers {
			resolution := evaluate(ctx, provider, flag, defaultValue, flatCtx)
			if resolution.Error() != nil {
				resolutionErrors = append(resolutionErrors, resolution.Error())
				continue
			}
			resolution.FlagMetadata = setFlagMetadata(StrategyFirstSuccess, provider.Name, resolution.FlagMetadata)
			return resolution
		}
		return BuildDefaultResult(StrategyFirstSuccess, defaultValue, errors.Join(resolutionErrors...))
	}
}
