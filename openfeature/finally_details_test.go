package openfeature

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func evaluationParts[T any](details GenericEvaluationDetails[T], err error) (any, EvaluationDetails, error) {
	return details.Value, details.EvaluationDetails, err
}

// The evaluation details passed to finally match those returned to the application
// (4.3.8). An after-hook error returns the default value (4.4.8) and reports abnormal
// execution in the error code and reason (1.4.7 and 1.4.8).
func TestRequirement_4_3_8_AfterHookError(t *testing.T) {
	resolution := ProviderResolutionDetail{
		Reason: TargetingMatchReason, Variant: "treatment", FlagMetadata: FlagMetadata{"source": "test"},
	}
	tests := []struct {
		name          string
		defaultValue  any
		resolvedValue any
		stub          func(*MockFeatureProvider)
		evaluate      func(*Client) (any, EvaluationDetails, error)
	}{
		{
			name: "boolean", defaultValue: false, resolvedValue: true,
			stub: func(p *MockFeatureProvider) {
				p.EXPECT().BooleanEvaluation(gomock.Any(), "flag", false, gomock.Any()).
					Return(BoolResolutionDetail{Value: true, ProviderResolutionDetail: resolution})
			},
			evaluate: func(c *Client) (any, EvaluationDetails, error) {
				return evaluationParts(c.BooleanValueDetails(t.Context(), "flag", false, EvaluationContext{}))
			},
		},
		{
			name: "string", defaultValue: "default", resolvedValue: "resolved",
			stub: func(p *MockFeatureProvider) {
				p.EXPECT().StringEvaluation(gomock.Any(), "flag", "default", gomock.Any()).
					Return(StringResolutionDetail{Value: "resolved", ProviderResolutionDetail: resolution})
			},
			evaluate: func(c *Client) (any, EvaluationDetails, error) {
				return evaluationParts(c.StringValueDetails(t.Context(), "flag", "default", EvaluationContext{}))
			},
		},
		{
			name: "float", defaultValue: 1.0, resolvedValue: 2.0,
			stub: func(p *MockFeatureProvider) {
				p.EXPECT().FloatEvaluation(gomock.Any(), "flag", 1.0, gomock.Any()).
					Return(FloatResolutionDetail{Value: 2.0, ProviderResolutionDetail: resolution})
			},
			evaluate: func(c *Client) (any, EvaluationDetails, error) {
				return evaluationParts(c.FloatValueDetails(t.Context(), "flag", 1.0, EvaluationContext{}))
			},
		},
		{
			name: "integer", defaultValue: int64(1), resolvedValue: int64(2),
			stub: func(p *MockFeatureProvider) {
				p.EXPECT().IntEvaluation(gomock.Any(), "flag", int64(1), gomock.Any()).
					Return(IntResolutionDetail{Value: 2, ProviderResolutionDetail: resolution})
			},
			evaluate: func(c *Client) (any, EvaluationDetails, error) {
				return evaluationParts(c.IntValueDetails(t.Context(), "flag", 1, EvaluationContext{}))
			},
		},
		{
			name: "object", defaultValue: map[string]any{"default": true}, resolvedValue: map[string]any{"resolved": true},
			stub: func(p *MockFeatureProvider) {
				p.EXPECT().ObjectEvaluation(gomock.Any(), "flag", map[string]any{"default": true}, gomock.Any()).
					Return(InterfaceResolutionDetail{Value: map[string]any{"resolved": true}, ProviderResolutionDetail: resolution})
			},
			evaluate: func(c *Client) (any, EvaluationDetails, error) {
				return evaluationParts(c.ObjectValueDetails(t.Context(), "flag", map[string]any{"default": true}, EvaluationContext{}))
			},
		},
	}
	for _, tc := range tests {
		for _, fail := range []bool{true, false} {
			name := tc.name + "/success"
			if fail {
				name = tc.name + "/after-error"
			}
			t.Run(name, func(t *testing.T) {
				m := hydratedMocksForClientTests(t, 1)
				tc.stub(m.providerAPI)
				hook := NewMockHook(gomock.NewController(t))
				client := &Client{providerBinding: m.providerBinding, clientEventing: m.clientHandlerAPI, hooks: []Hook{hook}}
				hook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any())
				var hookErr error
				if fail {
					hookErr = errors.New("application validation failed")
				}
				after := hook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(hookErr)
				if fail {
					after = hook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Do(func(_ context.Context, _ HookContext, err error, _ HookHints) { assert.ErrorIs(t, err, hookErr) }).After(after)
				}
				var final InterfaceEvaluationDetails
				hook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, _ HookContext, details InterfaceEvaluationDetails, _ HookHints) {
						final = details
					}).After(after)
				value, details, err := tc.evaluate(client)
				if fail {
					require.ErrorIs(t, err, hookErr)
					assert.Equal(t, tc.defaultValue, value)
					assert.Equal(t, ErrorReason, details.Reason)
					assert.Equal(t, GeneralCode, details.ErrorCode)
					assert.NotEmpty(t, details.ErrorMessage)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.resolvedValue, value)
					assert.Equal(t, TargetingMatchReason, details.Reason)
					assert.Empty(t, details.ErrorCode)
				}
				assert.Equal(t, value, final.Value)
				assert.Equal(t, details, final.EvaluationDetails)
			})
		}
	}
}
