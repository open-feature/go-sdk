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
		GetBooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error)
		GetStringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error)
		GetNumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error)
		GetObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error)
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
		GetBooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		GetStringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		GetNumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		GetObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
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

	t.Run("GetBooleanValueDetails", func(t *testing.T) {
		evDetails, err := client.GetBooleanValueDetails(flagKey, true, EvaluationContext{}, EvaluationOptions{})
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

	t.Run("GetStringValueDetails", func(t *testing.T) {
		evDetails, err := client.GetStringValueDetails(flagKey, "", EvaluationContext{}, EvaluationOptions{})
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

	t.Run("GetNumberValueDetails", func(t *testing.T) {
		evDetails, err := client.GetNumberValueDetails(flagKey, 1, EvaluationContext{}, EvaluationOptions{})
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

	t.Run("GetObjectValueDetails", func(t *testing.T) {
		evDetails, err := client.GetObjectValueDetails(flagKey, 1, EvaluationContext{}, EvaluationOptions{})
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
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().GetBooleanEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(BoolResolutionDetail{
				Value: false,
				ResolutionDetail: ResolutionDetail{
					Value:     false,
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.GetBooleanValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetBooleanValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from GetBooleanValue, got %v", value)
		}

		valueDetails, err := client.GetBooleanValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetBooleanValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(bool) != defaultValue {
			t.Errorf("expected default value from GetBooleanValueDetails, got %v", value)
		}
	})

	t.Run("String", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := "default"
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().GetStringEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(StringResolutionDetail{
				Value: "foo",
				ResolutionDetail: ResolutionDetail{
					Value:     "foo",
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.GetStringValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetStringValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from GetStringValue, got %v", value)
		}

		valueDetails, err := client.GetStringValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetStringValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(string) != defaultValue {
			t.Errorf("expected default value from GetStringValueDetails, got %v", value)
		}
	})

	t.Run("Number", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := 3.14159
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().GetNumberEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(NumberResolutionDetail{
				Value: 0,
				ResolutionDetail: ResolutionDetail{
					Value:     0,
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.GetNumberValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetNumberValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from GetNumberValue, got %v", value)
		}

		valueDetails, err := client.GetNumberValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetNumberValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(float64) != defaultValue {
			t.Errorf("expected default value from GetNumberValueDetails, got %v", value)
		}
	})

	t.Run("Object", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		type obj struct {
			foo string
		}
		defaultValue := obj{foo: "bar"}
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().GetObjectEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(ResolutionDetail{
				Value:     obj{foo: "foo"},
				ErrorCode: "GENERAL",
				Reason:    "forced test error",
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.GetObjectValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetObjectValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from GetObjectValue, got %v", value)
		}

		valueDetails, err := client.GetObjectValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected GetObjectValueDetails to return an error, got nil")
		}

		if valueDetails.Value.(obj) != defaultValue {
			t.Errorf("expected default value from GetObjectValueDetails, got %v", value)
		}
	})
}

// TODO Requirement_1_4_10

// Requirement_1_4_11 is satisfied by goroutines.

// TODO Requirement_1_5_1

// TODO Requirement_1_6_1
