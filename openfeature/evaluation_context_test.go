package openfeature

import (
	"context"
	"reflect"
	"testing"

	"github.com/open-feature/go-sdk/openfeature/internal"
	"go.uber.org/mock/gomock"
)

// The `evaluation context` structure MUST define an optional `targeting key` field of type string,
// identifying the subject of the flag evaluation.
func TestRequirement_3_1_1(t *testing.T) {
	evalCtx := &EvaluationContext{}

	metaValue := reflect.ValueOf(evalCtx).Elem()
	fieldName := "targetingKey"

	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the EvaluationContext struct", fieldName)
	}
}

// The evaluation context MUST support the inclusion of custom fields,
// having keys of type `string`, and values of type `boolean | string | number | datetime | structure`.
func TestRequirement_3_1_2(t *testing.T) {
	evalCtx := EvaluationContext{}

	tpe := reflect.TypeOf(evalCtx.attributes)
	if tpe.Kind() != reflect.Map {
		t.Fatalf("expected EvaluationContext.attributes kind to be map, got %s", tpe.Kind())
	}
	if tpe.Key().Kind() != reflect.String {
		t.Errorf("expected EvaluationContext.attributes key to be string, got %s", tpe.Key().Kind())
	}
	if tpe.Elem().Kind() != reflect.Interface {
		t.Errorf("expected EvaluationContext.attributes value to be any, got %s", tpe.Elem().Kind())
	}
}

// The API, Client and invocation MUST have a method for supplying `evaluation context`.
func TestRequirement_3_2_1(t *testing.T) {
	t.Cleanup(initSingleton)

	t.Run("API MUST have a method for supplying `evaluation context`", func(t *testing.T) {
		SetEvaluationContext(EvaluationContext{})
	})

	t.Run("client MUST have a method for supplying `evaluation context`", func(t *testing.T) {
		client := NewClient("test")

		type requirement interface {
			SetEvaluationContext(evalCtx EvaluationContext)
		}

		var clientI any = client
		if _, ok := clientI.(requirement); !ok {
			t.Error("client doesn't implement the required SetEvaluationContext func signature")
		}
	})

	t.Run("invocation MUST have a method for supplying `evaluation context`", func(t *testing.T) {
		client := NewClient("test")

		type requirement interface {
			BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
			StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
			FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
			IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
			ObjectValue(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (any, error)
			BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
			StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
			FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
			IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
			ObjectValueDetails(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)
		}

		var clientI any = client
		if _, ok := clientI.(requirement); !ok {
			t.Error("client doesn't implement the required func signatures containing EvaluationContext")
		}
	})
}

// Evaluation context MUST be merged in the order: API (global) - transaction - client - invocation,
// with duplicate values being overwritten.
func TestRequirement_3_2_2(t *testing.T) {
	t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	apiEvalCtx := EvaluationContext{
		targetingKey: "API",
		attributes: map[string]any{
			"invocationEvalCtx": true,
			"foo":               3,
			"user":              3,
		},
	}
	SetEvaluationContext(apiEvalCtx)

	transactionEvalCtx := EvaluationContext{
		targetingKey: "Transcation",
		attributes: map[string]any{
			"transactionEvalCtx": true,
			"foo":                2,
			"user":               2,
		},
	}
	transactionCtx := WithTransactionContext(context.Background(), transactionEvalCtx)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()

	err := SetNamedProviderAndWait(t.Name(), mockProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	client := GetNamedClient(t.Name())
	clientEvalCtx := EvaluationContext{
		targetingKey: "Client",
		attributes: map[string]any{
			"clientEvalCtx": true,
			"foo":           1,
			"user":          1,
		},
	}
	client.SetEvaluationContext(clientEvalCtx)

	invocationEvalCtx := EvaluationContext{
		targetingKey: "",
		attributes: map[string]any{
			"apiEvalCtx": true,
			"foo":        "bar",
		},
	}

	mockProvider.EXPECT().Hooks().AnyTimes()
	expectedMergedEvalCtx := EvaluationContext{
		targetingKey: "Client",
		attributes: map[string]any{
			"apiEvalCtx":         true,
			"transactionEvalCtx": true,
			"invocationEvalCtx":  true,
			"clientEvalCtx":      true,
			"foo":                "bar",
			"user":               1,
		},
	}
	flatCtx := flattenContext(expectedMergedEvalCtx)
	mockProvider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), flatCtx)

	_, err = client.StringValue(transactionCtx, "foo", "bar", invocationEvalCtx)
	if err != nil {
		t.Error(err)
	}
}

