// Package telemetry provides utilities for extracting data from the OpenFeature SDK for use in telemetry signals.
package telemetry

import (
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// EvaluationEvent represents an event that is emitted when a flag is evaluated.
// It is intended to be used to record flag evaluation events as OpenTelemetry log records.
// See the OpenFeature specification [Appendix D: Observability] and
// the OpenTelemetry [Semantic conventions for feature flags in logs] for more information.
//
// [Appendix D: Observability]: https://openfeature.dev/specification/appendix-d
// [Semantic conventions for feature flags in logs]: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/
type EvaluationEvent struct {
	// Name is the name of the event.
	// It is always "feature_flag.evaluation".
	Name string
	// Attributes represents the event's attributes.
	Attributes map[string]any
}


// Use OTel semconv constants for feature flag attributes.
const (
    FlagKey          = semconv.FeatureFlagKeyKey
    ErrorTypeKey     = semconv.ErrorTypeKey
    ResultValueKey   = semconv.FeatureFlagResultValueKey
    ResultVariantKey = semconv.FeatureFlagResultVariantKey
    ErrorMessageKey  = semconv.ErrorMessageKey
    ContextIDKey     = semconv.FeatureFlagContextIDKey
    ProviderNameKey  = semconv.FeatureFlagProviderNameKey
    ResultReasonKey  = semconv.FeatureFlagResultReasonKey
    FlagSetIDKey     = semconv.FeatureFlagSetIDKey
    VersionKey       = semconv.FeatureFlagVersionKey
)

// FlagEvaluationKey is the name of the feature flag evaluation event.
const FlagEvaluationKey = semconv.FeatureFlagEvaluationEvent

const (
	flagMetaContextIDKey string = "contextId"
	flagMetaFlagSetIDKey string = "flagSetId"
	flagMetaVersionKey   string = "version"
)

// CreateEvaluationEvent creates an [EvaluationEvent].
// It is intended to be used in the `Finally` stage of a [openfeature.Hook].
func CreateEvaluationEvent(hookContext openfeature.HookContext, details openfeature.InterfaceEvaluationDetails) EvaluationEvent {
	attributes := map[string]any{
		FlagKey:         hookContext.FlagKey(),
		ProviderNameKey: hookContext.ProviderMetadata().Name,
	}

	attributes[ResultReasonKey] = strings.ToLower(string(openfeature.UnknownReason))
	if details.Reason != "" {
		attributes[ResultReasonKey] = strings.ToLower(string(details.Reason))
	}

	if details.Variant != "" {
		attributes[ResultVariantKey] = details.Variant
	} else {
		attributes[ResultValueKey] = details.Value
	}

	attributes[ContextIDKey] = hookContext.EvaluationContext().TargetingKey()
	if contextID, ok := details.FlagMetadata[flagMetaContextIDKey]; ok {
		attributes[ContextIDKey] = contextID
	}

	if setID, ok := details.FlagMetadata[flagMetaFlagSetIDKey]; ok {
		attributes[FlagSetIDKey] = setID
	}

	if version, ok := details.FlagMetadata[flagMetaVersionKey]; ok {
		attributes[VersionKey] = version
	}

	if details.Reason != openfeature.ErrorReason {
		return EvaluationEvent{
			Name:       FlagEvaluationKey,
			Attributes: attributes,
		}
	}

	attributes[ErrorTypeKey] = strings.ToLower(string(openfeature.GeneralCode))
	if details.ErrorCode != "" {
		attributes[ErrorTypeKey] = strings.ToLower(string(details.ErrorCode))
	}

	if details.ErrorMessage != "" {
		attributes[ErrorMessageKey] = details.ErrorMessage
	}

	return EvaluationEvent{
		Name:       FlagEvaluationKey,
		Attributes: attributes,
	}
}
