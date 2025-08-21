package multiprovider

import (
	"context"
	"maps"
	"regexp"
	"strings"

	of "github.com/open-feature/go-sdk/openfeature"
)

const (
	// Strategies

	// StrategyFirstMatch First provider whose response that is not FlagNotFound will be returned. This is executed
	// sequentially, and not in parallel.
	StrategyFirstMatch EvaluationStrategy = "strategy-first-match"
	// StrategyFirstSuccess First provider response that is not an error will be returned. This is executed in parallel
	StrategyFirstSuccess EvaluationStrategy = "strategy-first-success"
	// StrategyComparison All providers are called in parallel. If all responses agree the value will be returned.
	// Otherwise, the value from the designated fallback provider's response will be returned. The fallback provider
	// will be assigned to the first provider registered.
	StrategyComparison EvaluationStrategy = "strategy-comparison"
	// StrategyCustom allows for using a custom Strategy implementation. If this is set you MUST use the WithCustomStrategy
	// option to set it
	StrategyCustom EvaluationStrategy = "strategy-custom"

	ReasonAggregated         of.Reason = "AGGREGATED"
	ReasonAggregatedFallback of.Reason = "AGGREGATED_FALLBACK"

	ErrAggregationNotAllowedText = "object evaluation not allowed for non-comparable types without custom comparable func"
)

type (
	// EvaluationStrategy Defines a strategy to use for resolving the result from multiple providers.
	EvaluationStrategy = string

	// FlagTypes defines the types that can be used for flag values.
	FlagTypes interface {
		int64 | float64 | string | bool | any
	}
	// StrategyFn is a function type that defines the signature for a strategy function.
	StrategyFn[T FlagTypes] func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T]
)

// Common Components

// setFlagMetadata Sets common metadata for evaluations.
func setFlagMetadata(strategyUsed EvaluationStrategy, successProviderName string, metadata of.FlagMetadata) of.FlagMetadata {
	if metadata == nil {
		metadata = make(of.FlagMetadata)
	}
	metadata[MetadataSuccessfulProviderName] = successProviderName
	metadata[MetadataStrategyUsed] = strategyUsed
	return metadata
}

// cleanErrorMessage Removes prefixes from error messages.
func cleanErrorMessage(msg string) string {
	codeRegex := strings.Join([]string{
		string(of.ProviderNotReadyCode),
		string(of.ProviderFatalCode),
		string(of.FlagNotFoundCode),
		string(of.ParseErrorCode),
		string(of.TypeMismatchCode),
		string(of.TargetingKeyMissingCode),
		string(of.GeneralCode),
	}, "|")
	re := regexp.MustCompile("(?:" + codeRegex + "): (.*)")
	matches := re.FindSubmatch([]byte(msg))
	matchCount := len(matches)
	switch matchCount {
	case 0, 1:
		return msg
	default:
		return strings.TrimSpace(string(matches[1]))
	}
}

// mergeFlagMeta Merges flag metadata together into a single [of.FlagMetadata] instance by performing a shallow merge.
func mergeFlagMeta(tags ...of.FlagMetadata) of.FlagMetadata {
	size := len(tags)
	switch size {
	case 0:
		return make(of.FlagMetadata)
	case 1:
		return tags[0]
	default:
		merged := make(of.FlagMetadata)
		for _, t := range tags {
			maps.Copy(merged, t)
		}
		return merged
	}
}

// BuildDefaultResult The method should be called when a strategy is in a failure state and needs to return a default
// value. This method will build a resolution detail with the internal provided error set.
func BuildDefaultResult[R FlagTypes](strategy EvaluationStrategy, defaultValue R, err error) of.GenericResolutionDetail[R] {
	var rErr of.ResolutionError
	var reason of.Reason
	if err != nil {
		rErr = of.NewGeneralResolutionError(cleanErrorMessage(err.Error()))
		reason = of.ErrorReason
	} else {
		rErr = of.NewFlagNotFoundResolutionError("not found in any provider")
		reason = of.DefaultReason
	}

	return of.GenericResolutionDetail[R]{
		Value: defaultValue,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			ResolutionError: rErr,
			Reason:          reason,
			FlagMetadata:    of.FlagMetadata{MetadataSuccessfulProviderName: "none", MetadataStrategyUsed: strategy},
		},
	}
}

// evaluate Generic method used to resolve a flag from a single provider without losing type information.
func evaluate[T FlagTypes](ctx context.Context, provider *NamedProvider, flag string, defaultVal T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
	var resolution of.GenericResolutionDetail[T]
	switch v := any(defaultVal).(type) {
	case bool:
		res := provider.BooleanEvaluation(ctx, flag, v, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = any(res.Value).(T)
	case string:
		res := provider.StringEvaluation(ctx, flag, v, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = any(res.Value).(T)
	case float64:
		res := provider.FloatEvaluation(ctx, flag, v, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = any(res.Value).(T)
	case int64:
		res := provider.IntEvaluation(ctx, flag, v, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = any(res.Value).(T)
	default:
		res := provider.ObjectEvaluation(ctx, flag, defaultVal, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = any(res.Value).(T)
	}

	if resolution.FlagMetadata == nil {
		resolution.FlagMetadata = make(of.FlagMetadata, 2)
	}

	resolution.FlagMetadata[MetadataProviderName] = provider.Name
	resolution.FlagMetadata[MetadataProviderType] = provider.Metadata().Name

	return resolution
}
