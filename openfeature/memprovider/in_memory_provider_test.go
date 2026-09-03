package memprovider

import (
	"math"
	"slices"
	"sync"
	"testing"
	"time"

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
		"float32Flag": {
			Key:            "float32Flag",
			State:          Enabled,
			DefaultVariant: "fOne",
			Variants: map[string]any{
				"fOne": float32(3.5),
				"fTwo": float32(4.5),
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

	t.Run("test float32 conversion success", func(t *testing.T) {
		evaluation := memoryProvider.FloatEvaluation(ctx, "float32Flag", 1.0, nil)
		expected := 3.5
		if evaluation.Value != expected {
			t.Errorf("incorrect evaluation, expected %f, got %f", expected, evaluation.Value)
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

func TestInMemoryProvider_NilContextEvaluatorFunc(t *testing.T) {
	// ContextEvaluator is a non-nil pointer to a nil func. The pointer passes a
	// plain != nil guard, so a naive dereference would panic. The flag must fall
	// back to the default variant instead.
	var nilFn func(callerFlag InMemoryFlag, flatCtx openfeature.FlattenedContext) (any, openfeature.ProviderResolutionDetail)

	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"nilFnFlag": {
			Key:            "nilFnFlag",
			State:          Enabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: &nilFn,
		},
	})

	ctx := t.Context()

	evaluation := memoryProvider.BooleanEvaluation(ctx, "nilFnFlag", false, nil)

	if evaluation.Value != true {
		t.Errorf("incorrect evaluation, expected %v, got %v", true, evaluation.Value)
	}
	if evaluation.Variant != "true" {
		t.Errorf("incorrect variant, expected %q, got %q", "true", evaluation.Variant)
	}
	if evaluation.Reason != openfeature.StaticReason {
		t.Errorf("incorrect reason, expected %q, got %q", openfeature.StaticReason, evaluation.Reason)
	}
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

func boolFlag(key, variant string) InMemoryFlag {
	return InMemoryFlag{
		Key:            key,
		State:          Enabled,
		DefaultVariant: variant,
		Variants:       map[string]any{"true": true, "false": false},
	}
}

// waitForEvent reads the next event off the provider's channel.
func waitForEvent(t *testing.T, provider *InMemoryProvider) openfeature.Event {
	t.Helper()

	select {
	case event := <-provider.EventChannel():
		return event
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout - no event emitted")
		return openfeature.Event{}
	}
}

func TestInMemoryProvider_UpdateFlags(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "true"),
	})

	if got := memoryProvider.BooleanEvaluation(t.Context(), "flagA", false, nil); got.Value != true {
		t.Errorf("expected flagA to resolve true before the update, got %v", got.Value)
	}

	memoryProvider.UpdateFlags(map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "false"),
		"flagB": boolFlag("flagB", "true"),
	})

	if got := memoryProvider.BooleanEvaluation(t.Context(), "flagA", true, nil); got.Value != false {
		t.Errorf("expected flagA to resolve false after the update, got %v", got.Value)
	}

	if got := memoryProvider.BooleanEvaluation(t.Context(), "flagB", false, nil); got.Value != true {
		t.Errorf("expected flagB to resolve true after the update, got %v", got.Value)
	}
}

func TestInMemoryProvider_UpdateFlagsRemovesFlags(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "true"),
	})

	memoryProvider.UpdateFlags(map[string]InMemoryFlag{})

	detail := memoryProvider.BooleanEvaluation(t.Context(), "flagA", false, nil)
	if detail.Value != false {
		t.Errorf("expected the default value for a removed flag, got %v", detail.Value)
	}

	if detail.ResolutionError.Error() == "" {
		t.Error("expected a resolution error for a removed flag")
	}

	if detail.Reason != openfeature.ErrorReason {
		t.Errorf("expected %s, got %s", openfeature.ErrorReason, detail.Reason)
	}
}

