package telemetry

import (
	"strings"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
)

func TestCreateEvaluationEvent_InvalidFlagMetadataTypes(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"integer", 3},
		{"boolean", true},
		{"slice", []string{"value"}},
		{"map", map[string]string{"key": "value"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			hookContext := openfeature.NewHookContext(
				"flag", openfeature.Boolean, false,
				openfeature.NewClientMetadata("client"), openfeature.Metadata{Name: "provider"},
				openfeature.NewEvaluationContext("target", nil),
			)
			details := openfeature.InterfaceEvaluationDetails{
				EvaluationDetails: openfeature.EvaluationDetails{
					ResolutionDetail: openfeature.ResolutionDetail{
						FlagMetadata: openfeature.FlagMetadata{
							flagMetaContextIDKey: test.value,
							flagMetaFlagSetIDKey: test.value,
							flagMetaVersionKey:   test.value,
						},
					},
				},
			}
			event := CreateEvaluationEvent(hookContext, details)
			if got := event.Attributes[ContextIDKey]; got != "target" {
				t.Errorf("invalid contextId should fall back to targeting key, got %#v", got)
			}
			for _, key := range []string{FlagSetIDKey, VersionKey} {
				if got, ok := event.Attributes[key]; ok {
					t.Errorf("invalid metadata should omit %s, got %#v", key, got)
				}
			}
		})
	}
}

func TestCreateEvaluationEvent_1_3_1_BasicEvent(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, true, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: true,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey:  flagKey,
			FlagType: openfeature.Boolean,
			ResolutionDetail: openfeature.ResolutionDetail{
				Reason:       openfeature.StaticReason,
				FlagMetadata: openfeature.FlagMetadata{},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Name != "feature_flag.evaluation" {
		t.Errorf("Expected event name to be 'feature_flag.evaluation', got '%s'", event.Name)
	}

	if event.Attributes[FlagKey] != flagKey {
		t.Errorf("Expected event attribute 'KEY' to be '%s', got '%s'", flagKey, event.Attributes[FlagKey])
	}

	if event.Attributes[ResultReasonKey] != strings.ToLower(string(openfeature.StaticReason)) {
		t.Errorf("Expected evaluation reason to be '%s', got '%s'", strings.ToLower(string(openfeature.StaticReason)), event.Attributes[ResultReasonKey])
	}

	if event.Attributes[ProviderNameKey] != "test-provider" {
		t.Errorf("Expected provider name to be 'test-provider', got '%s'", event.Attributes[ProviderNameKey])
	}

	if event.Attributes[ResultValueKey] != true {
		t.Errorf("Expected event attribute 'VALUE' to be 'true', got '%v'", event.Attributes[ResultValueKey])
	}
}

func TestCreateEvaluationEvent_1_4_6_WithVariant(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, true, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: true,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey:  flagKey,
			FlagType: openfeature.Boolean,
			ResolutionDetail: openfeature.ResolutionDetail{
				Variant: "true",
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Name != "feature_flag.evaluation" {
		t.Errorf("Expected event name to be 'feature_flag.evaluation', got '%s'", event.Name)
	}

	if event.Attributes[FlagKey] != flagKey {
		t.Errorf("Expected event attribute 'KEY' to be '%s', got '%s'", flagKey, event.Attributes[FlagKey])
	}

	if event.Attributes[ResultVariantKey] != "true" {
		t.Errorf("Expected event attribute 'VARIANT' to be 'true', got '%s'", event.Attributes[ResultVariantKey])
	}
}

func TestCreateEvaluationEvent_1_4_14_WithFlagMetaData(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, false, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: false,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey:  flagKey,
			FlagType: openfeature.Boolean,
			ResolutionDetail: openfeature.ResolutionDetail{
				FlagMetadata: openfeature.FlagMetadata{
					flagMetaFlagSetIDKey: "test-set",
					flagMetaContextIDKey: "metadata-context",
					flagMetaVersionKey:   "v1.0",
				},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Attributes[FlagSetIDKey] != "test-set" {
		t.Errorf("Expected 'Flag SetID' in Flag Metadata name to be 'test-set', got '%s'", event.Attributes[flagMetaFlagSetIDKey])
	}

	if event.Attributes[ContextIDKey] != "metadata-context" {
		t.Errorf("Expected 'Flag ContextID' in Flag Metadata name to be 'metadata-context', got '%s'", event.Attributes[flagMetaContextIDKey])
	}

	if event.Attributes[VersionKey] != "v1.0" {
		t.Errorf("Expected 'Flag Version' in Flag Metadata name to be 'v1.0', got '%s'", event.Attributes[flagMetaVersionKey])
	}
}

func TestCreateEvaluationEvent_1_4_8_WithErrors(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, false, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: false,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey: flagKey,
			ResolutionDetail: openfeature.ResolutionDetail{
				Reason:       openfeature.ErrorReason,
				ErrorCode:    openfeature.FlagNotFoundCode,
				ErrorMessage: "a test error",
				FlagMetadata: openfeature.FlagMetadata{},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Attributes[ErrorTypeKey] != strings.ToLower(string(openfeature.FlagNotFoundCode)) {
		t.Errorf("Expected 'ERROR_CODE' to be 'GENERAL', got '%s'", event.Attributes[ErrorTypeKey])
	}

	if event.Attributes[ErrorMessageKey] != "a test error" {
		t.Errorf("Expected 'ERROR_MESSAGE' to be 'a test error', got '%s'", event.Attributes[ErrorMessageKey])
	}
}

func TestCreateEvaluationEvent_1_4_8_WithGeneralErrors(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, false, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: false,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey: flagKey,
			ResolutionDetail: openfeature.ResolutionDetail{
				Reason:       openfeature.ErrorReason,
				ErrorMessage: "a test error",
				FlagMetadata: openfeature.FlagMetadata{},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Attributes[ErrorTypeKey] != strings.ToLower(string(openfeature.GeneralCode)) {
		t.Errorf("Expected 'ERROR_CODE' to be 'GENERAL', got '%s'", event.Attributes[ErrorTypeKey])
	}

	if event.Attributes[ErrorMessageKey] != "a test error" {
		t.Errorf("Expected 'ERROR_MESSAGE' to be 'a test error', got '%s'", event.Attributes[ErrorMessageKey])
	}
}

func TestCreateEvaluationEvent_1_4_7_WithUnknownReason(t *testing.T) {
	flagKey := "test-flag"

	mockProviderMetadata := openfeature.Metadata{
		Name: "test-provider",
	}

	mockClientMetadata := openfeature.NewClientMetadata("test-client")

	mockEvalCtx := openfeature.NewEvaluationContext(
		"test-target-key", map[string]any{
			"is": "a test",
		})

	mockHookContext := openfeature.NewHookContext(flagKey, openfeature.Boolean, true, mockClientMetadata, mockProviderMetadata, mockEvalCtx)

	mockDetails := openfeature.InterfaceEvaluationDetails{
		Value: true,
		EvaluationDetails: openfeature.EvaluationDetails{
			FlagKey: flagKey,
			ResolutionDetail: openfeature.ResolutionDetail{
				FlagMetadata: openfeature.FlagMetadata{},
			},
		},
	}

	event := CreateEvaluationEvent(mockHookContext, mockDetails)

	if event.Attributes[ResultReasonKey] != strings.ToLower(string(openfeature.UnknownReason)) {
		t.Errorf("Expected evaluation reason to be '%s', got '%s'", strings.ToLower(string(openfeature.UnknownReason)), event.Attributes[ResultReasonKey])
	}
}
