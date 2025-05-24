// Package telemetry provides utilities for extracting data from the OpenFeature SDK for use in telemetry signals.
package telemetry

import (
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
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
	// Body is the flag's value and is only present if the flag's variant is empty.
	Body map[string]any
}

// The OpenTelemetry compliant event attributes for flag evaluation.
const (
	FlagKey          string = "feature_flag.key"
	ErrorTypeKey     string = "error.type"
	ResultValueKey   string = "feature_flag.result.value"
	ResultVariantKey string = "feature_flag.result.variant"
	ErrorMessageKey  string = "error.message"
	ContextIDKey     string = "feature_flag.context.id"
	ProviderNameKey  string = "feature_flag.provider.name"
	ResultReasonKey  string = "feature_flag.result.reason"
	FlagSetIDKey     string = "feature_flag.set.id"
	VersionKey       string = "feature_flag.version"
)

// FlagEvaluationKey is the name of the feature flag evaluation event.
const FlagEvaluationKey string = "feature_flag.evaluation"

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

	body := make(map[string]any)

	if details.Variant != "" {
		attributes[ResultVariantKey] = details.Variant
	} else {
		body[ResultValueKey] = details.Value
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
			Body:       body,
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
		Body:       body,
	}
}
