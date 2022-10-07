package openfeature

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Hook context MUST provide: the `flag key`, `flag value type`, `evaluation context`, and the `default value`.
func TestRequirement_4_1_1(t *testing.T) {
	hookCtx := new(HookContext)

	metaValue := reflect.ValueOf(hookCtx).Elem()

	for _, name := range []string{"flagKey", "flagType", "evaluationContext", "defaultValue"} {
		field := metaValue.FieldByName(name)
		if field == (reflect.Value{}) {
			t.Errorf("field %s doesn't exist in the HookContext struct", name)
		}
	}
}

// The `hook context` SHOULD provide: access to the `client metadata` and the `provider metadata` fields.
func TestRequirement_4_1_2(t *testing.T) {
	hookCtx := HookContext{}

	type requirements interface {
		ClientMetadata() ClientMetadata
		ProviderMetadata() Metadata
	}

	var hookI interface{} = hookCtx
	if _, ok := hookI.(requirements); !ok {
		t.Error("HookContext doesn't implement the 4.1.2 requirements interface")
	}
}

// The `flag key`, `flag type`, and `default value` properties MUST be immutable.
// If the language does not support immutability, the hook MUST NOT modify these properties.
func TestRequirement_4_1_3(t *testing.T) {
	hookCtx := new(HookContext)

	metaValue := reflect.ValueOf(hookCtx).Elem()

	caser := cases.Title(language.English)

	for _, name := range []string{"flagKey", "flagType", "defaultValue"} {
		field := metaValue.FieldByName(name)
		if field == (reflect.Value{}) {
			t.Errorf("field %s doesn't exist in the HookContext struct", name)
		}

		if caser.String(name) == name {
			t.Errorf("field %s is uppercased and therefore mutable", name)
		}
	}
}

// Requirement_4_1_4
// The evaluation context MUST be mutable only within the `before` hook.
//
// Is satisfied by the evaluation context being immutable within the HookContext struct itself
// and by the Before signature returning an evaluation context that is used to mutate the HookContext directly.

// `hook hints` MUST be a structure supports definition of arbitrary properties,
// with keys of type `string`, and values of type `boolean | string | number | datetime | structure`..
func TestRequirement_4_2_1(t *testing.T) {
	hookHints := HookHints{}

	tpe := reflect.TypeOf(hookHints.mapOfHints)
	if tpe.Kind() != reflect.Map {
		t.Fatalf("expected HookHints kind to be map, got %s", tpe.Kind())
	}
	if tpe.Key().Kind() != reflect.String {
		t.Errorf("expected HookHints key to be string, got %s", tpe.Key().Kind())
	}
	if tpe.Elem().Kind() != reflect.Interface {
		t.Errorf("expected HookHints element to be interface{}, got %s", tpe.Elem().Kind())
	}
}

// Condition: `Hook hints` MUST be immutable.
func TestRequirement_4_2_2_1(t *testing.T) {
	hookHints := new(HookHints)

	metaValue := reflect.ValueOf(hookHints).Elem()

	fieldName := "mapOfHints"
	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the HookHints struct", fieldName)
	}

	caser := cases.Title(language.English)
	if caser.String(fieldName) == fieldName {
		t.Errorf("field %s is uppercased and therefore mutable", fieldName)
	}
}

// Condition: The client `metadata` field in the `hook context` MUST be immutable.
func TestRequirement_4_2_2_2(t *testing.T) {
	hookCtx := new(HookContext)

	metaValue := reflect.ValueOf(hookCtx).Elem()

	fieldName := "clientMetadata"
	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the HookContext struct", fieldName)
	}

	caser := cases.Title(language.English)
	if caser.String(fieldName) == fieldName {
		t.Errorf("field %s is uppercased and therefore mutable", fieldName)
	}
}

// Condition: The provider `metadata` field in the `hook context` MUST be immutable.
func TestRequirement_4_2_2_3(t *testing.T) {
	hookCtx := new(HookContext)

	metaValue := reflect.ValueOf(hookCtx).Elem()

	fieldName := "providerMetadata"
	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the HookContext struct", fieldName)
	}

	caser := cases.Title(language.English)
	if caser.String(fieldName) == fieldName {
		t.Errorf("field %s is uppercased and therefore mutable", fieldName)
	}
}

