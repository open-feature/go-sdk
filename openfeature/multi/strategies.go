package multi

import (
	"context"
	"maps"
	"regexp"
	"strings"

	of "github.com/open-feature/go-sdk/openfeature"
)

// EvaluationStrategy options
const (
	// StrategyFirstMatch returns the result of the first [of.FeatureProvider] whose response is not [of.FlagNotFoundCode].
	// This is executed sequentially, and not in parallel. Any returned errors from a provider other than flag not found
	// will result in a default response with a set error.
	StrategyFirstMatch EvaluationStrategy = "strategy-first-match"
	// StrategyFirstSuccess returns the result of the first [of.FeatureProvider] whose response that is not an error.
	// This is very similar to [StrategyFirstMatch], but does not raise errors. This is executed sequentially.
	StrategyFirstSuccess EvaluationStrategy = "strategy-first-success"
	// StrategyComparison returns a response of all [of.FeatureProvider] instances in agreement. All providers are
	// called in parallel and then the results of each non-error result are compared to each other. If all responses
	// agree, then that value will be returned. Otherwise, the value from the designated fallback [of.FeatureProvider]
	// instance's response will be returned. The fallback provider will be assigned to the first provider registered if
	// the [WithFallbackProvider] Option is not explicitly set.
	StrategyComparison EvaluationStrategy = "strategy-comparison"
	// StrategyCustom allows for using a custom [StrategyFn] implementation. If this is set you MUST use the WithCustomStrategy
	// Option to set it
	StrategyCustom EvaluationStrategy = "strategy-custom"
)

// Additional [of.Reason] options
const (
	// ReasonAggregated - the resolved value was the agreement of all providers in the multi.Provider using the
	// [StrategyComparison] strategy
	ReasonAggregated of.Reason = "AGGREGATED"
	// ReasonAggregatedFallback ReasonAggregated - the resolved value was result of the fallback provider because the
	// providers in multi.Provider were not in agreement using the [StrategyComparison] strategy.
	ReasonAggregatedFallback of.Reason = "AGGREGATED_FALLBACK"
)

// errAggregationNotAllowedText sentinel returned if [of.FeatureProvider.ObjectEvaluation] is called without a set custom
// strategy when response objects are not comparable.
const errAggregationNotAllowedText = "object evaluation not allowed for non-comparable types without custom comparable func"

type (
	// EvaluationStrategy Defines a strategy to use for resolving the result from multiple providers.
	EvaluationStrategy = string

	// FlagTypes defines the types that can be used for flag values.
	FlagTypes interface {
		int64 | float64 | string | bool | any
	}
	// StrategyFn defines the signature for a strategy function.
	StrategyFn[T FlagTypes] func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T]
	// StrategyConstructor defines the signature for the function that will be called to retrieve the closure that acts
	// as the custom strategy implementation. This function should return a [StrategyFn]
	StrategyConstructor func(providers []NamedProvider) StrategyFn[FlagTypes]
)

// Common Components

// setFlagMetadata sets common metadata for evaluations.
func setFlagMetadata(strategyUsed EvaluationStrategy, successProviderName string, metadata of.FlagMetadata) of.FlagMetadata {
	if metadata == nil {
		metadata = make(of.FlagMetadata)
	}
	metadata[MetadataSuccessfulProviderName] = successProviderName
	metadata[MetadataStrategyUsed] = strategyUsed
	return metadata
}

// cleanErrorMessage removes prefixes from error messages.
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

// mergeFlagMeta merges flag metadata together into a single [of.FlagMetadata] instance by performing a shallow merge.
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

// BuildDefaultResult should be called when a [StrategyFn] is in a failure state and needs to return a default value.
// This method will build a resolution detail with the internal provided error set. This method is exported for those
// writing their own custom [StrategyFn].
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

// Evaluate is a generic method used to resolve a flag from a single [NamedProvider] without losing type information.
// This method is exported for those writing their own custom [StrategyFn]. Since any is an allowed [FlagTypes] this can
// be set to any type, but this should be done with care outside the specified primitive [FlagTypes]
func Evaluate[T FlagTypes](ctx context.Context, provider NamedProvider, flag string, defaultVal T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
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
		resolution.Value = res.Value.(T)
	}

	if resolution.FlagMetadata == nil {
		resolution.FlagMetadata = make(of.FlagMetadata, 2)
	}

	resolution.FlagMetadata[MetadataProviderName] = provider.Name()
	resolution.FlagMetadata[MetadataProviderType] = provider.Metadata().Name

	return resolution
}
