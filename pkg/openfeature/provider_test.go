package openfeature

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestRequirement_2_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)

	type requirements interface {
		Metadata() Metadata
	}

	var mockProviderI interface{} = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the Metadata signature")
	}

	metadata := Metadata{}

	metaValue := reflect.ValueOf(&metadata).Elem()
	fieldName := "Name"

	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the Metadata struct", fieldName)
	}
}

func TestRequirement_2_2(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)

	type requirements interface {
		BooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) BoolResolutionDetail
		StringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) StringResolutionDetail
		FloatEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) FloatResolutionDetail
		IntEvaluation(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) IntResolutionDetail
		ObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) ResolutionDetail
	}

	var mockProviderI interface{} = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the evaluation signatures")
	}
}

// Conditional_Requirement_2_3_1 is satisfied by TestRequirement_2_2.
