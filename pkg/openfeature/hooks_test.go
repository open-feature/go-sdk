package openfeature

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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

// Requirement_4_1_4 is satisfied by the evaluation context being immutable within the HookContext struct itself
// and by the Before signature returning an evaluation context that is used to mutate the HookContext directly.

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

// Requirement_4_3_1 has no suitable test as a Hook satisfies the interface by implementing the required functions.
// We can't be sure whether the function of each hook (before, after, error, finally) is empty or not, nor do we care
// as an empty function won't affect anything.

func TestRequirement_4_3_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	t.Run("before stage MUST run before flag resolution occurs", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)

		flagKey := "foo"
		defaultValue := "bar"
		evalCtx := EvaluationContext{}
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		// assert that the Before hooks are executed prior to the flag evaluation
		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()))
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
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

func TestRequirement_4_3_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	SetProvider(mockProvider)
	mockHook1 := NewMockHook(ctrl)
	mockHook2 := NewMockHook(ctrl)
	client := NewClient("test")

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}
	evalOptions := NewEvaluationOptions([]Hook{mockHook1, mockHook2}, HookHints{})

	mockProvider.EXPECT().Metadata().Times(2)
	mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)

	hook1Ctx := HookContext{
		flagKey:           flagKey,
		flagType:          String,
		defaultValue:      defaultValue,
		clientMetadata:    client.metadata,
		providerMetadata:  mockProvider.Metadata(),
		evaluationContext: evalCtx,
	}
	hook1EvalCtxResult := &EvaluationContext{TargetingKey: "mockHook1"}
	mockHook1.EXPECT().Before(hook1Ctx, gomock.Any()).Return(hook1EvalCtxResult, nil)

	// assert that the evaluation context returned by the first hook is passed into the second hook
	hook2Ctx := hook1Ctx
	hook2Ctx.evaluationContext = *hook1EvalCtxResult
	mockHook2.EXPECT().Before(hook2Ctx, gomock.Any())

	mockHook1.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
	mockHook2.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

	_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRequirement_4_3_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	SetProvider(mockProvider)
	mockHook := NewMockHook(ctrl)
	client := NewClient("test")

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{
		Attributes: map[string]interface{}{
			"key": "initial value",
			"on":  true,
		},
	}
	evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

	mockProvider.EXPECT().Metadata()

	hookEvalCtxResult := &EvaluationContext{
		Attributes: map[string]interface{}{
			"key":        "hook value",
			"multiplier": 3,
		},
	}
	mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(hookEvalCtxResult, nil)

	// assert that the EvaluationContext returned by Before hooks is merged with the invocation EvaluationContext
	expectedMergedContext := EvaluationContext{
		Attributes: map[string]interface{}{
			"key":        "initial value", // invocation takes precedence
			"multiplier": 3,
			"on":         true,
		},
	}
	mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, expectedMergedContext, evalOptions)
	mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

	_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRequirement_4_3_5(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	t.Run("after hook MUST run after flag resolution occurs", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)

		flagKey := "foo"
		defaultValue := "bar"
		evalCtx := EvaluationContext{}
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		// assert that the After hooks are executed after the flag evaluation
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("function signature", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)

		type requirement interface {
			After(hookContext HookContext, flagEvaluationDetails EvaluationDetails, hookHints HookHints) error
		}

		var hookI interface{} = mockHook
		if _, ok := hookI.(requirement); !ok {
			t.Error("hook doesn't implement the required After func signature")
		}
	})
}

func TestRequirement_4_3_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	t.Run("error hook MUST run when errors are encountered in the before stage", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		// assert that the Error hooks are executed after the failed Before hooks
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced")))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("error hook MUST run when errors are encountered during flag evaluation", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		// assert that the Error hooks are executed after the failed flag evaluation
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(
				mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions).
					Return(StringResolutionDetail{ResolutionDetail: ResolutionDetail{ErrorCode: "forced"}}),
			)
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("error hook MUST run when errors are encountered during flag evaluation", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any())
		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)
		// assert that the Error hooks are executed after the failed After hooks
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("forced")))
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
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

func TestRequirement_4_3_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	t.Run("finally hook MUST run after the before & after stages", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		// assert that the Finally hook runs after the Before & After stages
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockHook.EXPECT().Before(gomock.Any(), gomock.Any()))
		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	})

	t.Run("finally hook MUST run after the error stage", func(t *testing.T) {
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

		mockProvider.EXPECT().Metadata()

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
		// assert that the Finally hook runs after the Error stage
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()))

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
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

	t.Run("invocation MUST have a method for registering hooks", func(t *testing.T) {
		client := NewClient("test")

		// EvaluationOptions contains the hooks registered at invocation
		type requirement interface {
			BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error)
			StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error)
			FloatValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error)
			IntValue(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (int64, error)
			ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error)
			BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
			StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
			FloatValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
			IntValueDetails(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
			ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
		}

		var clientI interface{} = client
		if _, ok := clientI.(requirement); !ok {
			t.Error("client doesn't implement the required func signatures containing EvaluationOptions")
		}
	})
}

