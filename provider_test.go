package openfeature

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"
)

// The provider interface MUST define a `metadata` member or accessor, containing a `name` field or accessor
// of type string, which identifies the provider implementation.
func TestRequirement_2_1_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := NewMockProvider(ctrl)

	type requirements interface {
		Metadata() Metadata
	}

	var mockProviderI any = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the Metadata signature")
	}

	metadata := Metadata{}

	metaValue := reflect.ValueOf(&metadata).Elem()
	fieldName := "Name"

	field := metaValue.FieldByName(fieldName)
	if !field.IsValid() {
		t.Errorf("field %s doesn't exist in the Metadata struct", fieldName)
	}
}

// The `feature provider` interface MUST define methods to resolve flag values,
// with parameters `flag key` (string, required), `default value` (boolean | number | string | structure, required)
// and `evaluation context` (optional), which returns a `flag resolution` structure.
func TestRequirement_2_2_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := NewMockProvider(ctrl)

	type requirements interface {
		BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx FlattenedContext) BoolResolutionDetail
		StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx FlattenedContext) StringResolutionDetail
		FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx FlattenedContext) FloatResolutionDetail
		IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx FlattenedContext) IntResolutionDetail
		ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx FlattenedContext) ObjectResolutionDetail
	}

	var mockProviderI any = mockProvider
	if _, ok := mockProviderI.(requirements); !ok {
		t.Error("provider interface doesn't define the evaluation signatures")
	}
}

// Conditional_Requirement_2_2_2_1
// The `feature provider` interface MUST define methods for typed flag resolution,
// including boolean, numeric, string, and structure.
//
// Is satisfied by TestRequirement_2_2.

func TestTrackingEventDetails_Copy(t *testing.T) {
	tests := map[string]struct {
		inputDetails TrackingEventDetails
		copiedValue  float64
		outputDetail TrackingEventDetails
	}{
		"copied correctly": {
			inputDetails: NewTrackingEventDetails(1).Add("foo", "bar"),
			copiedValue:  2,
			outputDetail: NewTrackingEventDetails(2).Add("foo", "bar"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			copiedDetails := tc.inputDetails.Copy(tc.copiedValue)

			if !reflect.DeepEqual(copiedDetails, tc.outputDetail) {
				t.Errorf("copied value %v doesn't match the expected value  %v", copiedDetails, tc.outputDetail)
			}
		})
	}
}

func TestTrackingEventDetails_Add(t *testing.T) {
	type dummyStruct struct {
		qux string
	}

	tests := map[string]struct {
		inputDetails TrackingEventDetails
		addKeyPair   map[string]any
	}{
		"added correctly": {
			inputDetails: NewTrackingEventDetails(1),
			addKeyPair: map[string]any{
				"foo": "bar",
				"baz": 1,
				"qux": dummyStruct{
					qux: "qux",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for key, value := range tc.addKeyPair {
				tc.inputDetails.Add(key, value)
			}

			if !reflect.DeepEqual(tc.inputDetails.Attributes(), tc.addKeyPair) {
				t.Errorf("added key-pair %v not match with input key-pair %v", tc.addKeyPair, tc.inputDetails.Attributes())
			}
		})
	}
}
