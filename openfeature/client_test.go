package openfeature

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

type clientMocks struct {
	clientHandlerAPI *MockclientEvent
	evaluationAPI    *MockevaluationImpl
	providerAPI      *MockFeatureProvider
}

func hydratedMocksForClientTests(t *testing.T, expectedEvaluations int) clientMocks {
	ctrl := gomock.NewController(t)
	mockClientApi := NewMockclientEvent(ctrl)
	mockEvaluationApi := NewMockevaluationImpl(ctrl)
	mockProvider := NewMockFeatureProvider(ctrl)

	mockClientApi.EXPECT().State(gomock.Any()).AnyTimes().Return(ReadyState)

	mockProvider.EXPECT().Metadata().AnyTimes()
	mockProvider.EXPECT().Hooks().AnyTimes()
	mockEvaluationApi.EXPECT().ForEvaluation(gomock.Any()).Times(expectedEvaluations).DoAndReturn(func(_ string) (*MockFeatureProvider, []Hook, EvaluationContext) {
		return mockProvider, nil, EvaluationContext{}
	})

	return clientMocks{
		clientHandlerAPI: mockClientApi,
		evaluationAPI:    mockEvaluationApi,
		providerAPI:      mockProvider,
	}
}

// The client MUST provide a method to add `hooks` which accepts one or more API-conformant `hooks`,
// and appends them to the collection of any previously added hooks.
// When new hooks are added, previously added hooks are not removed.
func TestRequirement_1_2_1(t *testing.T) {
	t.Cleanup(initSingleton)
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
// containing an immutable `domain` field or accessor of type string,
// which corresponds to the `domain` value supplied during client creation.
func TestRequirement_1_2_2(t *testing.T) {
	t.Cleanup(initSingleton)
	clientName := "test-client"

	client := NewClient(clientName)

	if client.Metadata().Domain() != clientName {
		t.Errorf("client domain not initiated as expected, got %s, want %s", client.Metadata().Domain(), clientName)
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
	t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
		StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
		FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
		IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
		ObjectValue(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (any, error)
	}

	var clientI any = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.3.* requirements interface")
	}
}

// The `client` MUST provide methods for detailed flag value evaluation with parameters `flag key` (string, required),
// `default value` (boolean | number | string | structure, required), `evaluation context` (optional),
// and `evaluation options` (optional), which returns an `evaluation details` structure.
func TestRequirement_1_4_1(t *testing.T) {
	t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
		StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
		FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
		IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
		ObjectValueDetails(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)
	}

	var clientI any = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.4.1 requirements interface")
	}
}

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

