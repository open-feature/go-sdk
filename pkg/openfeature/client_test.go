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
		BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
		StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
		FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
		IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
		ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.4.1 requirements interface")
	}
}

// Requirement_1_4_2
// The `evaluation details` structure's `value` field MUST contain the evaluated flag value.

// Requirement_1_4_5
// In cases of normal execution, the `evaluation details` structure's `variant` field MUST
// contain the value of the `variant` field in the `flag resolution` structure returned
// by the configured `provider`, if the field is set.

// Requirement_1_4_6
// In cases of normal execution, the `evaluation details` structure's `reason` field MUST
// contain the value of the `reason` field in the `flag resolution` structure returned
// by the configured `provider`, if the field is set.
func TestRequirement_1_4_2__1_4_5__1_4_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")
	const (
		booleanValue = true
		stringValue  = "str"
		intValue     = 10
		floatValue   = 0.1

		booleanVariant = "boolean"
		stringVariant  = "string"
		intVariant     = "ten"
		floatVariant   = "tenth"
		objectVariant  = "object"

		testReason = "TEST_REASON"

		incorrectValue   = "Incorrect value returned!"
		incorrectVariant = "Incorrect variant returned!"
		incorrectReason  = "Incorrect reason returned!"
	)
	var objectValue = map[string]interface{}{"foo": 1, "bar": true, "baz": "buz"}

	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)

	flagKey := "foo"

	t.Run("BooleanValueDetails", func(t *testing.T) {
		mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(BoolResolutionDetail{
				Value: booleanValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: booleanVariant,
					Reason:  testReason,
				},
			})

		evDetails, err := client.BooleanValueDetails(context.Background(), flagKey, false, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.Value != booleanValue {
			t.Error(err)
		}
	})

	t.Run("StringValueDetails", func(t *testing.T) {
		mockProvider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(StringResolutionDetail{
				Value: stringValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: stringVariant,
					Reason:  testReason,
				},
			})

		evDetails, err := client.StringValueDetails(context.Background(), flagKey, "", EvaluationContext{})

		if err != nil {
			t.Error(err)
		}
		if evDetails.Value != stringValue {
			t.Error(incorrectValue)
		}
		if evDetails.Variant != stringVariant {
			t.Error(incorrectVariant)
		}
		if evDetails.Reason != testReason {
			t.Error(incorrectReason)
		}
	})

	t.Run("FloatValueDetails", func(t *testing.T) {
		mockProvider.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(FloatResolutionDetail{
				Value: floatValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: floatVariant,
					Reason:  testReason,
				},
			})

		evDetails, err := client.FloatValueDetails(context.Background(), flagKey, 0, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.Value != floatValue {
			t.Error(incorrectValue)
		}
		if evDetails.Variant != floatVariant {
			t.Error(incorrectVariant)
		}
		if evDetails.Reason != testReason {
			t.Error(incorrectReason)
		}
	})

	t.Run("IntValueDetails", func(t *testing.T) {
		mockProvider.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(IntResolutionDetail{
				Value: intValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: intVariant,
					Reason:  testReason,
				},
			})

		evDetails, err := client.IntValueDetails(context.Background(), flagKey, 0, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if evDetails.Value != intValue {
			t.Error(incorrectValue)
		}
		if evDetails.Variant != intVariant {
			t.Error(incorrectVariant)
		}
		if evDetails.Reason != testReason {
			t.Error(incorrectReason)
		}
	})

	t.Run("ObjectValueDetails", func(t *testing.T) {
		mockProvider.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(InterfaceResolutionDetail{
				Value: objectValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: objectVariant,
					Reason:  testReason,
				},
			})

		evDetails, err := client.ObjectValueDetails(context.Background(), flagKey, nil, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(evDetails.Value, objectValue) {
			t.Error(incorrectValue)
		}
		if evDetails.Variant != objectVariant {
			t.Error(incorrectVariant)
		}
		if evDetails.Reason != testReason {
			t.Error(incorrectReason)
		}
	})
}

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

// In cases of abnormal execution, the `evaluation details` structure's
// `error code` field MUST contain an `error code`.
func TestRequirement_1_4_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: false,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})
	SetProvider(mockProvider)

	res, err := client.evaluate(
		context.Background(), "foo", Boolean, true, EvaluationContext{}, EvaluationOptions{},
	)
	if err == nil {
		t.Error("expected err, got nil")
	}

	expectedErrorCode := GeneralCode
	if res.ErrorCode != expectedErrorCode {
		t.Errorf("expected error code to be '%s', got '%s'", expectedErrorCode, res.ErrorCode)
	}
}

