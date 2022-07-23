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
		BooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) BoolResolutionDetail
		StringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) StringResolutionDetail
		NumberEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) NumberResolutionDetail
		ObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) ResolutionDetail
	}

	var mockProviderI interface{} = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the evaluation signatures")
	}
}

// Conditional_Requirement_2_3_1 is satisfied by TestRequirement_2_2.
