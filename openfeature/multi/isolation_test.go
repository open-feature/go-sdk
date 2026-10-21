package multi

import (
	"errors"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_HookIsolator_Hooks_ReturnsNoClientHooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, []of.Hook{})
	hooks := isolator.Hooks()
	assert.Empty(t, hooks)
	assert.Equal(t, "test-provider", isolator.Name())
}

func Test_HookIsolator_ExecutesHooksDuringEvaluation_NoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	testHook := of.NewMockHook(ctrl)
	testHook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	testHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	testHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	testHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(of.Metadata{Name: "test-provider"})
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
	provider.EXPECT().Metadata().Return(of.Metadata{Name: "test-provider"})
	provider.EXPECT().Hooks().Return([]of.Hook{testHook})

	isolator := isolateProvider(&namedProvider{
		FeatureProvider: provider,
		name:            "test-provider",
	}, nil)
	result := isolator.BooleanEvaluation(t.Context(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.False(t, result.Value)
	// empty record rather than nil, see #542
	assert.Equal(t, of.FlagMetadata{}, result.FlagMetadata)
}

func Test_HookIsolator_ExecutesHooksDuringEvaluation_WithAfterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	testHook := of.NewMockHook(ctrl)
	testHook.EXPECT().Before(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	testHook.EXPECT().After(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test error"))
	testHook.EXPECT().Finally(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	testHook.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(of.Metadata{Name: "test-provider"})
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

func Test_toProviderResolutionDetail_ErrorCodesMapToErrorReason(t *testing.T) {
	cases := []struct {
		name string
		code of.ErrorCode
	}{
		{"general", of.GeneralCode},
		{"flag not found", of.FlagNotFoundCode},
		{"targeting key missing", of.TargetingKeyMissingCode},
		{"type mismatch", of.TypeMismatchCode},
		{"parse error", of.ParseErrorCode},
		{"invalid context", of.InvalidContextCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prd := toProviderResolutionDetail(of.InterfaceEvaluationDetails{
				EvaluationDetails: of.EvaluationDetails{
					ResolutionDetail: of.ResolutionDetail{ErrorCode: tc.code},
				},
			})
			assert.Equal(t, of.ErrorReason, prd.Reason)
			assert.Error(t, prd.Error())
		})
	}
}