// Requirement_4_3_1
// Hooks MUST specify at least one stage.
//
// Has no suitable test as a Hook satisfies the interface by implementing the required functions.
// We can't be sure whether the function of each hook (before, after, error, finally) is empty or not, nor do we care
// as an empty function won't affect anything.

// The `before` stage MUST run before flag resolution occurs. It accepts a `hook context` (required) and
// `hook hints` (optional) as parameters and returns either an `evaluation context` or nothing.
func TestRequirement_4_3_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	t.Run("before stage MUST run before flag resolution occurs", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		flagKey := "foo"
		defaultValue := "bar"
		evalCtx := EvaluationContext{}
		flatCtx := flattenContext(evalCtx)

		mockProvider.EXPECT().Hooks().AnyTimes()

		// assert that the Before hooks are executed prior to the flag evaluation
		mockProvider.EXPECT().StringEvaluation(gomock.Any(), flagKey, defaultValue, flatCtx).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()))
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("function signature", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		type requirement interface {
			Before(hookContext HookContext, hookHints HookHints) (*EvaluationContext, error)
		}

		var hookI interface{} = mockHook
		if _, ok := hookI.(requirement); !ok {
			t.Error("hook doesn't implement the required Before func signature")
		}
	})
}

// Any `evaluation context` returned from a `before` hook MUST be passed to subsequent `before` hooks (via `HookContext`).
func TestRequirement_4_3_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	SetProvider(mockProvider)
	mockHook1 := NewMockHook(ctrl)
	mockHook2 := NewMockHook(ctrl)
	client := NewClient("test")

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{
		attributes: map[string]interface{}{
			"is": "a test",
		},
	}

	mockProvider.EXPECT().Hooks().AnyTimes()

	hook1Ctx := HookContext{
		flagKey:           flagKey,
		flagType:          String,
		defaultValue:      defaultValue,
		clientMetadata:    client.metadata,
		providerMetadata:  mockProvider.Metadata(),
		evaluationContext: evalCtx,
	}
	hook1EvalCtxResult := &EvaluationContext{targetingKey: "mockHook1"}
	mockHook1.EXPECT().Before(hook1Ctx, gomock.Any()).Return(hook1EvalCtxResult, nil)
	mockProvider.EXPECT().StringEvaluation(gomock.Any(), flagKey, defaultValue, map[string]interface{}{
		"is":         "a test",
		TargetingKey: "mockHook1",
	})

	// assert that the evaluation context returned by the first hook is passed into the second hook
	hook2Ctx := hook1Ctx
	hook2Ctx.evaluationContext = *hook1EvalCtxResult
	mockHook2.EXPECT().Before(hook2Ctx, gomock.Any())

	mockHook1.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
	mockHook2.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

	_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook1, mockHook2))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

// When `before` hooks have finished executing, any resulting `evaluation context` MUST be merged with the existing
// `evaluation context` in the following order:
// before-hook (highest precedence), invocation, client, api (lowest precedence).
func TestRequirement_4_3_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	SetProvider(mockProvider)
	mockHook := NewMockHook(ctrl)
	client := NewClient("test")

	apiEvalCtx := EvaluationContext{
		attributes: map[string]interface{}{
			"key":            "api context",
			"lowestPriority": true,
		},
	}
	SetEvaluationContext(apiEvalCtx)

	clientEvalCtx := EvaluationContext{
		attributes: map[string]interface{}{
			"key":            "client context",
			"lowestPriority": false,
			"beatsClient":    false,
		},
	}
	client.SetEvaluationContext(clientEvalCtx)

	flagKey := "foo"
	defaultValue := "bar"
	invEvalCtx := EvaluationContext{
		attributes: map[string]interface{}{
			"key":         "invocation context",
			"on":          true,
			"beatsClient": true,
		},
	}

	mockProvider.EXPECT().Hooks().AnyTimes()

	hookEvalCtxResult := &EvaluationContext{
		attributes: map[string]interface{}{
			"key":        "hook value",
			"multiplier": 3,
		},
	}
	mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(hookEvalCtxResult, nil)

	// assert that the EvaluationContext returned by Before hooks is merged with the invocation EvaluationContext
	expectedMergedContext := EvaluationContext{
		attributes: map[string]interface{}{
			"key":            "hook value", // hook takes precedence
			"multiplier":     3,
			"on":             true,
			"lowestPriority": false,
			"beatsClient":    true,
		},
	}
	mockProvider.EXPECT().StringEvaluation(gomock.Any(), flagKey, defaultValue, flattenContext(expectedMergedContext))
	mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

	_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, invEvalCtx, WithHooks(mockHook))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

