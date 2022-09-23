package openfeature

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"

	"github.com/golang/mock/gomock"
)

// The client MUST provide a method to add `hooks` which accepts one or more API-conformant `hooks`,
// and appends them to the collection of any previously added hooks.
// When new hooks are added, previously added hooks are not removed.
func TestRequirement_1_2_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	client := NewClient("test-client")
	client.AddHooks(mockHook)
	client.AddHooks(mockHook, mockHook)

	if len(client.hooks) != 3 {
		t.Error("func client.AddHooks didn't append the list of hooks to the client's existing collection of hooks")
	}
}

// The client interface MUST define a `metadata` member or accessor,
// containing an immutable `name` field or accessor of type string,
// which corresponds to the `name` value supplied during client creation.
func TestRequirement_1_2_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	clientName := "test-client"
	client := NewClient(clientName)

	if client.Metadata().Name() != clientName {
		t.Errorf("client name not initiated as expected, got %s, want %s", client.Metadata().Name(), clientName)
	}
}

// TestRequirements_1_3 tests all the 1.3.* requirements by asserting that the returned client matches the interface
// defined by the 1.3.* requirements
//
// Requirement_1_3_1
// The `client` MUST provide methods for typed flag evaluation, including boolean, numeric, string,
// and structure, with parameters `flag key` (string, required), `default value` (boolean | number | string | structure, required),
// `evaluation context` (optional), and `evaluation options` (optional), which returns the flag value.
//
// Requirement_1_3_2_1
// The client SHOULD provide functions for floating-point numbers and integers, consistent with language idioms.
//
// Requirement_1_3_3
// The `client` SHOULD guarantee the returned value of any typed flag evaluation method is of the expected type.
// If the value returned by the underlying provider implementation does not match the expected type,
// it's to be considered abnormal execution, and the supplied `default value` should be returned.
func TestRequirements_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
		StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
		FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
		IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
		ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.3.* requirements interface")
	}
}

// The `client` MUST provide methods for detailed flag value evaluation with parameters `flag key` (string, required),
// `default value` (boolean | number | string | structure, required), `evaluation context` (optional),
// and `evaluation options` (optional), which returns an `evaluation details` structure.
func TestRequirement_1_4_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
		StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
		FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
		IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
		ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.4.1 requirements interface")
	}
}

// Requirement_1_4_2
// The `evaluation details` structure's `value` field MUST contain the evaluated flag value.
//
// Has no suitable test as the provider implementation populates the EvaluationDetails value field

// TODO Requirement_1_4_3 once upgraded Go to 1.18 for generics