func TestRequirement_4_4_2(t *testing.T) {
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	t.Run("before, after & finally hooks MUST be evaluated in the following order", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		mockAPIHook := NewMockHook(ctrl)
		AddHooks(mockAPIHook)
		mockClientHook := NewMockHook(ctrl)
		client := NewClient("test")
		client.AddHooks(mockClientHook)
		mockInvocationHook := NewMockHook(ctrl)
		evalOptions := NewEvaluationOptions([]Hook{mockInvocationHook}, HookHints{})

		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()

		// before: API, Client, Invocation
		mockInvocationHook.EXPECT().Before(gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().Before(gomock.Any(), gomock.Any())).
			After(mockAPIHook.EXPECT().Before(gomock.Any(), gomock.Any()))

		// after: Invocation, Client, API
		mockAPIHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()))

		// finally: Invocation, Client, API
		mockAPIHook.EXPECT().Finally(gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().Finally(gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().Finally(gomock.Any(), gomock.Any()))

		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
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
		evalOptions := NewEvaluationOptions([]Hook{mockInvocationHook}, HookHints{})

		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()

		mockAPIHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))

		// error: Invocation, Client, API
		mockAPIHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).
			After(mockClientHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())).
			After(mockInvocationHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()))

		mockInvocationHook.EXPECT().Finally(gomock.Any(), gomock.Any())
		mockClientHook.EXPECT().Finally(gomock.Any(), gomock.Any())
		mockAPIHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// Requirement_4_4_3 is satisfied by the Finally hook func signature not returning an error.

// Requirement_4_4_4 is satisfied by the Error hook func signature not returning an error.

// Requirement_4_4_5 is satisfied by TestRequirement_4_3_6.

func TestRequirement_4_4_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	t.Run(
		"if an error occurs during the evaluation of before hooks, any remaining before hooks MUST NOT be invoked",
		func(t *testing.T) {
			mockHook1 := NewMockHook(ctrl)
			mockHook2 := NewMockHook(ctrl)
			client := NewClient("test")
			mockProvider := NewMockFeatureProvider(ctrl)
			SetProvider(mockProvider)
			evalOptions := NewEvaluationOptions([]Hook{mockHook1, mockHook2}, HookHints{})

			mockProvider.EXPECT().Metadata()

			mockHook1.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
			// the lack of mockHook2.EXPECT().Before() asserts that remaining hooks aren't invoked after an error
			mockHook1.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

			_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
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
			SetProvider(mockProvider)
			evalOptions := NewEvaluationOptions([]Hook{mockHook1, mockHook2}, HookHints{})

			mockProvider.EXPECT().Metadata()

			mockHook1.EXPECT().Before(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Before(gomock.Any(), gomock.Any())
			mockHook1.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("forced"))
			// the lack of mockHook2.EXPECT().After() asserts that remaining hooks aren't invoked after an error
			mockHook1.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook1.EXPECT().Finally(gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			mockHook2.EXPECT().Finally(gomock.Any(), gomock.Any())

			mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)

			_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
			if err == nil {
				t.Error("expected error, got nil")
			}
		},
	)
}

func TestRequirement_4_4_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	mockHook := NewMockHook(ctrl)
	client := NewClient("test")
	mockProvider := NewMockFeatureProvider(ctrl)
	SetProvider(mockProvider)
	evalOptions := NewEvaluationOptions([]Hook{mockHook}, HookHints{})

	mockProvider.EXPECT().Metadata()

	mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
	mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
	mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

	res, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
	if err == nil {
		t.Error("expected error, got nil")
	}

	resString := res.Value.(string)
	if resString != defaultValue {
		t.Errorf("expected default value, got %s", resString)
	}
}

func TestRequirement_4_5_1(t *testing.T) {
	NewEvaluationOptions(nil, NewHookHints(map[string]interface{}{"foo": "bar"}))
}

func TestRequirement_4_5_2(t *testing.T) {
	ctrl := gomock.NewController(t)

	flagKey := "foo"
	defaultValue := "bar"
	evalCtx := EvaluationContext{}

	t.Run("hook hints must be passed to before, after & finally hooks", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()

		hookHints := NewHookHints(map[string]interface{}{"foo": "bar"})
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, hookHints)

		mockHook.EXPECT().Before(gomock.Any(), hookHints)
		mockHook.EXPECT().After(gomock.Any(), gomock.Any(), hookHints)
		mockHook.EXPECT().Finally(gomock.Any(), hookHints)

		mockProvider.EXPECT().StringEvaluation(flagKey, defaultValue, evalCtx, evalOptions)

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("hook hints must be passed to error hooks", func(t *testing.T) {
		defer t.Cleanup(initSingleton)
		mockHook := NewMockHook(ctrl)
		client := NewClient("test")
		mockProvider := NewMockFeatureProvider(ctrl)
		SetProvider(mockProvider)
		mockProvider.EXPECT().Metadata()

		hookHints := NewHookHints(map[string]interface{}{"foo": "bar"})
		evalOptions := NewEvaluationOptions([]Hook{mockHook}, hookHints)

		mockHook.EXPECT().Before(gomock.Any(), gomock.Any()).Return(nil, errors.New("forced"))
		mockHook.EXPECT().Error(gomock.Any(), gomock.Any(), hookHints)
		mockHook.EXPECT().Finally(gomock.Any(), gomock.Any())

		_, err := client.StringValueDetails(flagKey, defaultValue, evalCtx, evalOptions)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

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