var objectValue = map[string]any{"foo": 1, "bar": true, "baz": "buz"}

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
	t.Cleanup(initSingleton)

	flagKey := "foo"

	t.Run("BooleanValueDetails", func(t *testing.T) {
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
	t.Cleanup(initSingleton)
	flagKey := "foo"
	t.Run("BooleanValueDetails", func(t *testing.T) {
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(BoolResolutionDetail{
				Value: booleanValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: booleanVariant,
					Reason:  testReason,
				},
			})
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
		if evDetails.FlagKey != flagKey {
			t.Errorf(
				"flag key isn't as expected in EvaluationDetail, got %s, expected %s",
				evDetails.FlagKey, flagKey,
			)
		}
	})

	t.Run("FloatValueDetails", func(t *testing.T) {
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(FloatResolutionDetail{
				Value: floatValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: floatVariant,
					Reason:  testReason,
				},
			})
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(IntResolutionDetail{
				Value: intValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: intVariant,
					Reason:  testReason,
				},
			})
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
		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		mocks.providerAPI.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(InterfaceResolutionDetail{
				Value: objectValue,
				ProviderResolutionDetail: ProviderResolutionDetail{
					Variant: objectVariant,
					Reason:  testReason,
				},
			})
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
	t.Cleanup(initSingleton)
	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

	mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: false,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})

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
	t.Cleanup(initSingleton)
	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
	mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: false,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})

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
	flagKey := "flag-key"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("Boolean", func(t *testing.T) {
		t.Cleanup(initSingleton)

		mocks := hydratedMocksForClientTests(t, 2)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

		defaultValue := true
		mocks.providerAPI.EXPECT().BooleanEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
				},
			}).Times(2)

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
		t.Cleanup(initSingleton)

		mocks := hydratedMocksForClientTests(t, 2)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

		defaultValue := "default"
		mocks.providerAPI.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(StringResolutionDetail{
				Value: "foo",
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
				},
			}).Times(2)

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
		t.Cleanup(initSingleton)
		mocks := hydratedMocksForClientTests(t, 2)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

		defaultValue := 3.14159
		mocks.providerAPI.EXPECT().FloatEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(FloatResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
				},
			}).Times(2)

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
		t.Cleanup(initSingleton)
		mocks := hydratedMocksForClientTests(t, 2)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		var defaultValue int64 = 3
		mocks.providerAPI.EXPECT().IntEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(IntResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
				},
			}).Times(2)

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
		t.Cleanup(initSingleton)
		mocks := hydratedMocksForClientTests(t, 2)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		type obj struct {
			foo string
		}
		defaultValue := obj{foo: "bar"}
		mocks.providerAPI.EXPECT().ObjectEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(InterfaceResolutionDetail{
				ProviderResolutionDetail: ProviderResolutionDetail{
					ResolutionError: NewGeneralResolutionError("test"),
				},
			}).Times(2)
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
	t.Cleanup(initSingleton)

	errMessage := "error forced by test"

	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
	mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: true,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError(errMessage),
			},
		})
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

// Requirement_1_4_13
// If the `flag metadata` field in the `flag resolution` structure returned by the configured `provider` is set,
// the `evaluation details` structure's `flag metadata` field MUST contain that value. Otherwise,
// it MUST contain an empty record.
func TestRequirement_1_4_13(t *testing.T) {
	flagKey := "flag-key"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("No Metadata", func(t *testing.T) {
		t.Cleanup(initSingleton)

		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		defaultValue := true
		mocks.providerAPI.EXPECT().BooleanEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: ProviderResolutionDetail{
					FlagMetadata: nil,
				},
			}).Times(1)

		evDetails, err := client.BooleanValueDetails(context.Background(), flagKey, defaultValue, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(evDetails.FlagMetadata, FlagMetadata{}) {
			t.Errorf(
				"flag metadata is not as expected in EvaluationDetail, got %v, expected %v",
				evDetails.FlagMetadata, FlagMetadata{},
			)
		}
	})

	t.Run("Metadata present", func(t *testing.T) {
		t.Cleanup(initSingleton)

		mocks := hydratedMocksForClientTests(t, 1)
		client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
		defaultValue := true
		metadata := FlagMetadata{
			"bing": "bong",
		}
		mocks.providerAPI.EXPECT().BooleanEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
			Return(BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: ProviderResolutionDetail{
					FlagMetadata: metadata,
				},
			}).Times(1)

		evDetails, err := client.BooleanValueDetails(context.Background(), flagKey, defaultValue, EvaluationContext{})
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(metadata, evDetails.FlagMetadata) {
			t.Errorf(
				"flag metadata is not as expected in EvaluationDetail, got %v, expected %v",
				evDetails.FlagMetadata, metadata,
			)
		}
	})
}

// Requirement_1_5_1
// The `evaluation options` structure's `hooks` field denotes an ordered collection of hooks that the client MUST
// execute for the respective flag evaluation, in addition to those already configured.
//
// Is tested by TestRequirement_4_4_2.

// TODO Requirement_1_6_1
// The `client` SHOULD transform the `evaluation context` using the `provider's` `context transformer` function
// if one is defined, before passing the result of the transformation to the provider's flag resolution functions.

// TestRequirement_6_1 tests the 6.1.1 and 6.1.2 requirements by asserting that the returned client matches the interface
// defined by the 6.1.1 and 6.1.2 requirements

