package openfeature

import "strings"

// // TelemetryAttribute defines an OpenTelemetry compliant event attributes for flag evaluation.
// // Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/
// type TelemetryAttribute struct {
// 	Key          string // The lookup key of the feature flag.
// 	ErrorCode    string // Describes an error that the operation ended with.
// 	Variant      string // A semantic identifier for an evaluated flag.
// 	ContextID    string // The unique identifier for the the flag evaluation context. For example, the targeting key.
// 	ErrorMessage string // A message explaining the nature of an  error occurring during flag evaluation.
// 	Reason       string // The reason code which shows how a feature flag value was determined.
// 	Provider     string // Describes an error the operation ended with.
// 	FlagSetID    string // The identifier of the flag set to which the feature flag belongs.
// 	Version      string // The version of the ruleset used during the evaluation. This may be any stable value which uniquely identifies the ruleset.
// }

//	TelemetryEvaluationData Event data, sometimes referred as "body", is specific to a specific event. In this case, the event is `feature_flag.evaluation`. That's why the prefix is omitted from the values.
// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/
// type TelemetryEvaluationData struct {
// 	value string // The evaluated value of the feature flag.
// }

// Well-known flag metadata attributes for telemetry events.
// Specification: https://openfeature.dev/specification/appendix-d#flag-metadata
// type TelemetryFlagMetadata struct {
// 	ContextId  string
// 	FlagSetId  string
// 	Version    string
// }

type EvaluationEvent struct {
	Name string
	Attributes map[string]interface{}
	Data map[string]interface{}
}

const ( 
	FlagEvaluationEventName string = "feature_flag.evaluation"

	// The OpenTelemetry compliant event attributes for flag evaluation.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/

	// The lookup key of the feature flag. 
	// requirement level:`required`
	// example: `logo-color`
	TelemetryKey string = "feature_flag.key"

	// Describes a category/type of error the operation ended with.
	// requirement level: `conditionally
	// required condition: `reason` is `error`
	// example: `flag_not_found`
	TelemetryErrorCode string = "error.type"

	// A semantic identifier for an evaluated flag value.
	// requirement level: `conditionally required`
	// condition: variant is defined on the evaluation details
	// example: `blue`; `on`; `true`
	TelemetryVariant string = "feature_flag.variant"

	// The unique identifier for the flag evaluation context. For example, the targeting key.
	// requirement level: `recommended`
	// example: `5157782b-2203-4c80-a857-dbbd5e7761db`
	TelemetryContextID string = "feature_flag.context.id"

	// A message explaining the nature of an error occurring during flag evaluation.
	// requirement level: `recommended`
	// example: `Flag not found`
	TelemetryErrorMsg string = "feature_flag.evaluation.error.message"

	// The reason code which shows how a feature flag value was determined.
	// requirement level: `recommended`
	// example: `targeting_match`
	TelemetryReason string = "feature_flag.evaluation.reason"

	// Describes a category/type error the operation ended with.
	// requirement level: `recommended`
	// example: `flag_not_found`
	TelemetryProvider string = "feature_flag.provider_name"

	// The identifier of the flag set to which the feature flag belongs.
	// requirement level: `recommended`
	// example: `proj-1`; `ab98sgs`; `service1/dev`
	TelemetryFlagSetID string = "feature_flag.set.id"

	// The version of the ruleset used during the evaluation. 
	// This may be any stable value which uniquely identifies the ruleset.
	// requirement level: `recommended`
	// example: `1.0.0`; `2021-01-01`
	TelemetryVersion string = "feature_flag.version"


	// Well-known flag metadata attributes for telemetry events.
	// Specification: https://openfeature.dev/specification/appendix-d#flag-metadata


	// The context identifier returned in the flag metadata uniquely identifies
    // the subject of the flag evaluation. If not available, the targeting key
    // should be used.
	TelemetryFlagMetaContextId string = "contextId"

	// A logical identifier for the flag set.
	TelemetryFlagMetaFlagSetId string = "flagSetId"

	// A version string (format unspecified) for the flag or flag set.
	TelemetryFlagMetaVersion string = "version"

	//	TelemetryEvaluationData Event data, sometimes referred as "body", is specific 
	// to a specific event. In this case, the event is `feature_flag.evaluation`. 
	// That's why the prefix is omitted from the values.
	// Specification: https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-logs/

	// The evaluated value of the feature flag.
	// requirement level: `conditionally required`
    // condition: variant is not defined on the evaluation details
    // example: `#ff0000`; `1`; `true`
	TelemetryEvalData string = "value"
)

func CreateEvaluationEvent(hookContext HookContext, evaluationDetails InterfaceEvaluationDetails) EvaluationEvent {
	attributes := map[string]interface{}{
		TelemetryKey: hookContext.flagKey,
		TelemetryProvider: hookContext.providerMetadata.Name,
	}

	if evaluationDetails.ResolutionDetail.Reason == "" {
		attributes[TelemetryReason] = strings.ToLower(string(UnknownReason))
	}

	data := map[string]interface{}{}

	if evaluationDetails.Variant != "" {
		attributes[TelemetryVariant] = evaluationDetails.Variant
	} else {
		data[TelemetryEvalData] = evaluationDetails.Value
	}

	contextID, exists := evaluationDetails.FlagMetadata[TelemetryFlagMetaContextId]
	if !exists || contextID == "" {
		contextID = hookContext.evaluationContext.targetingKey
	}

	if contextID != "" {
		attributes[TelemetryContextID] = contextID
	}

	setID, exists := evaluationDetails.FlagMetadata[TelemetryFlagMetaFlagSetId]
	if exists && setID != "" {
		attributes[TelemetryFlagSetID] = setID
	}

	return EvaluationEvent{
		Name: FlagEvaluationEventName,
		Attributes: attributes,
		Data: data,
	}
}
