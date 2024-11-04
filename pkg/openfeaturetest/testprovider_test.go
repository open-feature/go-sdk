package openfeaturetest

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

func TestParallelSingletonUsage(t *testing.T) {
	t.Parallel()

	testProvider := NewTestProvider()
	err := openfeature.GetApiInstance().SetProvider(testProvider)
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
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			defer testProvider.Cleanup()
			t.Parallel()
			testProvider.UsingFlags(t, tt.flags)

			got := functionUnderTest()

			if got != tt.want {
				t.Fatalf("uh oh, value is not as expected: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestAwareProvider(t *testing.T) {
	taw := NewTestProvider()

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
		result := taw.BooleanEvaluation(context.TODO(), "ff-bool", false, openfeature.FlattenedContext{})
		if result.Value != true {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test string evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.StringEvaluation(context.TODO(), "ff-string", "otherStr", openfeature.FlattenedContext{})
		if result.Value != "str" {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test int evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.IntEvaluation(context.TODO(), "ff-int", int64(2), openfeature.FlattenedContext{})
		if result.Value != 1 {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test float evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.FloatEvaluation(context.TODO(), "ff-float", float64(2), openfeature.FlattenedContext{})
		if result.Value != float64(1) {
			t.Errorf("got %v, want %v", result, true)
		}
	})
	t.Run("test obj evaluation", func(t *testing.T) {
		taw.UsingFlags(t, flags)
		result := taw.ObjectEvaluation(context.TODO(), "ff-obj", "stringobj", openfeature.FlattenedContext{})
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

		taw := NewTestProvider()
		taw.BooleanEvaluation(context.TODO(), "my-flag", true, openfeature.FlattenedContext{})
	})
}

func functionUnderTest() bool {
	got := openfeature.GetApiInstance().
		GetClient().
		Boolean(context.TODO(), "my_flag", false, openfeature.EvaluationContext{})
	return got
}