// Requirement_6_1_1
// The `client` MUST define a function for tracking the occurrence of a particular action or application state,
// with parameters `tracking event name` (string, required), `evaluation context` (optional) and `tracking event details` (optional),
// which returns nothing.

// Requirement_6_1_2
// The `client` MUST define a function for tracking the occurrence of a particular action or application state,
// with parameters `tracking event name` (string, required) and `tracking event details` (optional), which returns nothing.
func TestRequirement_6_1(t *testing.T) {
	client := NewClient("test-client")

	type requirements interface {
		Track(ctx context.Context, trackingEventName string, evalCtx EvaluationContext, details TrackingEventDetails)
	}

	var clientI any = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.6.* requirements interface")
	}
}

// Requirement_6_1_3
// The evaluation context passed to the provider's track function MUST be merged in the order, with duplicate values being overwritten:
// - API (global; lowest precedence)
// - transaction
// - client
// - invocation (highest precedence)

// Requirement_6_1_4
// If the client's `track` function is called and the associated provider does not implement tracking, the client's `track` function MUST no-op.
// Allow backward compatible to non-Tracker Provider
func TestTrack(t *testing.T) {
	type inputCtx struct {
		api        EvaluationContext
		txn        EvaluationContext
		client     EvaluationContext
		invocation EvaluationContext
	}

	// mockTrackingProvider is a feature provider that implements tracker contract.
	type mockTrackingProvider struct {
		*MockTracker
		*MockFeatureProvider
	}

	type testcase struct {
		inCtx     inputCtx
		eventName string
		outCtx    EvaluationContext
		// allow asserting the input to provider
		provider func(tc *testcase, provider *MockFeatureProvider) FeatureProvider
	}

	tests := map[string]*testcase{
		"merging in correct order": {
			eventName: "example-event",
			inCtx: inputCtx{
				api: EvaluationContext{
					attributes: map[string]any{
						"1": "api",
						"2": "api",
						"3": "api",
						"4": "api",
					},
				},
				txn: EvaluationContext{
					attributes: map[string]any{
						"2": "txn",
						"3": "txn",
						"4": "txn",
					},
				},
				client: EvaluationContext{
					attributes: map[string]any{
						"3": "client",
						"4": "client",
					},
				},
				invocation: EvaluationContext{
					attributes: map[string]any{
						"4": "invocation",
					},
				},
			},
			outCtx: EvaluationContext{
				attributes: map[string]any{
					"1": "api",
					"2": "txn",
					"3": "client",
					"4": "invocation",
				},
			},
			provider: func(tc *testcase, mockProvider *MockFeatureProvider) FeatureProvider {
				provider := &mockTrackingProvider{
					MockTracker:         NewMockTracker(mockProvider.ctrl),
					MockFeatureProvider: mockProvider,
				}
				// assert AnyTimesif Track is called once with evalCtx expected
				provider.MockTracker.EXPECT().Track(gomock.Any(), gomock.Any(), tc.outCtx, TrackingEventDetails{}).Times(1)

				return provider
			},
		},
		"do no-op if Provider do not implement Tracker": {
			inCtx:     inputCtx{},
			eventName: "example-event",
			outCtx:    EvaluationContext{},
			provider: func(tc *testcase, provider *MockFeatureProvider) FeatureProvider {
				return provider
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			mocks := hydratedMocksForClientTests(t, 0)
			client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

			provider := test.provider(test, mocks.providerAPI)

			mocks.evaluationAPI.EXPECT().ForEvaluation("test-client").AnyTimes().DoAndReturn(func(_ string) (FeatureProvider, []Hook, EvaluationContext) {
				return provider, nil, test.inCtx.api
			})
			client.evaluationContext = test.inCtx.client
			ctx := WithTransactionContext(context.Background(), test.inCtx.txn)

			// action
			client.Track(ctx, test.eventName, test.inCtx.invocation, TrackingEventDetails{})
		})
	}
}