// The `after` stage MUST run after flag resolution occurs. It accepts a `hook context` (required),
// `flag evaluation details` (required) and `hook hints` (optional). It has no return value.
func TestRequirement_4_3_5(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	t.Run("after hook MUST run after flag resolution occurs", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		flagKey := "foo"
		defaultValue := "bar"
		evalCtx := EvaluationContext{}
		flatCtx := flattenContext(evalCtx)

		mockProvider.EXPECT().Hooks().AnyTimes()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		// assert that the After hooks are executed after the flag evaluation
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockProvider.EXPECT().StringEvaluation(gomock.Any(), flagKey, defaultValue, flatCtx))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("function signature", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		type requirement interface {
			After(hookContext HookContext, flagEvaluationDetails InterfaceEvaluationDetails, hookHints HookHints) error
		}

		var hookI interface{} = mockHook
		if _, ok := hookI.(requirement); !ok {
			t.Error("hook doesn't implement the required After func signature")
		}
	})
}

// The `error` hook MUST run when errors are encountered in the `before` stage, the `after` stage or during flag
// resolution. It accepts `hook context` (required), `exception` representing what went wrong (required),
// and `hook hints` (optional). It has no return value.
func TestRequirement_4_3_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("error hook MUST run when errors are encountered in the before stage", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		mockProvider.EXPECT().Hooks().AnyTimes()

		// assert that the Error hooks are executed after the failed Before hooks
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced")))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("error hook MUST run when errors are encountered during flag evaluation", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		mockProvider.EXPECT().Hooks().AnyTimes()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		// assert that the Error hooks are executed after the failed flag evaluation
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(
				mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx).
					Return(StringResolutionDetail{
						ProviderResolutionDetail: ProviderResolutionDetail{
							ResolutionError: NewGeneralResolutionError("test"),
						},
					}),
			)
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("error hook MUST run when errors are encountered during flag evaluation", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		mockProvider.EXPECT().Hooks().AnyTimes()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx)
		// assert that the Error hooks are executed after the failed After hooks
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("forced")))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("function signature", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		type requirement interface {
			Error(hookContext HookContext, err error, hookHints HookHints)
		}

		var hookI interface{} = mockHook
		if _, ok := hookI.(requirement); !ok {
			t.Error("hook doesn't implement the required Error func signature")
		}
	})
}

// The `finally` hook MUST run after the `before`, `after`, and `error` stages. It accepts a `hook context` (required)
// and `hook hints` (optional). There is no return value.
func TestRequirement_4_3_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("finally hook MUST run after the before & after stages", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		mockProvider.EXPECT().Hooks().AnyTimes()

		// assert that the Finally hook runs after the Before & After stages
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()))
		mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx)

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("finally hook MUST run after the error stage", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)

		mockProvider.EXPECT().Hooks().AnyTimes()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
		// assert that the Finally hook runs after the Error stage
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()))

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("function signature", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		type requirement interface {
			Finally(hookContext HookContext, hookHints HookHints)
		}

		var hookI interface{} = mockHook
		if _, ok := hookI.(requirement); !ok {
			t.Error("hook doesn't implement the required Finally func signature")
		}
	})
}

