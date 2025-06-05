package multiprovider

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
	"regexp"
	"strings"
)

const (
	// Strategies

	// StrategyFirstMatch First provider whose response that is not FlagNotFound will be returned. This is executed
	// sequentially, and not in parallel.
	StrategyFirstMatch = "strategy-first-match"
	// StrategyFirstSuccess First provider response that is not an error will be returned. This is executed in parallel
	StrategyFirstSuccess = "strategy-first-success"
	// StrategyComparison All providers are called in parallel. If all responses agree the value will be returned.
	// Otherwise, the value from the designated fallback provider's response will be returned. The fallback provider
	// will be assigned to the first provider registered.
	StrategyComparison = "strategy-comparison"

	ReasonAggregated         of.Reason = "AGGREGATED"
	ReasonAggregatedFallback of.Reason = "AGGREGATED_FALLBACK"

	ErrAggregationNotAllowedText = "object evaluation not allowed for non-comparable types without custom comparable func"
)

type (
	// EvaluationStrategy Defines a strategy to use for resolving the result from multiple providers
	EvaluationStrategy = string
	// Strategy Interface for evaluating providers within the multi-provider.
	Strategy interface {
		Name() EvaluationStrategy
		BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail
		StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail
		FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail
		IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail
		ObjectEvaluation(ctx context.Context, flag string, defaultValue any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail
	}
)

// Common Components

// setFlagMetadata sets common metadata for evaluations
func setFlagMetadata(strategyUsed EvaluationStrategy, successProviderName string, metadata of.FlagMetadata) of.FlagMetadata {
	if metadata == nil {
		metadata = make(of.FlagMetadata)
	}
	metadata[MetadataSuccessfulProviderName] = successProviderName
	metadata[MetadataStrategyUsed] = strategyUsed
	return metadata
}

// cleanErrorMessage Remove prefixes from error messages
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

// mergeFlagMeta Merges flag metadata together into a single FlagMetadata instance by performing a shallow merge
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
			for key, value := range t {
				merged[key] = value
			}
		}
		return merged
	}
}

// BuildDefaultResult when executing a strategy that results in a failure state this can be called to return a
// default result with an internal error
func BuildDefaultResult[R any](strategy EvaluationStrategy, defaultValue R, err error) of.InterfaceResolutionDetail {
	var rErr of.ResolutionError
	var reason of.Reason
	if err != nil {
		rErr = of.NewGeneralResolutionError(cleanErrorMessage(err.Error()))
		reason = of.ErrorReason
	} else {
		rErr = of.NewFlagNotFoundResolutionError("not found in any provider")
		reason = of.DefaultReason
	}

	return of.InterfaceResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			ResolutionError: rErr,
			Reason:          reason,
			FlagMetadata:    of.FlagMetadata{MetadataSuccessfulProviderName: "none", MetadataStrategyUsed: strategy},
		},
	}
}

// evaluate Generic method to resolve a flag from a single provider without losing type information
func evaluate[R any](ctx context.Context, provider *NamedProvider, flag string, flagType of.Type, defaultVal R, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	var resolution of.InterfaceResolutionDetail
	switch flagType {
	case of.Object:
		resolution = provider.ObjectEvaluation(ctx, flag, defaultVal, evalCtx)
	case of.Boolean:
		res := provider.BooleanEvaluation(ctx, flag, any(defaultVal).(bool), evalCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.String:
		res := provider.StringEvaluation(ctx, flag, any(defaultVal).(string), evalCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.Float:
		res := provider.FloatEvaluation(ctx, flag, any(defaultVal).(float64), evalCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.Int:
		res := provider.IntEvaluation(ctx, flag, any(defaultVal).(int64), evalCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	}

	resolution.FlagMetadata[MetadataProviderName] = provider.Name
	resolution.FlagMetadata[MetadataProviderType] = provider.Metadata().Name

	return resolution
}
