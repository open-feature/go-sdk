package openfeature

import (
	"github.com/golang/mock/gomock"
	"testing"
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
		GetBooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (bool, error)
		GetStringValue(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (string, error)
		GetNumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (float64, error)
		GetObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (interface{}, error)
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
		GetBooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		GetStringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		GetNumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
		GetObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
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
		evDetails, err := client.GetBooleanValueDetails(flagKey, true, nil)
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
		evDetails, err := client.GetStringValueDetails(flagKey, "", nil)
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
		evDetails, err := client.GetNumberValueDetails(flagKey, 1, nil)
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
		evDetails, err := client.GetObjectValueDetails(flagKey, 1, nil)
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

// Requirement_1_4_9 is satisfied by the error included in the return signatures, as is idiomatic in Go. Errors aren't
// fatal, the operations won't terminate as a result of an error.

// TODO Requirement_1_4_10

// TODO Requirement_1_4_11

// TODO Requirement_1_5_1

// TODO Requirement_1_6_1