func TestFlattenContext(t *testing.T) {
	tests := map[string]struct {
		inCtx  EvaluationContext
		outCtx FlattenedContext
	}{
		"happy path": {
			inCtx: EvaluationContext{
				attributes: map[string]any{
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
				attributes: map[string]any{
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
				attributes: map[string]any{
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
				t.Errorf(
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
	t.Cleanup(initSingleton)

	hookNilContext := UnimplementedHook{}

	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)

	attributes := map[string]any{"should": "persist"}
	evalCtx := EvaluationContext{attributes: attributes}
	mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), attributes)

	_, err := client.BooleanValue(
		context.Background(), "foo", false, evalCtx, WithHooks(hookNilContext),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestErrorCodeFromProviderReturnedInEvaluationDetails(t *testing.T) {
	t.Cleanup(initSingleton)

	generalErrorCode := GeneralCode

	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
	mocks.providerAPI.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(BoolResolutionDetail{
			Value: true,
			ProviderResolutionDetail: ProviderResolutionDetail{
				ResolutionError: NewGeneralResolutionError("test"),
			},
		})

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

func TestObjectEvaluationShouldSupportNilValue(t *testing.T) {
	t.Cleanup(initSingleton)

	variant := "variant1"
	reason := TargetingMatchReason
	var value any = nil

	mocks := hydratedMocksForClientTests(t, 1)
	client := newClient("test-client", mocks.evaluationAPI, mocks.clientHandlerAPI)
	mocks.providerAPI.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(InterfaceResolutionDetail{
			Value: value,
			ProviderResolutionDetail: ProviderResolutionDetail{
				Variant: variant,
				Reason:  reason,
			},
		})

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

func TestFlagMetadataAccessors(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		expectedValue := true
		key := "bool"
		key2 := "not-bool"
		metadata := FlagMetadata{
			key:  expectedValue,
			key2: "12",
		}
		val, err := metadata.GetBool(key)
		if err != nil {
			t.Error("unexpected error value, expected nil", err)
		}
		if val != expectedValue {
			t.Errorf("wrong value returned from FlagMetadata, expected %t, got %t", val, expectedValue)
		}
		_, err = metadata.GetBool(key2)
		if err == nil {
			t.Error("unexpected error value", err)
		}
		_, err = metadata.GetBool("not-in-map")
		if err == nil {
			t.Error("unexpected error value", err)
		}
	})

	t.Run("string", func(t *testing.T) {
		expectedValue := "string"
		key := "string"
		key2 := "not-string"
		metadata := FlagMetadata{
			key:  expectedValue,
			key2: true,
		}
		val, err := metadata.GetString(key)
		if err != nil {
			t.Error("unexpected error value, expected nil", err)
		}
		if val != expectedValue {
			t.Errorf("wrong value returned from FlagMetadata, expected %s, got %s", val, expectedValue)
		}
		_, err = metadata.GetString(key2)
		if err == nil {
			t.Error("unexpected error value", err)
		}
		_, err = metadata.GetString("not-in-map")
		if err == nil {
			t.Error("unexpected error value", err)
		}
	})

	t.Run("int", func(t *testing.T) {
		expectedValue := int64(12)
		metadata := FlagMetadata{
			"int":    int(12),
			"int8":   int8(12),
			"int16":  int16(12),
			"int32":  int32(12),
			"int164": int32(12),
		}
		for k := range metadata {
			val, err := metadata.GetInt(k)
			if err != nil {
				t.Error("unexpected error value, expected nil", err)
			}
			if val != expectedValue {
				t.Errorf("wrong value returned from FlagMetadata, expected %b, got %b", val, expectedValue)
			}
		}

		metadata = FlagMetadata{
			"not-int": true,
		}
		_, err := metadata.GetInt("not-int")
		if err == nil {
			t.Error("unexpected error value", err)
		}
		_, err = metadata.GetInt("not-in-map")
		if err == nil {
			t.Error("unexpected error value", err)
		}
	})

	t.Run("float", func(t *testing.T) {
		expectedValue := float64(12)
		metadata := FlagMetadata{
			"float32": float32(12),
			"float64": float64(12),
		}
		for k := range metadata {
			val, err := metadata.GetFloat(k)
			if err != nil {
				t.Error("unexpected error value, expected nil", err)
			}
			if val != expectedValue {
				t.Errorf("wrong value returned from FlagMetadata, expected %b, got %b", val, expectedValue)
			}
		}

		metadata = FlagMetadata{
			"not-float": true,
		}
		_, err := metadata.GetInt("not-float")
		if err == nil {
			t.Error("unexpected error value", err)
		}
		_, err = metadata.GetInt("not-in-map")
		if err == nil {
			t.Error("unexpected error value", err)
		}
	})
}

// The client MUST define a provider status accessor which indicates the readiness of the associated provider.
// with possible values NOT_READY, READY, STALE, ERROR, or FATAL.
func TestRequirement_1_7_1(t *testing.T) {
	client := NewClient("test-client")

	type requirements interface {
		State() State
	}

	var clientI any = client
	if _, ok := clientI.(requirements); !ok {
		t.Fatal("client does not define a status accessor which indicates the readiness of the associated provider")
	}

	TestRequirement_5_3_5(t)
}

// The client's provider status accessor MUST indicate READY if the initialize function of the associated provider
// terminates normally.
func TestRequirement_1_7_2(t *testing.T) {
	t.Cleanup(initSingleton)

	if GetApiInstance().GetNamedClient(t.Name()).State() != NotReadyState {
		t.Fatalf("expected client to report NOT READY state")
	}

	provider := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		&ProviderEventing{
			c: make(chan Event, 1),
		},
	}

	if err := SetNamedProviderAndWait(t.Name(), provider); err != nil {
		t.Fatalf("failed to set up provider: %v", err)
	}

	if GetApiInstance().GetNamedClient(t.Name()).State() != ReadyState {
		t.Fatalf("expected client to report READY state")
	}
}

// The client's provider status accessor MUST indicate ERROR if the initialize function of the associated provider
// terminates abnormally
func TestRequirement_1_7_3(t *testing.T) {
	t.Cleanup(initSingleton)
	provider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(e EvaluationContext) error {
				return errors.New("whoops... error from initialization")
			},
		},
		&ProviderEventing{},
	}

	_ = SetNamedProviderAndWait(t.Name(), provider)
	if GetApiInstance().GetNamedClient(t.Name()).State() != ErrorState {
		t.Fatalf("expected client to report ERROR state")
	}
}

// The client's provider status accessor MUST indicate FATAL if the initialize function of the associated provider
// terminates abnormally and indicates error code PROVIDER_FATAL.
func TestRequirement_1_7_4(t *testing.T) {
	t.Cleanup(initSingleton)
	provider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(e EvaluationContext) error {
				return errors.New("whoops... error from initialization")
			},
		},
		&ProviderEventing{},
	}

	_ = SetNamedProviderAndWait(t.Name(), provider)

	if GetApiInstance().GetNamedClient(t.Name()).State() != ErrorState {
		t.Fatalf("expected client to report ERROR state")
	}
}

// The client's provider status accessor MUST indicate FATAL if the initialize function of the associated provider
// terminates abnormally and indicates error code PROVIDER_FATAL.
func TestRequirement_1_7_5(t *testing.T) {
	t.Cleanup(initSingleton)
	provider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(e EvaluationContext) error {
				return &ProviderInitError{ErrorCode: ProviderFatalCode}
			},
		},
		&ProviderEventing{},
	}

	_ = SetNamedProviderAndWait(t.Name(), provider)

	if GetApiInstance().GetNamedClient(t.Name()).State() != FatalState {
		t.Fatalf("expected client to report ERROR state")
	}
}