func TestInMemoryProvider_UpdateFlagsEmitsConfigurationChanged(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"removed": boolFlag("removed", "true"),
		"kept":    boolFlag("kept", "true"),
	})

	memoryProvider.UpdateFlags(map[string]InMemoryFlag{
		"kept":  boolFlag("kept", "false"),
		"added": boolFlag("added", "true"),
	})

	event := waitForEvent(t, memoryProvider)

	if event.EventType != openfeature.ProviderConfigChange {
		t.Errorf("expected %s, got %s", openfeature.ProviderConfigChange, event.EventType)
	}

	if event.ProviderName != "InMemoryProvider" {
		t.Errorf("expected provider name InMemoryProvider, got %s", event.ProviderName)
	}

	// The union of all previous and all new flag keys, per Appendix A.
	want := []string{"added", "kept", "removed"}
	if !slices.Equal(event.FlagChanges, want) {
		t.Errorf("expected flag changes %v, got %v", want, event.FlagChanges)
	}
}

func TestInMemoryProvider_ConstructorCopiesFlags(t *testing.T) {
	flags := map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "true"),
	}

	memoryProvider := NewInMemoryProvider(flags)
	delete(flags, "flagA")
	flags["flagB"] = boolFlag("flagB", "true")

	if got := memoryProvider.BooleanEvaluation(t.Context(), "flagA", false, nil); got.Value != true {
		t.Error("mutating the caller's map must not affect the provider")
	}

	if got := memoryProvider.BooleanEvaluation(t.Context(), "flagB", false, nil); got.Value != false {
		t.Error("mutating the caller's map must not affect the provider")
	}
}

func TestInMemoryProvider_UpdateFlagsDoesNotBlockWithoutSubscriber(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Nothing drains the channel here, so this only returns if the send is
		// non-blocking once the buffer fills.
		for range eventChannelBuffer * 3 {
			memoryProvider.UpdateFlags(map[string]InMemoryFlag{"flagA": boolFlag("flagA", "true")})
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("UpdateFlags blocked with no subscriber draining the event channel")
	}
}

func TestInMemoryProvider_ConcurrentUpdateAndEvaluation(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "true"),
	})

	// Drain events so the writer is never throttled by a full buffer.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-memoryProvider.EventChannel():
			case <-stop:
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				memoryProvider.UpdateFlags(map[string]InMemoryFlag{"flagA": boolFlag("flagA", "false")})
			}
		})
		wg.Go(func() {
			for range 100 {
				memoryProvider.BooleanEvaluation(t.Context(), "flagA", false, nil)
			}
		})
		wg.Go(func() {
			for range 100 {
				memoryProvider.Track(t.Context(), "event", openfeature.EvaluationContext{}, openfeature.TrackingEventDetails{})
			}
		})
	}
	wg.Wait()
}

// TestInMemoryProvider_ConfigurationChangedReachesHandler covers the path the
// issue is about: an update on a registered provider reaching an API handler.
func TestInMemoryProvider_ConfigurationChangedReachesHandler(t *testing.T) {
	memoryProvider := NewInMemoryProvider(map[string]InMemoryFlag{
		"flagA": boolFlag("flagA", "true"),
	})

	if err := openfeature.SetProviderAndWait(memoryProvider); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(openfeature.Shutdown)

	received := make(chan openfeature.EventDetails, 1)
	callback := func(details openfeature.EventDetails) {
		received <- details
	}
	openfeature.AddHandler(openfeature.ProviderConfigChange, &callback)
	t.Cleanup(func() {
		openfeature.RemoveHandler(openfeature.ProviderConfigChange, &callback)
	})

	memoryProvider.UpdateFlags(map[string]InMemoryFlag{"flagB": boolFlag("flagB", "true")})

	select {
	case details := <-received:
		want := []string{"flagA", "flagB"}
		if !slices.Equal(details.FlagChanges, want) {
			t.Errorf("expected flag changes %v, got %v", want, details.FlagChanges)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout - PROVIDER_CONFIGURATION_CHANGED did not reach the handler")
	}
}
