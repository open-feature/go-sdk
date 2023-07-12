package testing

import (
	"context"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"testing"
)

func TestInMemoryProvider_boolean(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]interface{}{
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
			t.Errorf("incorect evaluation, expected %t, got %t", true, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_String(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"stringFlag": {
			Key:            "stringFlag",
			State:          Enabled,
			DefaultVariant: "stringOne",
			Variants: map[string]interface{}{
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
			t.Errorf("incorect evaluation, expected %s, got %s", "hello", evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Float(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"floatFlag": {
			Key:            "floatFlag",
			State:          Enabled,
			DefaultVariant: "fOne",
			Variants: map[string]interface{}{
				"fOne": 1.1,
				"fTwo": 2.2,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := context.Background()

	t.Run("test float success", func(t *testing.T) {
		evaluation := memoryProvider.FloatEvaluation(ctx, "fOne", 1.0, nil)

		if evaluation.Value != 1.0 {
			t.Errorf("incorect evaluation, expected %f, got %f", 1.0, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Int(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"intFlag": {
			Key:            "intFlag",
			State:          Enabled,
			DefaultVariant: "max",
			Variants: map[string]interface{}{
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
			t.Errorf("incorect evaluation, expected %d, got %d", 1, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_Object(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"objectFlag": {
			Key:            "objectFlag",
			State:          Enabled,
			DefaultVariant: "A",
			Variants: map[string]interface{}{
				"A": "SomeResult",
				"B": "OtherResult",
			},
		},
	})

	ctx := context.Background()

	t.Run("test object success", func(t *testing.T) {
		evaluation := memoryProvider.ObjectEvaluation(ctx, "objectFlag", "unknown", nil)

		if evaluation.Value != "SomeResult" {
			t.Errorf("incorect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}
	})
}

func TestInMemoryProvider_WithContext(t *testing.T) {
	var variantKey = "VariantSelector"

	// simple context handling - variant is selected from key and returned
	var evaluator = func(callerFlag InMemoryFlag, evalCtx openfeature.FlattenedContext) (interface{}, openfeature.ProviderResolutionDetail) {
		s := evalCtx[variantKey]
		return callerFlag.Variants[s.(string)], openfeature.ProviderResolutionDetail{}
	}

	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"contextFlag": {
			Key:            "contextFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]interface{}{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: &evaluator,
		},
	})

	ctx := context.Background()

	t.Run("test with context", func(t *testing.T) {

		evaluation := memoryProvider.BooleanEvaluation(ctx, "contextFlag", true, map[string]interface{}{
			variantKey: "false",
		})

		if evaluation.Value != false {
			t.Errorf("incorect evaluation, expected %v, got %v", false, evaluation.Value)
		}
	})
}

func TestInMemoryProvider_MissingFlag(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{})

	ctx := context.Background()

	t.Run("test missing flag", func(t *testing.T) {
		evaluation := memoryProvider.StringEvaluation(ctx, "missing-flag", "GoodBye", nil)

		if evaluation.Value != "GoodBye" {
			t.Errorf("incorect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}

		if evaluation.Reason != openfeature.ErrorReason {
			t.Errorf("incorect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}

		if evaluation.ResolutionDetail().ErrorCode != openfeature.FlagNotFoundCode {
			t.Errorf("incorect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.ResolutionDetail().ErrorCode)
		}
	})
}

func TestInMemoryProvider_TypeMismatch(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]interface{}{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := context.Background()

	t.Run("test missing flag", func(t *testing.T) {
		evaluation := memoryProvider.StringEvaluation(ctx, "boolFlag", "GoodBye", nil)

		if evaluation.Value != "GoodBye" {
			t.Errorf("incorect evaluation, expected %v, got %v", "SomeResult", evaluation.Value)
		}

		if evaluation.ResolutionDetail().ErrorCode != openfeature.TypeMismatchCode {
			t.Errorf("incorect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}
	})
}

func TestInMemoryProvider_Disabled(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          Disabled,
			DefaultVariant: "true",
			Variants: map[string]interface{}{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	ctx := context.Background()

	t.Run("test missing flag", func(t *testing.T) {
		evaluation := memoryProvider.BooleanEvaluation(ctx, "boolFlag", false, nil)

		if evaluation.Value != false {
			t.Errorf("incorect evaluation, expected %v, got %v", false, evaluation.Value)
		}

		if evaluation.Reason != openfeature.DisabledReason {
			t.Errorf("incorect reason, expected %v, got %v", openfeature.ErrorReason, evaluation.Reason)
		}
	})
}
