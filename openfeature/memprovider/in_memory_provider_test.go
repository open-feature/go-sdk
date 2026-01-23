package memprovider

import (
	"math"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
)

func TestInMemoryProvider_boolean(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := t.Context()

	t.Run("test boolean success", func(t *testing.T) {
		evaluation := memoryProvider.BooleanEvaluation(ctx, "boolFlag", false, nil)

		if evaluation.Value != true {
			t.Errorf("incorrect evaluation, expected %t, got %t", true, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_String(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"stringFlag": {
			Key:            "stringFlag",
			State:          Enabled,
			DefaultVariant: "stringOne",
			Variants: map[string]any{
				"stringOne": "hello",
				"stringTwo": "GoodBye",
			},
			ContextEvaluator: nil,
		},
	})

	ctx := t.Context()

	t.Run("test string success", func(t *testing.T) {
		evaluation := memoryProvider.StringEvaluation(ctx, "stringFlag", "none", nil)

		if evaluation.Value != "hello" {
			t.Errorf("incorrect evaluation, expected %s, got %s", "hello", evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Float(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"floatFlag": {
			Key:            "floatFlag",
			State:          Enabled,
			DefaultVariant: "fOne",
			Variants: map[string]any{
				"fOne": 1.1,
				"fTwo": 2.2,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := t.Context()

	t.Run("test float success", func(t *testing.T) {
		evaluation := memoryProvider.FloatEvaluation(ctx, "floatFlag", 1.0, nil)

		if evaluation.Value != 1.1 {
			t.Errorf("incorrect evaluation, expected %f, got %f", 1.1, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Int(t *testing.T) {
	// Test that both int and int64 variants work correctly.
	// The provider coerces int to int64 internally to match the API contract.
	tests := []struct {
		name         string
		variant      any
		defaultValue int64
		expected     int64
	}{
		{
			name:         "int64 max value",
			variant:      int64(math.MaxInt64),
			defaultValue: 1,
			expected:     math.MaxInt64,
		},
		{
			name:         "int64 min value",
			variant:      int64(math.MinInt64),
			defaultValue: 1,
			expected:     math.MinInt64,
		},
		{
			name:         "plain int coerced to int64",
			variant:      42,
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "int8 coerced to int64",
			variant:      int8(8),
			defaultValue: 0,
			expected:     8,
		},
		{
			name:         "int16 coerced to int64",
			variant:      int16(16),
			defaultValue: 0,
			expected:     16,
		},
		{
			name:         "int32 coerced to int64",
			variant:      int32(32),
			defaultValue: 0,
			expected:     32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
				"intFlag": {
					State:          Enabled,
					DefaultVariant: "value",
					Variants:       map[string]any{"value": tt.variant},
				},
			})

			evaluation := memoryProvider.IntEvaluation(t.Context(), "intFlag", tt.defaultValue, nil)

			if evaluation.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, evaluation.Value)
			}
		})
	}
}

func TestInMemoryProvider_Object(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"objectFlag": {
			Key:            "objectFlag",
			State:          Enabled,
			DefaultVariant: "A",
			Variants: map[string]any{
				"A": "SomeResult",
				"B": "OtherResult",
			},
		},
	})

	ctx := t.Context()

	t.Run("test object success", func(t *testing.T) {
		evaluation := memoryProvider.ObjectEvaluation(ctx, "objectFlag", "unknown", nil)

		if evaluation.Value != "SomeResult" {
			t.Errorf("incorrect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}
	})
}

func TestInMemoryProvider_WithContext(t *testing.T) {
	variantKey := "VariantSelector"

	// simple context handling - variant is selected from key and returned
	evaluator := func(callerFlag InMemoryFlag, flatCtx openfeature.FlattenedContext) (any, openfeature.ProviderResolutionDetail) {
		s := flatCtx[variantKey]
		return callerFlag.Variants[s.(string)], openfeature.ProviderResolutionDetail{}
	}

	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"contextFlag": {
			Key:            "contextFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: &evaluator,
		},
	})

	ctx := t.Context()

	t.Run("test with context", func(t *testing.T) {
		evaluation := memoryProvider.BooleanEvaluation(ctx, "contextFlag", true, map[string]any{
			variantKey: "false",
		})

		if evaluation.Value != false {
			t.Errorf("incorrect evaluation, expected %v, got %v", false, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_MissingFlag(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{})

	ctx := t.Context()

	t.Run("test missing flag", func(t *testing.T) {
		evaluation := memoryProvider.StringEvaluation(ctx, "missing-flag", "GoodBye", nil)

		if evaluation.Value != "GoodBye" {
			t.Errorf("incorrect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}

		if evaluation.Reason != openfeature.ErrorReason {
			t.Errorf("incorrect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}

		if evaluation.ResolutionDetail().ErrorCode != openfeature.FlagNotFoundCode {
			t.Errorf("incorrect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.ResolutionDetail().ErrorCode)
		}
	})
}

func TestInMemoryProvider_TypeMismatch(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := t.Context()

	t.Run("test type mismatch flag", func(t *testing.T) {
		evaluation := memoryProvider.StringEvaluation(ctx, "boolFlag", "GoodBye", nil)

		if evaluation.Value != "GoodBye" {
			t.Errorf("incorrect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}

		if evaluation.ResolutionDetail().ErrorCode != openfeature.TypeMismatchCode {
			t.Errorf("incorrect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}
	})
}

func TestInMemoryProvider_Disabled(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Disabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := t.Context()

	t.Run("test disabled flag", func(t *testing.T) {
		evaluation := memoryProvider.BooleanEvaluation(ctx, "boolFlag", false, nil)

		if evaluation.Value != false {
			t.Errorf("incorrect evaluation, expected %v, got %v", false, evaluation.Value)
		}

		if evaluation.Reason != openfeature.DisabledReason {
			t.Errorf("incorrect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}
	})
}

func TestInMemoryProvider_Metadata(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{})

	metadata := memoryProvider.Metadata()

	if metadata.Name == "" {
		t.Errorf("expected non-empty name for in-memory provider")
	}

	if metadata.Name != "InMemoryProvider" {
		t.Errorf("incorrect name for in-memory provider")
	}
}

func TestInMemoryProvider_Track(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{})
	memoryProvider.Track(t.Context(), "example-event-name", openfeature.EvaluationContext{}, openfeature.TrackingEventDetails{})
}