// The `evaluation details` structure's `flag key` field MUST contain the `flag key`
// argument passed to the detailed flag evaluation method.
func TestRequirement_1_4_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	flagKey := "foo"

	t.Run("BooleanValueDetails", func(t *testing.T) {
		evDetails, err := client.BooleanValueDetails(context.Background(), flagKey, true, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})

	t.Run("StringValueDetails", func(t *testing.T) {
		evDetails, err := client.StringValueDetails(context.Background(), flagKey, "", EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})

	t.Run("FloatValueDetails", func(t *testing.T) {
		evDetails, err := client.FloatValueDetails(context.Background(), flagKey, 1, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})

	t.Run("IntValueDetails", func(t *testing.T) {
		evDetails, err := client.IntValueDetails(context.Background(), flagKey, 1, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})

	t.Run("ObjectValueDetails", func(t *testing.T) {
		evDetails, err := client.ObjectValueDetails(context.Background(), flagKey, 1, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})
}

// Requirement_1_4_5
// In cases of normal execution, the `evaluation details` structure's `variant` field MUST
// contain the value of the `variant` field in the `flag resolution` structure returned
// by the configured `provider`, if the field is set.
//
// Has no suitable test as the provider implementation populates the EvaluationDetails variant field

// Requirement_1_4_6
// In cases of normal execution, the `evaluation details` structure's `reason` field MUST
// contain the value of the `reason` field in the `flag resolution` structure returned
// by the configured `provider`, if the field is set.
//
// Has no suitable test as the provider implementation populates the EvaluationDetails reason field

// Requirement_1_4_7
// In cases of abnormal execution, the `evaluation details` structure's `error code` field MUST
// contain a string identifying an error occurred during flag evaluation and the nature of the error.
//
// Has no suitable test as the provider implementation populates the EvaluationDetails error code field

// Requirement_1_4_8
// In cases of abnormal execution (network failure, unhandled error, etc) the `reason` field
// in the `evaluation details` SHOULD indicate an error.
//
// Has no suitable test as the provider implementation populates the EvaluationDetails reason field

// Methods, functions, or operations on the client MUST NOT throw exceptions, or otherwise abnormally terminate.
// Flag evaluation calls must always return the `default value` in the event of abnormal execution.
// Exceptions include functions or methods for the purposes for configuration or setup.
//
// This test asserts that the flag evaluation calls return the default value in the event of abnormal execution.
// The MUST NOT abnormally terminate clause of this requirement is satisfied by the error included in the return
// signatures, as is idiomatic in Go. Errors aren't fatal, the operations won't terminate as a result of an error.
func TestRequirement_1_4_9(t *testing.T) {
	client := NewClient("test-client")
	flagKey := "flag-key"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	ctrl := gomock.NewController(t)

	t.Run("Boolean", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := true
		mockProvider.EXPECT().Metadata().AnyTimes()
		mockProvider.EXPECT().Hooks().AnyTimes()
		mockProvider.EXPECT().BooleanEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(BoolResolutionDetail{
				Value: false,
				ResolutionDetail: ResolutionDetail{
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.BooleanValue(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected BooleanValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from BooleanValue, got %v", value)
		}

		valueDetails, err := client.BooleanValueDetails(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected BooleanValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(bool) != defaultValue {
			t.Errorf("expected default value from BooleanValueDetails, got %v", value)
		}
	})

	t.Run("String", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := "default"
		mockProvider.EXPECT().Metadata().AnyTimes()
		mockProvider.EXPECT().Hooks().AnyTimes()
		mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(StringResolutionDetail{
				Value: "foo",
				ResolutionDetail: ResolutionDetail{
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.StringValue(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected StringValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from StringValue, got %v", value)
		}

		valueDetails, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected StringValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(string) != defaultValue {
			t.Errorf("expected default value from StringValueDetails, got %v", value)
		}
	})

	t.Run("Float", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := 3.14159
		mockProvider.EXPECT().Metadata().AnyTimes()
		mockProvider.EXPECT().Hooks().AnyTimes()
		mockProvider.EXPECT().FloatEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(FloatResolutionDetail{
				Value: 0,
				ResolutionDetail: ResolutionDetail{
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.FloatValue(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected FloatValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from FloatValue, got %v", value)
		}

		valueDetails, err := client.FloatValueDetails(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected FloatValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(float64) != defaultValue {
			t.Errorf("expected default value from FloatValueDetails, got %v", value)
		}
	})

	t.Run("Int", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		var defaultValue int64 = 3
		mockProvider.EXPECT().Metadata().AnyTimes()
		mockProvider.EXPECT().Hooks().AnyTimes()
		mockProvider.EXPECT().IntEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(IntResolutionDetail{
				Value: 0,
				ResolutionDetail: ResolutionDetail{
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.IntValue(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected IntValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from IntValue, got %v", value)
		}

		valueDetails, err := client.IntValueDetails(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected FloatValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(int64) != defaultValue {
			t.Errorf("expected default value from IntValueDetails, got %v", value)
		}
	})

	t.Run("Object", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		type obj struct {
			foo string
		}
		defaultValue := obj{foo: "bar"}
		mockProvider.EXPECT().Metadata().AnyTimes()
		mockProvider.EXPECT().Hooks().AnyTimes()
		mockProvider.EXPECT().ObjectEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(InterfaceResolutionDetail{
				ResolutionDetail: ResolutionDetail{
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.ObjectValue(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected ObjectValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from ObjectValue, got %v", value)
		}

		valueDetails, err := client.ObjectValueDetails(context.Background(), flagKey, defaultValue, evalCtx)
		if err == nil {
			t.Error("expected ObjectValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(obj) != defaultValue {
			t.Errorf("expected default value from ObjectValueDetails, got %v", value)
		}
	})
}

// TODO Requirement_1_4_10
// In the case of abnormal execution, the client SHOULD log an informative error message.

// Requirement_1_4_11
// The `client` SHOULD provide asynchronous or non-blocking mechanisms for flag evaluation.
//
// Satisfied by goroutines.

// Requirement_1_5_1
// The `evaluation options` structure's `hooks` field denotes an ordered collection of hooks that the client MUST
// execute for the respective flag evaluation, in addition to those already configured.
//
// Is tested by TestRequirement_4_4_2.

// TODO Requirement_1_6_1
// The `client` SHOULD transform the `evaluation context` using the `provider's` `context transformer` function
// if one is defined, before passing the result of the transformation to the provider's flag resolution functions.

func TestFlattenContext(t *testing.T) {
	tests := map[string]struct {
		inCtx  EvaluationContext
		outCtx map[string]interface{}
	}{
		"happy path": {
			inCtx: EvaluationContext{
				Attributes: map[string]interface{}{
					"1": "string",
					"2": 0.01,
					"3": false,
				},
				TargetingKey: "user",
			},
			outCtx: map[string]interface{}{
				TargetingKey: "user",
				"1":          "string",
				"2":          0.01,
				"3":          false,
			},
		},
		"no targeting key": {
			inCtx: EvaluationContext{
				Attributes: map[string]interface{}{
					"1": "string",
					"2": 0.01,
					"3": false,
				},
			},
			outCtx: map[string]interface{}{
				"1": "string",
				"2": 0.01,
				"3": false,
			},
		},
		"duplicated key": {
			inCtx: EvaluationContext{
				TargetingKey: "user",
				Attributes: map[string]interface{}{
					TargetingKey: "not user",
					"1":          "string",
					"2":          0.01,
					"3":          false,
				},
			},
			outCtx: map[string]interface{}{
				TargetingKey: "user",
				"1":          "string",
				"2":          0.01,
				"3":          false,
			},
		},
		"no attributes": {
			inCtx: EvaluationContext{
				TargetingKey: "user",
			},
			outCtx: map[string]interface{}{
				TargetingKey: "user",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			out := flattenContext(test.inCtx)
			if !reflect.DeepEqual(test.outCtx, out) {
				t.Fatalf(
					"%s, unexpected value received from flatten context, expected %v got %v",
					name,
					test.outCtx,
					out,
				)
			}
		})
	}
}

// TestBeforeHookNilContext asserts that when a Before hook returns a nil EvaluationContext it doesn't overwrite the
// existing EvaluationContext
func TestBeforeHookNilContext(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)

	hookNilContext := UnimplementedHook{}

	client := NewClient("test")
	attributes := map[string]interface{}{"should": "persist"}
	evalCtx := EvaluationContext{Attributes: attributes}
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), attributes)

	_, err := client.BooleanValue(
		context.Background(), "foo", false, evalCtx, WithHooks(hookNilContext),
	)
	if err != nil {
		t.Fatal(err)
	}
}

type lr struct {
	callback func()
	logger
}

func (l lr) Info(level int, msg string, keysAndValues ...interface{}) {
	l.callback()
}

func TestClientLoggerUsesLatestGlobalLogger(t *testing.T) {
	defer t.Cleanup(initSingleton)

	called := false
	l := lr{callback: func() {
		called = true
	}}

	client := NewClient("test")
	SetLogger(logr.New(l))
	_, err := client.BooleanValue(context.Background(), "foo", false, EvaluationContext{})
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Error("client didn't use the updated global logger")
	}
}

func TestErrorCodeFromProviderReturnedInEvaluationDetails(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	const errorCode = "TIMEOUT"

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: true,
			ResolutionDetail: ResolutionDetail{
				ErrorCode: errorCode,
				Reason:    "",
				Variant:   "",
			},
		})

	client := NewClient("test")
	evalDetails, err := client.evaluate(
		context.Background(), "foo", Boolean, true, EvaluationContext{}, EvaluationOptions{},
	)
	if err == nil {
		t.Error("expected err, got nil")
	}

	if evalDetails.ErrorCode != errorCode {
		t.Errorf(
			"expected evaluation details to contain error code '%s', got '%s'",
			errorCode, evalDetails.ErrorCode,
		)
	}
}
