// Package telemetry provides utilities for extracting data from the OpenFeature SDK for use in telemetry signals.
package telemetry

import (
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// EventName is the name of the feature flag evaluation event.
const EventName string = "feature_flag.evaluation"

const (
	flagMetaContextIDKey string = "contextId"
	flagMetaFlagSetIDKey string = "flagSetId"
	flagMetaVersionKey   string = "version"
)

// EventAttributes returns a slice of OpenTelemetry attributes that can be used to create an event for a feature flag evaluation.
// It is intended to be used in the `Finally` stage of a [openfeature.Hook].
func EventAttributes(hookContext openfeature.HookContext, details openfeature.InterfaceEvaluationDetails) []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		semconv.FeatureFlagKey(hookContext.FlagKey()),
		semconv.FeatureFlagProviderName(hookContext.ProviderMetadata().Name),
	}

	reason := strings.ToLower(string(openfeature.UnknownReason))
	if details.Reason != "" {
		reason = strings.ToLower(string(details.Reason))
	}
	attributes = append(attributes, semconv.FeatureFlagResultReasonKey.String(reason))

	if details.Variant != "" {
		attributes = append(attributes, semconv.FeatureFlagResultVariant(details.Variant))
	}

	contextID := hookContext.EvaluationContext().TargetingKey()
	if flagMetaContextID, ok := details.FlagMetadata[flagMetaContextIDKey].(string); ok {
		contextID = flagMetaContextID
	}
	attributes = append(attributes, semconv.FeatureFlagContextID(contextID))

	if setID, ok := details.FlagMetadata[flagMetaFlagSetIDKey].(string); ok {
		attributes = append(attributes, semconv.FeatureFlagSetID(setID))
	}

	if version, ok := details.FlagMetadata[flagMetaVersionKey].(string); ok {
		attributes = append(attributes, semconv.FeatureFlagVersion(version))
	}

	if details.Reason != openfeature.ErrorReason {
		return attributes
	}

	errorType := strings.ToLower(string(openfeature.GeneralCode))
	if details.ErrorCode != "" {
		errorType = strings.ToLower(string(details.ErrorCode))
	}
	attributes = append(attributes, semconv.ErrorTypeKey.String(errorType))

	if details.ErrorMessage != "" {
		attributes = append(attributes, semconv.ErrorMessage(details.ErrorMessage))
	}

	return attributes
}