// The client MUST default, run error hooks, and indicate an error if flag resolution is attempted while the provider
// is in NOT_READY.
func TestRequirement_1_7_6(t *testing.T) {
	t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	mockHook := NewMockHook(ctrl)
	mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), ProviderNotReadyError, gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	notReadyEventingProvider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(e EvaluationContext) error {
				<-time.After(math.MaxInt)
				return nil
			},
		},
		&ProviderEventing{},
	}

	_ = GetApiInstance().SetProvider(notReadyEventingProvider)

	client := GetApiInstance().GetNamedClient("somOtherClient")
	client.AddHooks(mockHook)

	if client.State() != NotReadyState {
		t.Fatalf("expected client to report NOT READY state")
	}

	defaultVal := true
	res, err := client.BooleanValue(context.Background(), "a-flag", defaultVal, EvaluationContext{})
	if err == nil {
		t.Fatalf("expected client to report an error")
	}

	if res != defaultVal {
		t.Fatalf("expected resolved boolean value to default to %t, got %t", defaultVal, res)
	}
}

// The client MUST default, run error hooks, and indicate an error if flag resolution is attempted while the provider
// is in FATAL.
func TestRequirement_1_7_7(t *testing.T) {
	t.Cleanup(initSingleton)
	provider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(e EvaluationContext) error {
				return &ProviderInitError{ErrorCode: ProviderFatalCode}
			},
		},
		&ProviderEventing{},
	}

	err := SetNamedProviderAndWait(t.Name(), provider)
	if err == nil {
		t.Errorf("provider registration was expected to fail but succeeded unexpectedly")
	}

	ctrl := gomock.NewController(t)
	mockHook := NewMockHook(ctrl)
	mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), ProviderFatalError, gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	client := GetApiInstance().GetNamedClient(t.Name())
	client.AddHooks(mockHook)

	if client.State() != FatalState {
		t.Fatalf("expected client to report FATAL state")
	}

	defaultVal := true
	res, err := client.BooleanValue(context.Background(), "a-flag", defaultVal, EvaluationContext{})
	if err == nil {
		t.Fatalf("expected client to report an error")
	}

	if res != defaultVal {
		t.Fatalf("expected resolved boolean value to default to %t, got %t", defaultVal, res)
	}
}

