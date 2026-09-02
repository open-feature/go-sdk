package openfeature_test

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

// testConfig is the target type for the object evaluations under test. Secret is
// dropped by JSON, so it survives only when the provider's value is used as-is
// rather than round-tripped.
type testConfig struct {
	Text       string `json:"text"`
	Percentage int    `json:"percentage"`
	Secret     string `json:"-"`
}

// newObjectClient returns a client bound to an isolated API instance serving a
// single object flag named "config" with the given value, so that tests neither
// mutate nor leak from the global singleton.
func newObjectClient(t *testing.T, value any) *openfeature.Client {
	t.Helper()

	api := isolated.NewAPI()
	t.Cleanup(func() {
		if err := api.Shutdown(t.Context()); err != nil {
			t.Errorf("shutting down isolated api: %v", err)
		}
	})

	provider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
		"config": {
			Key:            "config",
			State:          memprovider.Enabled,
			DefaultVariant: "default",
			Variants:       map[string]any{"default": value},
		},
	})
	if err := api.SetProviderAndWait(t.Context(), &provider); err != nil {
		t.Fatalf("setting provider: %v", err)
	}

	return api.NewClient()
}

// TestRequirement_1_3_4 The client SHOULD guarantee the returned value of any typed flag
// evaluation method is of the expected type. If the value returned by the underlying
// provider implementation does not match the expected type, it's to be considered abnormal
// execution, and the supplied default value should be returned.
func TestRequirement_1_3_4(t *testing.T) {
	defaultValue := testConfig{Text: "N/A", Percentage: 75}

	tests := map[string]struct {
		value       any
		flag        string
		want        testConfig
		wantErr     bool
		wantCode    openfeature.ErrorCode
		wantReason  openfeature.Reason
		wantVariant string
	}{
		"value of the target type is used as-is": {
			value:       testConfig{Text: "hello", Percentage: 25, Secret: "kept"},
			want:        testConfig{Text: "hello", Percentage: 25, Secret: "kept"},
			wantReason:  openfeature.StaticReason,
			wantVariant: "default",
		},
		"structure is round-tripped through JSON": {
			value:       map[string]any{"text": "hello", "percentage": 25, "secret": "dropped"},
			want:        testConfig{Text: "hello", Percentage: 25},
			wantReason:  openfeature.StaticReason,
			wantVariant: "default",
		},
		"value that cannot be unmarshalled returns the default value": {
			value:       "not-an-object",
			want:        defaultValue,
			wantErr:     true,
			wantCode:    openfeature.TypeMismatchCode,
			wantReason:  openfeature.ErrorReason,
			wantVariant: "default",
		},
		// Channels are not serializable, so the round-trip fails at marshal time.
		"value that cannot be marshalled returns the default value": {
			value:       make(chan int),
			want:        defaultValue,
			wantErr:     true,
			wantCode:    openfeature.TypeMismatchCode,
			wantReason:  openfeature.ErrorReason,
			wantVariant: "default",
		},
		"failed evaluation returns the default value": {
			value:      testConfig{Text: "hello", Percentage: 25},
			flag:       "missing",
			want:       defaultValue,
			wantErr:    true,
			wantCode:   openfeature.FlagNotFoundCode,
			wantReason: openfeature.ErrorReason,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := newObjectClient(t, test.value)

			flag := test.flag
			if flag == "" {
				flag = "config"
			}

			details, err := openfeature.GetObjectValueDetails(
				t.Context(), client, flag, defaultValue, openfeature.EvaluationContext{},
			)

			switch {
			case test.wantErr && err == nil:
				t.Fatal("got nil error, want an error")
			case !test.wantErr && err != nil:
				t.Fatalf("got error %v, want nil", err)
			}

			if details.Value != test.want {
				t.Errorf("got value %+v, want %+v", details.Value, test.want)
			}
			if details.ErrorCode != test.wantCode {
				t.Errorf("got error code %q, want %q", details.ErrorCode, test.wantCode)
			}
			if test.wantErr && details.ErrorMessage == "" {
				t.Error("got empty error message, want a message")
			}
			if details.Reason != test.wantReason {
				t.Errorf("got reason %q, want %q", details.Reason, test.wantReason)
			}

			// Resolution metadata is carried on every return path.
			if details.FlagKey != flag {
				t.Errorf("got flag key %q, want %q", details.FlagKey, flag)
			}
			if details.FlagType != openfeature.Object {
				t.Errorf("got flag type %v, want %v", details.FlagType, openfeature.Object)
			}
			if details.Variant != test.wantVariant {
				t.Errorf("got variant %q, want %q", details.Variant, test.wantVariant)
			}
		})
	}
}

// TestGetObjectValueDetailsForAny verifies that a value is returned untouched when
// the target type imposes no constraint.
func TestGetObjectValueDetailsForAny(t *testing.T) {
	client := newObjectClient(t, map[string]any{"text": "hello"})

	details, err := openfeature.GetObjectValueDetails[any](
		t.Context(), client, "config", nil, openfeature.EvaluationContext{},
	)
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}

	value, ok := details.Value.(map[string]any)
	if !ok {
		t.Fatalf("got value of type %T, want map[string]any", details.Value)
	}
	if value["text"] != "hello" {
		t.Errorf("got text %v, want hello", value["text"])
	}
}

// TestGetObjectValueDetailsForwardsOptions verifies that evaluation options reach
// the underlying evaluation.
func TestGetObjectValueDetailsForwardsOptions(t *testing.T) {
	client := newObjectClient(t, testConfig{Text: "hello", Percentage: 25})

	hook := &recordingHook{}
	if _, err := openfeature.GetObjectValueDetails(
		t.Context(), client, "config", testConfig{}, openfeature.EvaluationContext{},
		openfeature.WithHooks(hook),
	); err != nil {
		t.Fatalf("got error %v, want nil", err)
	}

	if !hook.called {
		t.Error("got an unexecuted invocation hook, want it executed")
	}
}

func TestGetObjectValue(t *testing.T) {
	defaultValue := testConfig{Text: "N/A", Percentage: 75}
	stored := testConfig{Text: "hello", Percentage: 25}

	tests := map[string]struct {
		flag    string
		want    testConfig
		wantErr bool
	}{
		"returns the value":                       {flag: "config", want: stored},
		"returns the default value and the error": {flag: "missing", want: defaultValue, wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := newObjectClient(t, stored)

			got, err := openfeature.GetObjectValue(
				t.Context(), client, test.flag, defaultValue, openfeature.EvaluationContext{},
			)

			switch {
			case test.wantErr && err == nil:
				t.Fatal("got nil error, want an error")
			case !test.wantErr && err != nil:
				t.Fatalf("got error %v, want nil", err)
			}

			if got != test.want {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestGetObject(t *testing.T) {
	defaultValue := testConfig{Text: "N/A", Percentage: 75}
	stored := testConfig{Text: "hello", Percentage: 25}

	tests := map[string]struct {
		flag string
		want testConfig
	}{
		"returns the value":                          {flag: "config", want: stored},
		"swallows the error and returns the default": {flag: "missing", want: defaultValue},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := newObjectClient(t, stored)

			got := openfeature.GetObject(
				t.Context(), client, test.flag, defaultValue, openfeature.EvaluationContext{},
			)
			if got != test.want {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

// recordingHook records whether it was invoked.
type recordingHook struct {
	openfeature.UnimplementedHook

	called bool
}

func (h *recordingHook) Before(
	ctx context.Context, hookContext openfeature.HookContext, hookHints openfeature.HookHints,
) (*openfeature.EvaluationContext, error) {
	h.called = true

	return nil, nil
}