func TestEvaluationContext_AttributesNotPassedByReference(t *testing.T) {
	attributes := map[string]any{
		"foo": "bar",
	}
	evalCtx := NewEvaluationContext("foo", attributes)

	attributes["immutabilityCheck"] = "safe"

	if _, ok := evalCtx.attributes["immutabilityCheck"]; ok {
		t.Error("mutation of map passed to NewEvaluationContext caused a mutation of its attributes field")
	}
}

func TestRequirement_3_3_1(t *testing.T) {
	t.Run("The API MUST have a method for setting the evaluation context of the transaction context propagator for the current transaction.", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTransactionContext(ctx, EvaluationContext{})
		val, ok := ctx.Value(internal.TransactionContext).(EvaluationContext)

		if !ok {
			t.Fatalf("failed to find transcation context set from WithTransactionContext: %v", val)
		}
	})
}

func TestEvaluationContext_AttributesFuncNotPassedByReference(t *testing.T) {
	evalCtx := NewEvaluationContext("foo", map[string]any{
		"foo": "bar",
	})

	attributes := evalCtx.Attributes()
	attributes["immutabilityCheck"] = "safe"

	if _, ok := evalCtx.attributes["immutabilityCheck"]; ok {
		t.Error("mutation of map passed to SetAttributes caused a mutation of its attributes field")
	}
}

func TestNewTargetlessEvaluationContext(t *testing.T) {
	attributes := map[string]any{
		"foo": "bar",
	}
	evalCtx := NewTargetlessEvaluationContext(attributes)
	if evalCtx.targetingKey != "" {
		t.Error("targeting key should not be set with NewTargetlessEvaluationContext")
	}

	if !reflect.DeepEqual(evalCtx.Attributes(), attributes) {
		t.Errorf("we expect no difference in the attributes")
	}
}

func TestMergeTransactionContext(t *testing.T) {
	oldEvalCtx := NewEvaluationContext("old", map[string]any{
		"old":       true,
		"overwrite": "old",
	})
	newEvalCtx := NewEvaluationContext("new", map[string]any{
		"new":       true,
		"overwrite": "new",
	})

	ctx := WithTransactionContext(context.Background(), oldEvalCtx)
	ctx = MergeTransactionContext(ctx, newEvalCtx)

	expectedMergedEvalCtx := EvaluationContext{
		targetingKey: "new",
		attributes: map[string]any{
			"old":       true,
			"new":       true,
			"overwrite": "new",
		},
	}

	finalTransactionContext := TransactionContext(ctx)

	if finalTransactionContext.targetingKey != expectedMergedEvalCtx.targetingKey {
		t.Errorf(
			"targetingKey is not expected value, finalTransactionContext.targetingKey: %s, newEvalCtx.targetingKey: %s",
			finalTransactionContext.TargetingKey(),
			expectedMergedEvalCtx.TargetingKey(),
		)
	}

	if !reflect.DeepEqual(finalTransactionContext.Attributes(), expectedMergedEvalCtx.Attributes()) {
		t.Errorf(
			"attributes are not expected value, finalTransactionContext.Attributes(): %v, expectedMergedEvalCtx.Attributes(): %v",
			finalTransactionContext.Attributes(),
			expectedMergedEvalCtx.Attributes(),
		)
	}
}