// The API, Client, Provider and invocation MUST have a method for registering hooks
func TestRequirement_4_4_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	t.Run("API MUST have a method for registering hooks", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		AddHooks(mockHook)
	})

	t.Run("client MUST have a method for registering hooks", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		client.AddHooks(mockHook)

		type requirement interface {
			AddHooks(hooks ...Hook)
		}

		var clientI interface{} = client
		if _, ok := clientI.(requirement); !ok {
			t.Error("client doesn't implement the required AddHooks func signature")
		}
	})

	t.Run("provider MUST have a method for registering hooks", func(t *testing.T) {
		mockProvider := NewMockFeatureProvider(ctrl)

		type requirement interface {
			Hooks() []Hook
		}

		var providerI interface{} = mockProvider
		if _, ok := providerI.(requirement); !ok {
			t.Error("provider doesn't implement the required Hooks retrieval func signature")
		}
	})

	t.Run("invocation MUST have a method for registering hooks", func(t *testing.T) {
		client := NewClient("test")

		// EvaluationOptions contains the hooks registered at invocation
		type requirement interface {
			BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
			StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
			FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
			IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
			ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error)
			BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
			StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
			FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
			IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
			ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)
		}

		var clientI interface{} = client
		if _, ok := clientI.(requirement); !ok {
			t.Error("client doesn't implement the required func signatures containing EvaluationOptions")
		}
	})
}

// Hooks MUST be evaluated in the following order:  - before: API, Client, Invocation, Provider - after: Provider, Invocation, Client, API
// - error (if applicable): Provider, Invocation, Client, API - finally: Provider, Invocation, Client, API
func TestRequirement_4_4_2(t *testing.T) {
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("before, after & finally hooks MUST be evaluated in the following order", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		mockAPIHook := NewMockHook(ctrl)
		AddHooks(mockAPIHook)
		mockClientHook := NewMockHook(ctrl)
		client := NewClient("test")
		client.AddHooks(mockClientHook)
		mockInvocationHook := NewMockHook(ctrl)
		mockProviderHook := NewMockHook(ctrl)

		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)
		mockProvider.EXPECT().Hooks().Return([]Hook{mockProviderHook}).Times(2)

		// before: API, Client, Invocation, Provider
		mockProviderHook.EXPECT().Before(gomock.Any(), gomock.Any()).
			After(mockInvocationHook.EXPECT().Before(gomock.Any(), gomock.Any())).
			After(mockClientHook.EXPECT().Before(gomock.Any(), gomock.Any())).
			After(mockAPIHook.EXPECT().Before(gomock.Any(), gomock.Any()))

		// after: Invocation, Client, API
		mockAPIHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockProviderHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()))

		// finally: Invocation, Client, API
		mockAPIHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().Finally(gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().Finally(gomock.Any(), gomock.Any())).
			After(mockProviderHook.EXPECT().Finally(gomock.Any(), gomock.Any()))

		mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx)

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockInvocationHook))
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("error hooks MUST be evaluated in the following order", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		mockAPIHook := NewMockHook(ctrl)
		AddHooks(mockAPIHook)
		mockClientHook := NewMockHook(ctrl)
		client := NewClient("test")
		client.AddHooks(mockClientHook)
		mockInvocationHook := NewMockHook(ctrl)
		mockProviderHook := NewMockHook(ctrl)

		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)
		mockProvider.EXPECT().Hooks().Return([]Hook{mockProviderHook}).Times(2)

		mockAPIHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))

		// error: Provider, Invocation, Client, API
		mockAPIHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockProviderHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()))

		mockProviderHook.EXPECT().Finally(gomock.Any(), gomock.Any())
		mockInvocationHook.EXPECT().Finally(gomock.Any(), gomock.Any())
		mockClientHook.EXPECT().Finally(gomock.Any(), gomock.Any())
		mockAPIHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockInvocationHook))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// Requirement_4_4_3
// If a `finally` hook abnormally terminates, evaluation MUST proceed, including the execution of any
// remaining `finally` hooks.
//
// Is satisfied by the Finally hook func signature not returning an error.

// Requirement_4_4_4
// If an `error` hook abnormally terminates, evaluation MUST proceed, including the execution of any remaining
// `error` hooks.
//
// Is satisfied by the Error hook func signature not returning an error.

// Requirement_4_4_5
// If an error occurs in the `before` or `after` hooks, the `error` hooks MUST be invoked.
//
// Is satisfied by TestRequirement_4_3_6.