// In cases of abnormal execution (network failure, unhandled error, etc) the `reason` field
// in the `evaluation details` SHOULD indicate an error.
func TestRequirement_1_4_8(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: false,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})
	SetProvider(mockProvider)

	res, err := client.evaluate(
		context.Background(), "foo", Boolean, true, EvaluationContext{}, EvaluationOptions{},
	)
	if err == nil {
		t.Error("expected err, got nil")
	}

	expectedReason := ErrorReason
	if res.Reason != expectedReason {
		t.Errorf("expected reason to be '%s', got '%s'", expectedReason, res.Reason)
	}
}

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
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
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

		if valueDetails.Value != defaultValue {
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
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
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

		if valueDetails.Value != defaultValue {
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
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
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

		if valueDetails.Value != defaultValue {
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
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
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

		if valueDetails.Value != defaultValue {
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
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
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

// In cases of abnormal execution, the `evaluation details` structure's `error message` field MAY contain a
// string containing additional details about the nature of the error.
func TestRequirement_1_4_12(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	errMessage := "error forced by test"

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: true,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError(errMessage),
			},
		})

	client := NewClient("test")
	evalDetails, err := client.evaluate(
		context.Background(), "foo", Boolean, true, EvaluationContext{}, EvaluationOptions{},
	)
	if err == nil {
		t.Error("expected err, got nil")
	}

	if evalDetails.ErrorMessage != errMessage {
		t.Errorf(
			"expected evaluation details to contain error message '%s', got '%s'",
			errMessage, evalDetails.ErrorMessage,
		)
	}
}

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
		outCtx FlattenedContext
	}{
		"happy path": {
			inCtx: EvaluationContext{
				attributes: map[string]interface{}{
					"1": "string",
					"2": 0.01,
					"3": false,
				},
				targetingKey: "user",
			},
			outCtx: FlattenedContext{
				TargetingKey: "user",
				"1":          "string",
				"2":          0.01,
				"3":          false,
			},
		},
		"no targeting key": {
			inCtx: EvaluationContext{
				attributes: map[string]interface{}{
					"1": "string",
					"2": 0.01,
					"3": false,
				},
			},
			outCtx: FlattenedContext{
				"1": "string",
				"2": 0.01,
				"3": false,
			},
		},
		"duplicated key": {
			inCtx: EvaluationContext{
				targetingKey: "user",
				attributes: map[string]interface{}{
					TargetingKey: "not user",
					"1":          "string",
					"2":          0.01,
					"3":          false,
				},
			},
			outCtx: FlattenedContext{
				TargetingKey: "user",
				"1":          "string",
				"2":          0.01,
				"3":          false,
			},
		},
		"no attributes": {
			inCtx: EvaluationContext{
				targetingKey: "user",
			},
			outCtx: FlattenedContext{
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
	evalCtx := EvaluationContext{attributes: attributes}
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

	generalErrorCode := GeneralCode

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)
	mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: true,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})

	client := NewClient("test")
	evalDetails, err := client.evaluate(
		context.Background(), "foo", Boolean, true, EvaluationContext{}, EvaluationOptions{},
	)
	if err == nil {
		t.Error("expected err, got nil")
	}

	if evalDetails.ErrorCode != generalErrorCode {
		t.Errorf(
			"expected evaluation details to contain error code '%s', got '%s'",
			generalErrorCode, evalDetails.ErrorCode,
		)
	}
}

func TestSwitchingProvidersMidEvaluationCausesNoImpactToEvaluation(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider1 := NewMockFeatureProvider(ctrl)
	mockProvider2 := NewMockFeatureProvider(ctrl)
	mockProvider1Hook := NewMockHook(ctrl)
	mockProvider1.EXPECT().Metadata().AnyTimes()
	mockProvider2.EXPECT().Metadata().AnyTimes()
	mockProvider1.EXPECT().Hooks().Return([]Hook{mockProvider1Hook}).AnyTimes()

	// set new provider during initial provider's Before hook
	mockProvider1Hook.EXPECT().Before(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ HookContext, _ HookHints) (*EvaluationContext, error) {
			SetProvider(mockProvider2)
			return nil, nil
		})
	SetProvider(mockProvider1)

	mockProvider1.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	// ensure that the first provider's hooks are still fired
	mockProvider1Hook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockProvider1Hook.EXPECT().Finally(gomock.Any(), gomock.Any())

	client := NewClient("test")
	_, err := client.BooleanValue(context.Background(), "foo", true, EvaluationContext{})
	if err != nil {
		t.Error(err)
	}
}

func TestClientProviderLock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer t.Cleanup(initSingleton)

	mockMutex := NewMockmutex(ctrl)
	mockProvider1 := NewMockFeatureProvider(ctrl)
	mockProvider2 := NewMockFeatureProvider(ctrl)

	mockProvider1.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	mockProvider1.EXPECT().Hooks().AnyTimes()
	mockProvider1.EXPECT().Metadata().AnyTimes()

	mockMutex.EXPECT().RLock().DoAndReturn(func() {
		api.prvder = mockProvider1
	})

	// test that any provider change after RUnlock has no impact on this transaction
	mockMutex.EXPECT().RUnlock().DoAndReturn(func() {
		api.prvder = mockProvider2
	})
	api.mutex = mockMutex

	client := NewClient("test").WithLogger(logr.New(logger{}))
	_, err := client.BooleanValue(context.Background(), "foo", false, EvaluationContext{})
	if err != nil {
		t.Error(err)
	}
}

func TestObjectEvaluationShouldSupportNilValue(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	variant := "variant1"
	reason := TargetingMatchReason
	var value interface{} = nil

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	SetProvider(mockProvider)
	mockProvider.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(InterfaceResolutionDetail{
			Value: value,
			ProviderResolutionDetail: ProviderResolutionDetail{
				Variant: variant,
				Reason:  reason,
			},
		})

	client := NewClient("test")
	evDetails, err := client.ObjectValueDetails(context.Background(), "foo", nil, EvaluationContext{})
	if err != nil {
		t.Errorf("should not return an error: %s", err.Error())
	}
	if evDetails.Value != value {
		t.Errorf("unexpected value returned (expected: %s, value: %s)", value, evDetails.Value)
	}
	if evDetails.Variant != variant {
		t.Errorf("unexpected variant returned (expected: %s, value: %s)", variant, evDetails.Variant)
	}
	if evDetails.Reason != reason {
		t.Errorf("unexpected reason returned (expected: %s, value: %s)", reason, evDetails.Reason)
	}
	if evDetails.ErrorMessage != "" {
		t.Error("not supposed to have an error message")
	}
	if evDetails.ErrorCode != "" {
		t.Error("not supposed to have an error code")
	}
}
