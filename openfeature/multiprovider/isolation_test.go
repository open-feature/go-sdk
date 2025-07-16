package multiprovider

import (
	"context"
	"errors"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
	isolator := IsolateProvider(provider, []of.Hook{})
	assert.Zero(t, isolator.capturedContext)
	assert.Zero(t, isolator.capturedHints)
	evalCtx, err := isolator.Before(context.Background(), hookCtx, hookHints)
	require.NoError(t, err)
	assert.NotNil(t, evalCtx)
	assert.Equal(t, hookCtx, isolator.capturedContext)
	assert.Equal(t, hookHints, isolator.capturedHints)
}

func Test_HookIsolator_Hooks_ReturnsSelf(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	isolator := IsolateProvider(provider, []of.Hook{})
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

	isolator := IsolateProvider(provider, nil)
	result := isolator.BooleanEvaluation(context.Background(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
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

	isolator := IsolateProvider(provider, nil)
	result := isolator.BooleanEvaluation(context.Background(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.False(t, result.Value)
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

	isolator := IsolateProvider(provider, nil)
	result := isolator.BooleanEvaluation(context.Background(), "test-flag", false, of.FlattenedContext{"targetingKey": "anon"})
	assert.False(t, result.Value)
}
