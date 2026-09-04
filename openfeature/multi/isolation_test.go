package multi

import (
	"context"
	"errors"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// stateHandler is a minimal StateHandler with configurable callbacks for tests.
type stateHandler struct {
	initF     func(of.EvaluationContext) error
	shutdownF func()
}

func (s stateHandler) Init(e of.EvaluationContext) error {
	if s.initF != nil {
		return s.initF(e)
	}
	return nil
}

func (s stateHandler) Shutdown() {
	if s.shutdownF != nil {
		s.shutdownF()
	}
}

func Test_HookIsolator_BeforeCapturesData(t *testing.T) {
	hookCtx := of.NewHookContext(
		"test-key",
		of.Boolean,
		false,
		of.ClientMetadata{},
		of.Metadata{},
		of.NewEvaluationContext("target", map[string]any{}),
	)
	hookHints := of.NewHookHints(map[string]any{"foo": "bar"})
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, []of.Hook{})
	assert.Zero(t, isolator.capturedContext)
	assert.Zero(t, isolator.capturedHints)
	evalCtx, err := isolator.Before(t.Context(), hookCtx, hookHints)
	require.NoError(t, err)
	assert.NotNil(t, evalCtx)
	assert.Equal(t, hookCtx, isolator.capturedContext)
	assert.Equal(t, hookHints, isolator.capturedHints)
	assert.Equal(t, "test-provider", isolator.Name())
}

func Test_HookIsolator_Hooks_ReturnsSelf(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, []of.Hook{})
	hooks := isolator.Hooks()
	assert.NotEmpty(t, hooks)
	assert.Same(t, isolator, hooks[0])
}

