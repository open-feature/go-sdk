package openfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
)

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
func TestRequirements_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error)
		StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error)
		FloatValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error)
		IntValue(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (int64, error)
		ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.3.* requirements interface")
	}
}

func TestRequirement_1_4_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	type requirements interface {
		BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		FloatValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		IntValueDetails(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	}

	var clientI interface{} = client
	if _, ok := clientI.(requirements); !ok {
		t.Error("client returned by NewClient doesn't implement the 1.4.1 requirements interface")
	}
}

// Requirement_1_4_2 has no suitable test as the provider implementation populates the EvaluationDetails value field

// TODO Requirement_1_4_3 once upgraded Go to 1.18 for generics

func TestRequirement_1_4_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	client := NewClient("test-client")

	flagKey := "foo"

	t.Run("BooleanValueDetails", func(t *testing.T) {
		evDetails, err := client.BooleanValueDetails(flagKey, true, EvaluationContext{}, EvaluationOptions{})
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
		evDetails, err := client.StringValueDetails(flagKey, "", EvaluationContext{}, EvaluationOptions{})
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
		evDetails, err := client.FloatValueDetails(flagKey, 1, EvaluationContext{}, EvaluationOptions{})
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
		evDetails, err := client.IntValueDetails(flagKey, 1, EvaluationContext{}, EvaluationOptions{})
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
		evDetails, err := client.ObjectValueDetails(flagKey, 1, EvaluationContext{}, EvaluationOptions{})
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
	client := NewClient("test-client")
	flagKey := "flag-key"

	ctrl := gomock.NewController(t)

	t.Run("Boolean", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockProvider := NewMockFeatureProvider(ctrl)
		defaultValue := true
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().BooleanEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(BoolResolutionDetail{
				Value: false,
				ResolutionDetail: ResolutionDetail{
					Value:     false,
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.BooleanValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected BooleanValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from BooleanValue, got %v", value)
		}

		valueDetails, err := client.BooleanValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
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
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(StringResolutionDetail{
				Value: "foo",
				ResolutionDetail: ResolutionDetail{
					Value:     "foo",
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.StringValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected StringValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from StringValue, got %v", value)
		}

		valueDetails, err := client.StringValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
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
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().FloatEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(FloatResolutionDetail{
				Value: 0,
				ResolutionDetail: ResolutionDetail{
					Value:     0,
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.FloatValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected FloatValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from FloatValue, got %v", value)
		}

		valueDetails, err := client.FloatValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
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
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().IntEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(IntResolutionDetail{
				Value: 0,
				ResolutionDetail: ResolutionDetail{
					Value:     0,
					ErrorCode: "GENERAL",
					Reason:    "forced test error",
				},
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.IntValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected IntValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from IntValue, got %v", value)
		}

		valueDetails, err := client.IntValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
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
		mockProvider.EXPECT().Metadata().Times(2)
		mockProvider.EXPECT().ObjectEvaluation(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{}).
			Return(ResolutionDetail{
				Value:     obj{foo: "foo"},
				ErrorCode: "GENERAL",
				Reason:    "forced test error",
			}).Times(2)
		SetProvider(mockProvider)

		value, err := client.ObjectValue(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected ObjectValue to return an error, got nil")
		}

		if value != defaultValue {
			t.Errorf("expected default value from ObjectValue, got %v", value)
		}

		valueDetails, err := client.ObjectValueDetails(flagKey, defaultValue, EvaluationContext{}, EvaluationOptions{})
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

// Requirement_1_5_1 is tested by TestRequirement_4_4_2.

// TODO Requirement_1_6_1

func TestClient_ProviderEvaluationReturnsUnexpectedType(t *testing.T) {
	client := NewClient("test-client")

	t.Run("Boolean", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		ctrl := gomock.NewController(t)
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()
		mockProvider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(BoolResolutionDetail{ResolutionDetail: ResolutionDetail{Value: 3}})

		_, err := client.BooleanValue("", false, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected BooleanValue to return an error, got nil")
		}
	})

	t.Run("String", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		ctrl := gomock.NewController(t)
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()
		mockProvider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(StringResolutionDetail{ResolutionDetail: ResolutionDetail{Value: 3}})

		_, err := client.StringValue("", "", EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected StringValue to return an error, got nil")
		}
	})

	t.Run("Float", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		ctrl := gomock.NewController(t)
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()
		mockProvider.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(FloatResolutionDetail{ResolutionDetail: ResolutionDetail{Value: false}})

		_, err := client.FloatValue("", 3, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected FloatValue to return an error, got nil")
		}
	})

	t.Run("Int", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		ctrl := gomock.NewController(t)
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()
		mockProvider.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(IntResolutionDetail{ResolutionDetail: ResolutionDetail{Value: false}})

		_, err := client.IntValue("", 3, EvaluationContext{}, EvaluationOptions{})
		if err == nil {
			t.Error("expected IntValue to return an error, got nil")
		}
	})
}