// If an error occurs during the evaluation of `before` or `after` hooks, any remaining hooks in the `before` or `after`
// stages MUST NOT be invoked.
func TestRequirement_4_4_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run(
		"if an error occurs during the evaluation of before hooks, any remaining before hooks MUST NOT be invoked",
		func(t *testing.T) {
			mockHook1 := NewMockHook(ctrl)
			mockHook2 := NewMockHook(ctrl)
			client := NewClient("test")
			mockProvider := NewMockFeatureProvider(ctrl)
			mockProvider.EXPECT().Metadata().AnyTimes()
			SetProvider(mockProvider)

			mockProvider.EXPECT().Hooks().AnyTimes()

			mockHook1.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
			// the lack of mockHook2.EXPECT().Before() asserts that remaining hooks aren't invoked after an error
			mockHook1.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

			_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook1, mockHook2))
			if err == nil {
				t.Error("expected error, got nil")
			}
		},
	)

	t.Run(
		"if an error occurs during the evaluation of after hooks, any remaining after hooks MUST NOT be invoked",
		func(t *testing.T) {
			mockHook1 := NewMockHook(ctrl)
			mockHook2 := NewMockHook(ctrl)
			client := NewClient("test")
			mockProvider := NewMockFeatureProvider(ctrl)
			mockProvider.EXPECT().Metadata().AnyTimes()
			SetProvider(mockProvider)

			mockProvider.EXPECT().Hooks().AnyTimes()

			mockHook1.EXPECT().Before(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Before(gomock.Any(), gomock.Any())
			mockHook1.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("forced"))
			// the lack of mockHook2.EXPECT().After() asserts that remaining hooks aren't invoked after an error
			mockHook1.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

			mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx)

			_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook1, mockHook2))
			if err == nil {
				t.Error("expected error, got nil")
			}
		},
	)
}

// If an error occurs in the `before` hooks, the default value MUST be returned.
func TestRequirement_4_4_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	mockHook := NewMockHook(ctrl)
	client := NewClient("test")
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	SetProvider(mockProvider)

	mockProvider.EXPECT().Hooks().AnyTimes()

	mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
	mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

	res, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook))
	if err == nil {
		t.Error("expected error, got nil")
	}

	if res.Value != defaultValue {
		t.Errorf("expected default value, got %s", res.Value)
	}
}

// `Flag evaluation options` MAY contain `hook hints`, a map of data to be provided to hook invocations.
func TestRequirement_4_5_1(t *testing.T) {
	evalOptions := &EvaluationOptions{}
	option := WithHookHints(NewHookHints(map[string]interface{}{"foo": "bar"}))
	option(evalOptions)
	if evalOptions.hookHints.Value("foo") != "bar" {
		t.Error("hook hints not set to EvaluationOptions")
	}
}

// `hook hints` MUST be passed to each hook.
func TestRequirement_4_5_2(t *testing.T) {
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	flatCtx := flattenContext(evalCtx)

	t.Run("hook hints must be passed to before, after & finally hooks", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)
		mockProvider.EXPECT().Hooks().AnyTimes()

		hookHints := NewHookHints(map[string]interface{}{"foo": "bar"})

		mockHook.EXPECT().Before(gomock.Any(), hookHints)
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), hookHints)
		mockHook.EXPECT().Finally(gomock.Any(), hookHints)

		mockProvider.EXPECT().StringEvaluation(context.Background(), flagKey, defaultValue, flatCtx)

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook), WithHookHints(hookHints))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("hook hints must be passed to error hooks", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		mockProvider.EXPECT().Metadata().AnyTimes()
		SetProvider(mockProvider)
		mockProvider.EXPECT().Hooks().AnyTimes()

		hookHints := NewHookHints(map[string]interface{}{"foo": "bar"})

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), hookHints)
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, evalCtx, WithHooks(mockHook), WithHookHints(hookHints))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// The hook MUST NOT alter the `hook hints` structure.
func TestRequirement_4_5_3(t *testing.T) {
	hookHints := NewHookHints(map[string]interface{}{"foo": "bar"})

	metaValue := reflect.ValueOf(&hookHints).Elem()

	// checking that the HookHints struct has an unexported field containing the map of hints is enough
	// to assert that the map cannot be altered
	fieldName := "mapOfHints"
	field := metaValue.FieldByName(fieldName)

	if field == (reflect.Value{}) {
		t.Errorf("field %s doesn't exist in the HookHints struct", fieldName)
	}

	if hookHints.Value("foo") != hookHints.mapOfHints["foo"] {
		t.Errorf("expected to retrieve the hint from the underlying map")
	}
}
