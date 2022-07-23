package openfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestRequirement_1_2_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	client := GetClient("test-client")
	client.AddHooks(mockHook)
	client.AddHooks(mockHook, mockHook)

	if len(client.hooks) != 3 {
		t.Error("func client.AddHooks didn't append the list of hooks to the client's existing collection of hooks")
	}
}

func TestRequirement_1_2_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	clientName := "test-client"
	client := GetClient(clientName)

	if client.Metadata().Name() != clientName {
		t.Errorf("client name not initiated as expected, got %s, want %s", client.Metadata().Name(), clientName)
	}
}

// TestRequirements_1_3 tests all the 1.3.* requirements by asserting that the returned client matches the interface
// defined by the 1.3.* requirements
func TestRequirements_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := GetClient("test-client")

	type requirements interface {
		BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (bool, error)
		StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (string, error)
		NumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (float64, error)
		ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (interface{}, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by GetClient doesn't implement the 1.3.* requirements interface")
	}
}

func TestRequirement_1_4_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := GetClient("test-client")

	type requirements interface {
		BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		NumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by GetClient doesn't implement the 1.4.1 requirements interface")
	}
}

// Requirement_1_4_2 has no suitable test as the provider implementation populates the EvaluationDetails value field

// TODO Requirement_1_4_3 once upgraded Go to 1.18 for generics

func TestRequirement_1_4_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := GetClient("test-client")

	flagKey := "foo"

	t.Run("BooleanValueDetails", func(t *testing.T) {
		evDetails, err := client.BooleanValueDetails(flagKey, true, nil)
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
		evDetails, err := client.StringValueDetails(flagKey, "", nil)
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

	t.Run("NumberValueDetails", func(t *testing.T) {
		evDetails, err := client.NumberValueDetails(flagKey, 1, nil)
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
		evDetails, err := client.ObjectValueDetails(flagKey, 1, nil)
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

// Requirement_1_4_5 has no suitable test as the provider implementation populates the EvaluationDetails variant field

// Requirement_1_4_6 has no suitable test as the provider implementation populates the EvaluationDetails reason field

// Requirement_1_4_7 has no suitable test as the provider implementation populates the EvaluationDetails error code field

// Requirement_1_4_8 has no suitable test as the provider implementation populates the EvaluationDetails reason field

// TestRequirement_1_4_9 tests that the flag evaluation calls return the default value in the event of abnormal execution.
// The MUST NOT abnormally terminate clause of this requirement is satisfied by the error included in the return
// signatures, as is idiomatic in Go. Errors aren't fatal, the operations won't terminate as a result of an error.
func TestRequirement_1_4_9(t *testing.T) {
	client := GetClient("test-client")
	flagKey := "flag-key"

	ctrl := gomock.NewController(t)

	t.Run("Boolean", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := true
		mockProvider.EXPECT().BooleanEvaluation(flagKey, defaultValue, nil).Return(BoolResolutionDetail{
			Value: false,
			ResolutionDetail: ResolutionDetail{
				Value:     false,
				ErrorCode: "GENERAL",
				Reason:    "forced test error",
			},
		}).Times(2)
		SetProvider(mockProvider)

		value, err := client.BooleanValue(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected BooleanValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from BooleanValue, got %v", value)
		}

		valueDetails, err := client.BooleanValueDetails(flagKey, defaultValue, nil)
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
		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, nil).Return(StringResolutionDetail{
			Value: "foo",
			ResolutionDetail: ResolutionDetail{
				Value:     "foo",
				ErrorCode: "GENERAL",
				Reason:    "forced test error",
			},
		}).Times(2)
		SetProvider(mockProvider)

		value, err := client.StringValue(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected StringValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from StringValue, got %v", value)
		}

		valueDetails, err := client.StringValueDetails(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected StringValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(string) != defaultValue {
			t.Errorf("expected default value from StringValueDetails, got %v", value)
		}
	})

	t.Run("Number", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := 3.14159
		mockProvider.EXPECT().NumberEvaluation(flagKey, defaultValue, nil).Return(NumberResolutionDetail{
			Value: 0,
			ResolutionDetail: ResolutionDetail{
				Value:     0,
				ErrorCode: "GENERAL",
				Reason:    "forced test error",
			},
		}).Times(2)
		SetProvider(mockProvider)

		value, err := client.NumberValue(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected NumberValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from NumberValue, got %v", value)
		}

		valueDetails, err := client.NumberValueDetails(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected NumberValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(float64) != defaultValue {
			t.Errorf("expected default value from NumberValueDetails, got %v", value)
		}
	})

	t.Run("Object", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		type obj struct {
			foo string
		}
		defaultValue := obj{foo: "bar"}
		mockProvider.EXPECT().ObjectEvaluation(flagKey, defaultValue, nil).Return(ResolutionDetail{
			Value:     obj{foo: "foo"},
			ErrorCode: "GENERAL",
			Reason:    "forced test error",
		}).Times(2)
		SetProvider(mockProvider)

		value, err := client.ObjectValue(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected ObjectValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from ObjectValue, got %v", value)
		}

		valueDetails, err := client.ObjectValueDetails(flagKey, defaultValue, nil)
		if err == nil {
			t.Error("expected ObjectValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(obj) != defaultValue {
			t.Errorf("expected default value from ObjectValueDetails, got %v", value)
		}
	})
}

// TODO Requirement_1_4_10

// Requirement_1_4_11 is satisfied by goroutines.

// TODO Requirement_1_5_1

// TODO Requirement_1_6_1