func Test_HookIsolator_ExecutesHooksDuringEvaluation_NoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	testHook := of.NewMockHook(ctrl)
	testHook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	testHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	testHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	testHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{testHook})
	provider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(of.BoolResolutionDetail{
		Value:                    true,
		ProviderResolutionDetail: of.ProviderResolutionDetail{},
	})

	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, nil)
	result := isolator.BooleanEvaluation(t.Context(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.True(t, result.Value)
}

func Test_HookIsolator_ExecutesHooksDuringEvaluation_BeforeErrorAbortsExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	testHook := of.NewMockHook(ctrl)
	testHook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	testHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	testHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	testHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{testHook})

	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, nil)
	result := isolator.BooleanEvaluation(t.Context(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.False(t, result.Value)
}

func Test_toProviderResolutionDetail(t *testing.T) {
	tests := []struct {
		name             string
		input            of.InterfaceEvaluationDetails
		wantErrorCode    of.ErrorCode
		wantReason       of.Reason
		wantErr          bool
		wantVariant      string
		wantFlagMetadata of.FlagMetadata
	}{
		{
			name: "ProviderNotReadyCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.ProviderNotReadyCode,
						ErrorMessage: "not ready",
					},
				},
			},
			wantErrorCode: of.ProviderNotReadyCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "ProviderFatalCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.ProviderFatalCode,
						ErrorMessage: "fatal error",
					},
				},
			},
			wantErrorCode: of.ProviderFatalCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "GeneralCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.GeneralCode,
						ErrorMessage: "general error",
					},
				},
			},
			wantErrorCode: of.GeneralCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "FlagNotFoundCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.FlagNotFoundCode,
						ErrorMessage: "flag not found",
					},
				},
			},
			wantErrorCode: of.FlagNotFoundCode,
			wantReason:    of.DefaultReason,
			wantErr:       true,
		},
		{
			name: "TargetingKeyMissingCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.TargetingKeyMissingCode,
						ErrorMessage: "targeting key missing",
					},
				},
			},
			wantErrorCode: of.TargetingKeyMissingCode,
			wantReason:    of.TargetingMatchReason,
			wantErr:       true,
		},
		{
			name: "TypeMismatchCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.TypeMismatchCode,
						ErrorMessage: "type mismatch",
					},
				},
			},
			wantErrorCode: of.TypeMismatchCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "ParseErrorCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.ParseErrorCode,
						ErrorMessage: "parse error",
					},
				},
			},
			wantErrorCode: of.ParseErrorCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "InvalidContextCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.InvalidContextCode,
						ErrorMessage: "invalid context",
					},
				},
			},
			wantErrorCode: of.InvalidContextCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "unknown non-empty error code defaults to GeneralCode",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.ErrorCode("UNKNOWN_CODE"),
						ErrorMessage: "some unknown error",
					},
				},
			},
			wantErrorCode: of.GeneralCode,
			wantReason:    of.ErrorReason,
			wantErr:       true,
		},
		{
			name: "empty error code produces no error",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode: "",
					},
				},
			},
			wantErrorCode: "",
			wantReason:    "",
			wantErr:       false,
		},
		{
			name: "variant and flag metadata are preserved",
			input: of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{
						ErrorCode:    of.ProviderFatalCode,
						ErrorMessage: "fatal",
						Variant:      "variant-a",
						FlagMetadata: of.FlagMetadata{"key": "value"},
					},
				},
			},
			wantErrorCode:    of.ProviderFatalCode,
			wantReason:       of.ErrorReason,
			wantErr:          true,
			wantVariant:      "variant-a",
			wantFlagMetadata: of.FlagMetadata{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProviderResolutionDetail(tt.input)
			assert.Equal(t, tt.wantReason, result.Reason)
			assert.Equal(t, tt.wantVariant, result.Variant)
			assert.Equal(t, tt.wantFlagMetadata, result.FlagMetadata)

			err := result.Error()
			if tt.wantErr {
				require.Error(t, err)
				detail := result.ResolutionDetail()
				assert.Equal(t, tt.wantErrorCode, detail.ErrorCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newTestOnlyFirstProviderStrategy(providers []NamedProvider) StrategyFn[FlagTypes] {
	return newTestOnlyFirstProviderStrategyFn[FlagTypes](providers)
}

func newTestOnlyFirstProviderStrategyFn[T FlagTypes](providers []NamedProvider) StrategyFn[T] {
	return func(ctx context.Context, flag string, defaultValue T, flatCtx of.FlattenedContext) of.GenericResolutionDetail[T] {
		return Evaluate(ctx, providers[0], flag, defaultValue, flatCtx)
	}
}

func Test_HookIsolator_Evaluation_ProviderNotReadyCode(t *testing.T) {
	done := make(chan struct{})
	api := isolated.NewAPI()
	t.Cleanup(func() {
		<-done
		_ = api.Shutdown(context.Background()) //nolint:usetesting
	})

	p := struct {
		of.FeatureProvider
		of.StateHandler
	}{
		FeatureProvider: of.NoopProvider{},
		StateHandler: stateHandler{
			initF: func(e of.EvaluationContext) error {
				<-t.Context().Done()
				close(done)
				return nil
			},
		},
	}

	mp, err := NewProvider(
		StrategyCustom,
		WithProvider("never-ready", p),
		WithCustomStrategy(newTestOnlyFirstProviderStrategy),
	)
	require.NoError(t, err)
	err = api.SetProvider(t.Context(), mp)
	require.NoError(t, err)

	client := api.NewClient()
	result, err := client.BooleanValueDetails(t.Context(), "test-flag", false, of.NewEvaluationContext("test", nil))
	require.Error(t, err)
	assert.False(t, result.Value)
	assert.Equal(t, of.ProviderNotReadyCode, result.ErrorCode)
	assert.Equal(t, of.ErrorReason, result.Reason)
}

func Test_HookIsolator_Evaluation_ProviderFatalCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	provider.EXPECT().Metadata().Return(of.Metadata{Name: t.Name()}).MinTimes(1)
	provider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(of.BoolResolutionDetail{
		Value: false,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			ResolutionError: of.NewProviderFatalResolutionError("fatal"),
			Reason:          of.ErrorReason,
		},
	})

	api := isolated.NewAPI()
	t.Cleanup(func() {
		_ = api.Shutdown(context.Background()) //nolint:usetesting
	})
	mp, err := NewProvider(
		StrategyCustom,
		WithProvider("fatal", provider),
		WithCustomStrategy(newTestOnlyFirstProviderStrategy),
	)
	require.NoError(t, err)
	err = api.SetProviderAndWait(t.Context(), mp)
	require.NoError(t, err)

	client := api.NewClient()
	result, err := client.BooleanValueDetails(t.Context(), "test-flag", false, of.NewEvaluationContext("test", nil))
	require.Error(t, err)
	assert.False(t, result.Value)
	assert.Equal(t, of.ProviderFatalCode, result.ErrorCode)
	assert.Equal(t, of.ErrorReason, result.Reason)
}

func Test_HookIsolator_ExecutesHooksDuringEvaluation_WithAfterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	testHook := of.NewMockHook(ctrl)
	testHook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	testHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test error"))
	testHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	testHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{testHook})
	provider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(of.BoolResolutionDetail{
		Value:                    false,
		ProviderResolutionDetail: of.ProviderResolutionDetail{},
	})

	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, nil)
	result := isolator.BooleanEvaluation(t.Context(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.False(t, result.Value)
}
