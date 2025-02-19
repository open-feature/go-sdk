package openfeature

import "strings"

type EvaluationEvent struct {
	Name string
	Attributes map[string]interface{}
	Data map[string]interface{}
}

const ( 
	FlagEvaluationEventName string = "feature_flag.evaluation"

	// The OpenTelemetry compliant event attributes for flag evaluation.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/

	TelemetryKey string = "feature_flag.key"
	TelemetryErrorCode string = "error.type"
	TelemetryVariant string = "feature_flag.variant"
	TelemetryContextID string = "feature_flag.context.id"
	TelemetryErrorMsg string = "feature_flag.evaluation.error.message"
	TelemetryReason string = "feature_flag.evaluation.reason"
	TelemetryProvider string = "feature_flag.provider_name"
	TelemetryFlagSetID string = "feature_flag.set.id"
	TelemetryVersion string = "feature_flag.version"


	// Well-known flag metadata attributes for telemetry events.
	// Specification: https://openfeature.dev/specification/appendix-d#flag-metadata
	TelemetryFlagMetaContextId string = "contextId"
	TelemetryFlagMetaFlagSetId string = "flagSetId"
	TelemetryFlagMetaVersion string = "version"

	// OpenTelemetry event body.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/
	TelemetryBody string = "value"
)

func CreateEvaluationEvent(hookContext HookContext, evalDetails InterfaceEvaluationDetails) EvaluationEvent {
	attributes := map[string]interface{}{
		TelemetryKey: hookContext.flagKey,
		TelemetryProvider: hookContext.providerMetadata.Name,
	}

	if evalDetails.ResolutionDetail.Reason == "" {
		attributes[TelemetryReason] = strings.ToLower(string(UnknownReason))
	}

	data := map[string]interface{}{}

	if evalDetails.Variant != "" {
		attributes[TelemetryVariant] = evalDetails.Variant
	} else {
		data[TelemetryBody] = evalDetails.Value
	}

	contextID, exists := evalDetails.FlagMetadata[TelemetryFlagMetaContextId]
	if !exists || contextID == "" {
		contextID = hookContext.evaluationContext.targetingKey
	}

	if contextID != "" {
		attributes[TelemetryContextID] = contextID
	}

	setID, exists := evalDetails.FlagMetadata[TelemetryFlagMetaFlagSetId]
	if exists && setID != "" {
		attributes[TelemetryFlagSetID] = setID
	}

	return EvaluationEvent{
		Name: FlagEvaluationEventName,
		Attributes: attributes,
		Data: data,
	}
}
