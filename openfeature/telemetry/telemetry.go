package telemetry

import (
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
)

type EvaluationEvent struct {
	Name       string
	Attributes map[string]any
	Body       map[string]any
}

const (
	// The OpenTelemetry compliant event attributes for flag evaluation.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/

	TelemetryKey       string = "feature_flag.key"
	TelemetryErrorCode string = "error.type"
	TelemetryVariant   string = "feature_flag.variant"
	TelemetryContextID string = "feature_flag.context.id"
	TelemetryErrorMsg  string = "feature_flag.evaluation.error.message"
	TelemetryReason    string = "feature_flag.evaluation.reason"
	TelemetryProvider  string = "feature_flag.provider_name"
	TelemetryFlagSetID string = "feature_flag.set.id"
	TelemetryVersion   string = "feature_flag.version"

	// Well-known flag metadata attributes for telemetry events.
	// Specification: https://openfeature.dev/specification/appendix-d#flag-metadata
	TelemetryFlagMetaContextId string = "contextId"
	TelemetryFlagMetaFlagSetId string = "flagSetId"
	TelemetryFlagMetaVersion   string = "version"

	// OpenTelemetry event body.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/
	TelemetryBody string = "value"

	FlagEvaluationEventName string = "feature_flag.evaluation"
)

func CreateEvaluationEvent(hookContext openfeature.HookContext, details openfeature.InterfaceEvaluationDetails) EvaluationEvent {
	attributes := map[string]any{
		TelemetryKey:      hookContext.FlagKey(),
		TelemetryProvider: hookContext.ProviderMetadata().Name,
	}

	if details.Reason != "" {
		attributes[TelemetryReason] = strings.ToLower(string(details.Reason))
	} else {
		attributes[TelemetryReason] = strings.ToLower(string(openfeature.UnknownReason))
	}

	body := map[string]any{}

	if details.Variant != "" {
		attributes[TelemetryVariant] = details.Variant
	} else {
		body[TelemetryBody] = details.Value
	}

	contextID, exists := details.FlagMetadata[TelemetryFlagMetaContextId]
	if !exists {
		contextID = hookContext.EvaluationContext().TargetingKey()
	}

	attributes[TelemetryContextID] = contextID

	setID, exists := details.FlagMetadata[TelemetryFlagMetaFlagSetId]
	if exists {
		attributes[TelemetryFlagSetID] = setID
	}

	version, exists := details.FlagMetadata[TelemetryFlagMetaVersion]
	if exists {
		attributes[TelemetryVersion] = version
	}

	if details.Reason == openfeature.ErrorReason {
		if details.ErrorCode != "" {
			attributes[TelemetryErrorCode] = details.ErrorCode
		} else {
			attributes[TelemetryErrorCode] = openfeature.GeneralCode
		}

		if details.ErrorMessage != "" {
			attributes[TelemetryErrorMsg] = details.ErrorMessage
		}
	}

	return EvaluationEvent{
		Name:       FlagEvaluationEventName,
		Attributes: attributes,
		Body:       body,
	}
}
