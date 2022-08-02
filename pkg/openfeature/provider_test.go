package openfeature

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

// The provider interface MUST define a `metadata` member or accessor, containing a `name` field or accessor
// of type string, which identifies the provider implementation.
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

// The `feature provider` interface MUST define methods to resolve flag values,
// with parameters `flag key` (string, required), `default value` (boolean | number | string | structure, required),
// `evaluation context` (optional), and `evaluation options` (optional), which returns a `flag resolution` structure.
func TestRequirement_2_2(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)

	type requirements interface {
		BooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) BoolResolutionDetail
		StringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) StringResolutionDetail
		NumberEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) NumberResolutionDetail
		ObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) ResolutionDetail
	}

	var mockProviderI interface{} = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the evaluation signatures")
	}
}

// Conditional_Requirement_2_3_1
// The `feature provider` interface MUST define methods for typed flag resolution,
// including boolean, numeric, string, and structure.
//
// Is satisfied by TestRequirement_2_2.
