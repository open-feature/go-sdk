package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.openfeature.dev/openfeature/v2"
	memprovider "go.openfeature.dev/openfeature/v2/providers/inmemory"
)

func TestParallelSingletonUsage(t *testing.T) {
	t.Parallel()

	testProvider := NewProvider()
	err := openfeature.SetProviderAndWait(t.Context(), testProvider)
	if err != nil {
		t.Errorf("unable to set provider")
	}

	tests := map[string]struct {
		flags map[string]memprovider.InMemoryFlag
		want  bool
	}{
		"test when flag is true": {
			flags: map[string]memprovider.InMemoryFlag{
				"my_flag": {
					State:          memprovider.Enabled,
					DefaultVariant: "on",
					Variants: map[string]any{
						"on": true,
					},
				},
			},
			want: true,
		},
		"test when flag is false": {
			flags: map[string]memprovider.InMemoryFlag{
				"my_flag": {
					State:          memprovider.Enabled,
					DefaultVariant: "off",
					Variants: map[string]any{
						"off": false,
					},
				},
			},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer testProvider.Cleanup(t.Context())
			t.Parallel()
			testProvider.UsingFlags(t, tt.flags)

			got := functionUnderTest(t)

			if got != tt.want {
				t.Fatalf("uh oh, value is not as expected: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestAwareProvider(t *testing.T) {
	taw := NewProvider()

	flags := map[string]memprovider.InMemoryFlag{
		"ff-bool": {
			State:          memprovider.Enabled,
			DefaultVariant: "variant_1",
			Variants: map[string]any{
				"variant_1": true,
			},
		},
		"ff-string": {
			State:          memprovider.Enabled,
			DefaultVariant: "variant_1",
			Variants: map[string]any{
				"variant_1": "str",
			},
		},
		"ff-int": {
			State:          memprovider.Enabled,
			DefaultVariant: "variant_1",
			Variants: map[string]any{
				"variant_1": 1,
			},
		},
		"ff-float": {
			State:          memprovider.Enabled,
			DefaultVariant: "variant_1",
			Variants: map[string]any{
				"variant_1": float64(1),
			},
		},
		"ff-obj": {
			State:          memprovider.Enabled,
			DefaultVariant: "variant_1",
			Variants: map[string]any{
				"variant_1": "obj",
			},
		},
	}

	t.Run("test bool evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.BooleanEvaluation(t.Context(), "ff-bool", false, openfeature.FlattenedContext{})
		if result.Value != true {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test string evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.StringEvaluation(t.Context(), "ff-string", "otherStr", openfeature.FlattenedContext{})
		if result.Value != "str" {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test int evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.IntEvaluation(t.Context(), "ff-int", int64(2), openfeature.FlattenedContext{})
		if result.Value != 1 {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test float evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.FloatEvaluation(t.Context(), "ff-float", float64(2), openfeature.FlattenedContext{})
		if result.Value != float64(1) {
			t.Errorf("got %v, want %v", result, true)
		}
	})
	t.Run("test obj evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.ObjectEvaluation(t.Context(), "ff-obj", "stringobj", openfeature.FlattenedContext{})
		if result.Value != "obj" {
			t.Errorf("got %v, want %v", result, true)
		}
	})
}

func Test_TestAwareProviderPanics(t *testing.T) {
	t.Run("provider panics if no test name was provided by calling SetProvider()", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("the test aware provider did not panic")
			}
		}()

		taw := NewProvider()
		taw.BooleanEvaluation(t.Context(), "my-flag", true, openfeature.FlattenedContext{})
	})
}

func TestServeWithAnotherGoroutine(t *testing.T) {
	testProvider := NewProvider()
	ctx := testProvider.UsingFlags(t, map[string]memprovider.InMemoryFlag{
		"myflag": {
			DefaultVariant: "defaultVariant",
			Variants:       map[string]any{"defaultVariant": true},
		},
	})
	t.Cleanup(func() {
		testProvider.Cleanup(ctx)
	})

	err := openfeature.SetProviderAndWait(ctx, testProvider)
	require.NoError(t, err)

	handlerDone := make(chan struct{})
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		_ = openfeature.NewClient().Boolean(ctx, "myflag", false, openfeature.EvaluationContextFromContext(ctx))

		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusOK)
		}
		close(handlerDone)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/drain", nil)
	w := httptest.NewRecorder()

	// Start the handler in a goroutine for _reasons_
	// This is what triggers the TestProvider bug
	go func() {
		handler(w, req)
	}()

	timedout := false

	// Wait for handler to complete
	select {
	case <-handlerDone:
		assert.Equal(t, http.StatusOK, w.Code)
	case <-time.After(time.Second):
		t.Log("drain not completed within timeout")
		timedout = true
	}

	assert.Equal(t, http.StatusOK, w.Code)
	require.False(t, timedout)
}

func functionUnderTest(tb testing.TB) bool {
	tb.Helper()
	got := openfeature.NewClient().
		Boolean(tb.Context(), "my_flag", false, openfeature.EvaluationContext{})
	return got
}