// Implementations SHOULD propagate the error code returned from any provider lifecycle methods.
func TestRequirement_1_7_8(t *testing.T) {
	t.Skip("Test not yet implemented")
}

// PROVIDER_ERROR events SHOULD populate the provider event details's error code field.
func TestRequirement_5_1_5(t *testing.T) {
	t.Skip("Test not yet implemented")
}

// If the provider emits an event, the value of the client's provider status MUST be updated accordingly.
func TestRequirement_5_3_5(t *testing.T) {
	t.Cleanup(initSingleton)

	eventually(t, func() bool {
		return GetApiInstance().GetClient().State() == NotReadyState
	}, time.Second, 100*time.Millisecond, "expected client to report NOT READY state")

	eventing := &ProviderEventing{
		c: make(chan Event, 1),
	}

	provider := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		eventing,
	}

	if err := SetNamedProviderAndWait(t.Name(), provider); err != nil {
		t.Fatalf("failed to set up provider: %v", err)
	}

	eventually(t, func() bool {
		return GetApiInstance().GetNamedClient(t.Name()).State() == ReadyState
	}, time.Second, 100*time.Millisecond, "expected client to report READY state")

	eventing.Invoke(Event{EventType: ProviderStale})
	eventually(t, func() bool {
		return GetApiInstance().GetNamedClient(t.Name()).State() == StaleState
	}, time.Second, 100*time.Millisecond, "expected client to report STALE state")

	eventing.Invoke(Event{EventType: ProviderError})
	eventually(t, func() bool {
		return GetApiInstance().GetNamedClient(t.Name()).State() == ErrorState
	}, time.Second, 100*time.Millisecond, "expected client to report ERROR state")

	eventing.Invoke(Event{EventType: ProviderReady})
	eventually(t, func() bool {
		return GetApiInstance().GetNamedClient(t.Name()).State() == ReadyState
	}, time.Second, 100*time.Millisecond, "expected client to report READY state")

	eventing.Invoke(Event{EventType: ProviderError, ProviderEventDetails: ProviderEventDetails{ErrorCode: ProviderFatalCode}})
	eventually(t, func() bool {
		return GetApiInstance().GetNamedClient(t.Name()).State() == FatalState
	}, time.Second, 100*time.Millisecond, "expected client to report FATAL state")
}
