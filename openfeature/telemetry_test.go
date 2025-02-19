package openfeature

import (
	"strings"
	"testing"
)

func TestCreateEvaluationEvent_1_3_1(t *testing.T) {
	flagKey := "test-flag"
	mockProviderMetadata := Metadata{
		Name: "test-provider",
	}
	mockClientMetadata := ClientMetadata{
		domain: "test-client",
	}
	mockHookContext := HookContext{
		flagKey:          flagKey,
		flagType:         Boolean,
		defaultValue:     true,
		clientMetadata:   mockClientMetadata,
		providerMetadata: mockProviderMetadata,
	}

	mockDetails := InterfaceEvaluationDetails{
		Value: true,
		EvaluationDetails: EvaluationDetails{
			FlagKey:  flagKey,
			FlagType: Boolean,
			ResolutionDetail: ResolutionDetail{
				Reason:       StaticReason,
				FlagMetadata: FlagMetadata{},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Name != "feature_flag.evaluation" {
		t.Errorf("Expected event name to be 'feature_flag.evaluation', got '%s'", event.Name)
	}

	if event.Attributes[TelemetryKey] != flagKey {
		t.Errorf("Expected event attribute 'KEY' to be '%s', got '%s'", flagKey, event.Attributes[TelemetryKey])
	}

	if event.Attributes[TelemetryReason] != strings.ToLower(string(StaticReason)) {
		t.Errorf("Expected evaluation reason to be '%s', got '%s'", strings.ToLower(string(StaticReason)), event.Attributes[TelemetryReason])
	}

	if event.Attributes[TelemetryProvider] != "test-provider" {
		t.Errorf("Expected provider name to be 'test-provider', got '%s'", event.Attributes[TelemetryProvider])
	}

	if event.Body[TelemetryBody] != true {
		t.Errorf("Expected event body 'VALUE' to be 'true', got '%v'", event.Body[TelemetryBody])
	}
}
func TestCreateEvaluationEvent_WithVariant(t *testing.T) {

}
func TestCreateEvaluationEvent_WithFlagMetaData(t *testing.T) {

}
func TestCreateEvaluationEvent_WithErrors(t *testing.T) {

}
func TestCreateEvaluationEvent_withUnknownReason(t *testing.T) {

}
