package openfeaturetest

import (
	"context"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
	"testing"
)

func TestParallelSingletonUsage(t *testing.T) {
	t.Parallel()

	testProvider := NewTestAwareProvider()
	err := openfeature.GetApiInstance().SetProvider(testProvider)
	if err != nil {
		t.Errorf("unable to set provider on TestAwareProvider")
	}

	tests := map[string]struct {
		givenProvider openfeature.FeatureProvider
		want          bool
	}{
		"test when flag is true": {
			givenProvider: memprovider.NewInMemoryProvider(
				map[string]memprovider.InMemoryFlag{
					"some_cool_feature": {
						State:          memprovider.Enabled,
						DefaultVariant: "variant_1",
						Variants: map[string]any{
							"variant_1": true,
						},
						ContextEvaluator: nil,
					},
				},
			),
			want: true,
		},
		"test when flag is false": {
			givenProvider: memprovider.NewInMemoryProvider(
				map[string]memprovider.InMemoryFlag{
					"some_cool_feature": {
						State:          memprovider.Enabled,
						DefaultVariant: "variant_1",
						Variants: map[string]any{
							"variant_1": false,
						},
						ContextEvaluator: nil,
					},
					"f2": {
						State:          memprovider.Enabled,
						DefaultVariant: "variant_1",
						Variants: map[string]any{
							"variant_1": "v1",
						},
						ContextEvaluator: nil,
					},
				},
			),
			want: false,
		},
	}

	for name, tt := range tests {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			defer testProvider.Cleanup()
			t.Parallel()
			testProvider.SetProvider(t, tt.givenProvider)

			// < CODE UNDER TEST >
			got := openfeature.GetApiInstance().
				GetClient().
				Boolean(context.TODO(), "some_cool_feature", false, openfeature.EvaluationContext{})
			// </ CODE UNDER TEST >

			if got != tt.want {
				t.Fatalf("uh oh, value is not as expected: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestAwareProvider(t *testing.T) {
	taw := NewTestAwareProvider()

	memProvider := memprovider.NewInMemoryProvider(
		map[string]memprovider.InMemoryFlag{
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
		},
	)

	t.Run("test bool evaluation", func(t *testing.T) {
		taw.SetProvider(t, memProvider)
		result := taw.BooleanEvaluation(context.TODO(), "ff-bool", false, openfeature.FlattenedContext{})
		if result.Value != true {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test string evaluation", func(t *testing.T) {
		taw.SetProvider(t, memProvider)
		result := taw.StringEvaluation(context.TODO(), "ff-string", "otherStr", openfeature.FlattenedContext{})
		if result.Value != "str" {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test int evaluation", func(t *testing.T) {
		taw.SetProvider(t, memProvider)
		result := taw.IntEvaluation(context.TODO(), "ff-int", int64(2), openfeature.FlattenedContext{})
		if result.Value != 1 {
			t.Errorf("got %v, want %v", result, true)
		}
	})

	t.Run("test float evaluation", func(t *testing.T) {
		taw.SetProvider(t, memProvider)
		result := taw.FloatEvaluation(context.TODO(), "ff-float", float64(2), openfeature.FlattenedContext{})
		if result.Value != float64(1) {
			t.Errorf("got %v, want %v", result, true)
		}
	})
	t.Run("test obj evaluation", func(t *testing.T) {
		taw.SetProvider(t, memProvider)
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

		taw := NewTestAwareProvider()
		taw.BooleanEvaluation(context.TODO(), "my-flag", true, openfeature.FlattenedContext{})
	})
}
