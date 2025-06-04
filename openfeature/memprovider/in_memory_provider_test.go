package memprovider

import (
	"context"
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

	ctx := context.Background()

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

	ctx := context.Background()

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

	ctx := context.Background()

	t.Run("test float success", func(t *testing.T) {
		evaluation := memoryProvider.FloatEvaluation(ctx, "floatFlag", 1.0, nil)

		if evaluation.Value != 1.1 {
			t.Errorf("incorrect evaluation, expected %f, got %f", 1.1, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Int(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"intFlag": {
			Key:            "intFlag",
			State:          Enabled,
			DefaultVariant: "max",
			Variants: map[string]any{
				"min": -9223372036854775808,
				"max": 9223372036854775807,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := context.Background()

	t.Run("test integer success", func(t *testing.T) {
		evaluation := memoryProvider.IntEvaluation(ctx, "intFlag", 1, nil)

		if evaluation.Value != 9223372036854775807 {
			t.Errorf("incorrect evaluation, expected %d, got %d", 1, evaluation.Value)
		}
	})
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

	ctx := context.Background()

	t.Run("test object success", func(t *testing.T) {
		evaluation := memoryProvider.ObjectEvaluation(ctx, "objectFlag", "unknown", nil)

		if evaluation.Value != "SomeResult" {
			t.Errorf("incorrect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}
	})
}

func TestInMemoryProvider_WithContext(t *testing.T) {
	var variantKey = "VariantSelector"

	// simple context handling - variant is selected from key and returned
	var evaluator = func(callerFlag InMemoryFlag, evalCtx openfeature.FlattenedContext) (any, openfeature.ProviderResolutionDetail) {
		s := evalCtx[variantKey]
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

	ctx := context.Background()

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

	ctx := context.Background()

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

	ctx := context.Background()

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

	ctx := context.Background()

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
	memoryProvider.Track(context.Background(), "example-event-name", openfeature.EvaluationContext{}, openfeature.TrackingEventDetails{})
}
