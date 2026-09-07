package telemetry

import (
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
)

func TestCreateEvaluationEvent_ContextID(t *testing.T) {
	tests := []struct {
		name          string
		targetingKey  string
		flagMetadata  openfeature.FlagMetadata
		wantContextID string
		wantAttribute bool
	}{
		{
			name:         "metadata takes precedence",
			targetingKey: "targeting-context",
			flagMetadata: openfeature.FlagMetadata{
				flagMetaContextIDKey: "metadata-context",
			},
			wantContextID: "metadata-context",
			wantAttribute: true,
		},
		{
			name:          "targeting key is used as fallback",
			targetingKey:  "targeting-context",
			flagMetadata:  openfeature.FlagMetadata{},
			wantContextID: "targeting-context",
			wantAttribute: true,
		},
		{
			name:         "empty metadata context uses targeting key fallback",
			targetingKey: "targeting-context",
			flagMetadata: openfeature.FlagMetadata{
				flagMetaContextIDKey: "",
			},
			wantContextID: "targeting-context",
			wantAttribute: true,
		},
		{
			name:          "context attribute is omitted when unavailable",
			targetingKey:  "",
			flagMetadata:  openfeature.FlagMetadata{},
			wantAttribute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagKey := "test-flag"
			providerMetadata := openfeature.Metadata{Name: "test-provider"}
			clientMetadata := openfeature.NewClientMetadata("test-client")
			evaluationContext := openfeature.NewEvaluationContext(tt.targetingKey, map[string]any{})
			hookContext := openfeature.NewHookContext(
				flagKey,
				openfeature.Boolean,
				true,
				clientMetadata,
				providerMetadata,
				evaluationContext,
			)
			details := openfeature.InterfaceEvaluationDetails{
				Value: true,
				EvaluationDetails: openfeature.EvaluationDetails{
					FlagKey:  flagKey,
					FlagType: openfeature.Boolean,
					ResolutionDetail: openfeature.ResolutionDetail{
						FlagMetadata: tt.flagMetadata,
					},
				},
			}

			event := CreateEvaluationEvent(hookContext, details)
			gotContextID, ok := event.Attributes[ContextIDKey]

			if ok != tt.wantAttribute {
				t.Fatalf("context ID attribute presence = %v, want %v", ok, tt.wantAttribute)
			}
			if tt.wantAttribute && gotContextID != tt.wantContextID {
				t.Errorf("context ID = %v, want %q", gotContextID, tt.wantContextID)
			}
		})
	}
}
