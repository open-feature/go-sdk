package multiprovider

import (
	"context"
	"errors"
	"regexp"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	imp "github.com/open-feature/go-sdk/openfeature/memprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMultiProvider_ProvidersMethod(t *testing.T) {
	testProvider1 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
	testProvider2 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2

	mp, err := NewMultiProvider(providers, StrategyFirstSuccess)
	require.NoError(t, err)

	p := mp.Providers()
	assert.Len(t, p, 2)
	assert.Regexp(t, regexp.MustCompile("provider[1-2]"), p[0].Name)
	assert.NotNil(t, p[0].FeatureProvider)
	assert.Regexp(t, regexp.MustCompile("provider[1-2]"), p[1].Name)
	assert.NotNil(t, p[1].FeatureProvider)
}

func TestMultiProvider_NewMultiProvider(t *testing.T) {
	t.Run("nil providerMap returns an error", func(t *testing.T) {
		_, err := NewMultiProvider(nil, StrategyFirstMatch)
		require.Errorf(t, err, "providerMap cannot be nil or empty")
	})

	t.Run("naming a provider the empty string returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers[""] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewMultiProvider(providers, StrategyFirstMatch)
		require.Errorf(t, err, "provider name cannot be the empty string")
	})

	t.Run("nil provider within map returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = nil
		_, err := NewMultiProvider(providers, StrategyFirstMatch)
		require.Errorf(t, err, "provider provider1 cannot be nil")
	})

	t.Run("unknown evaluation strategy returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewMultiProvider(providers, "unknown")
		require.Errorf(t, err, "unknown is an unknown evaluation strategy")
	})

	t.Run("setting custom strategy without custom strategy option returns error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewMultiProvider(providers, StrategyCustom)
		require.Errorf(t, err, "A custom strategy must be set via an option if StrategyCustom is set")
	})

	t.Run("success", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		mp, err := NewMultiProvider(providers, StrategyComparison)
		require.NoError(t, err)
		assert.NotZero(t, mp)
	})

	t.Run("success with custom provider", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		strategy := func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
			return of.GenericResolutionDetail[FlagTypes]{
				Value:                    defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.UnknownReason},
			}
		}
		mp, err := NewMultiProvider(providers, StrategyCustom, WithCustomStrategy(strategy))
		require.NoError(t, err)
		assert.NotZero(t, mp)
	})
}

func TestMultiProvider_ProvidersByNamesMethod(t *testing.T) {
	testProvider1 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
	testProvider2 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2

	mp, err := NewMultiProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)

	p := mp.ProvidersByName()

	assert.Len(t, p, 2)
	require.Contains(t, p, "provider1")
	assert.Equal(t, p["provider1"], testProvider1)
	require.Contains(t, p, "provider2")
	assert.Equal(t, p["provider2"], testProvider2)
}

func TestMultiProvider_MetaData(t *testing.T) {
	t.Run("two providers", func(t *testing.T) {
		testProvider1 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		ctrl := gomock.NewController(t)
		testProvider2 := of.NewMockFeatureProvider(ctrl)
		testProvider2.EXPECT().Metadata().Return(of.Metadata{
			Name: "MockProvider",
		})
		testProvider2.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)

		providers := make(ProviderMap)
		providers["provider1"] = testProvider1
		providers["provider2"] = testProvider2

		mp, err := NewMultiProvider(providers, StrategyFirstSuccess)
		require.NoError(t, err)

		metadata := mp.Metadata()
		require.NotZero(t, metadata)
		assert.Equal(t, "MultiProvider {provider1: InMemoryProvider, provider2: MockProvider}", metadata.Name)
	})

	t.Run("three providers", func(t *testing.T) {
		testProvider1 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		ctrl := gomock.NewController(t)
		testProvider2 := of.NewMockFeatureProvider(ctrl)
		testProvider2.EXPECT().Metadata().Return(of.Metadata{
			Name: "MockProvider",
		})
		testProvider2.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
		testProvider3 := of.NewMockFeatureProvider(ctrl)
		testProvider3.EXPECT().Metadata().Return(of.Metadata{
			Name: "MockProvider",
		})
		testProvider3.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)

		providers := make(ProviderMap)
		providers["provider1"] = testProvider1
		providers["provider2"] = testProvider2
		providers["provider3"] = testProvider3

		mp, err := NewMultiProvider(providers, StrategyFirstSuccess)
		require.NoError(t, err)

		metadata := mp.Metadata()
		require.NotZero(t, metadata)
		assert.Equal(t, "MultiProvider {provider1: InMemoryProvider, provider2: MockProvider, provider3: MockProvider}", metadata.Name)
	})
}

func TestMultiProvider_Init(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctrl := gomock.NewController(t)

	testProvider1 := of.NewMockFeatureProvider(ctrl)
	testProvider1.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider1.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	initProvider := of.NewMockFeatureProvider(ctrl)
	initProvider.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	initProvider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	initHandler := of.NewMockStateHandler(ctrl)
	initHandler.EXPECT().Init(gomock.Any()).Return(nil)
	initHandler.EXPECT().Shutdown().MaxTimes(1)
	testProvider2 := struct {
		of.FeatureProvider
		of.StateHandler
	}{
		initProvider,
		initHandler,
	}
	testProvider3 := of.NewMockFeatureProvider(ctrl)
	testProvider3.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider3.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2
	providers["provider3"] = testProvider3

	mp, err := NewMultiProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)

	t.Cleanup(func() {
		mp.Shutdown()
	})

	attributes := map[string]any{
		"foo": "bar",
	}
	evalCtx := of.NewTargetlessEvaluationContext(attributes)
	err = mp.Init(evalCtx)
	require.NoError(t, err)
	assert.Equal(t, of.ReadyState, mp.Status())
}

func TestMultiProvider_InitErrorWithProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	errProvider := of.NewMockFeatureProvider(ctrl)
	errProvider.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	errProvider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	errHandler := of.NewMockStateHandler(ctrl)
	errHandler.EXPECT().Init(gomock.Any()).Return(errors.New("test error"))
	testProvider3 := struct {
		of.FeatureProvider
		of.StateHandler
	}{
		errProvider,
		errHandler,
	}

	testProvider1 := of.NewMockFeatureProvider(ctrl)
	testProvider1.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	testProvider1.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider2 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2
	providers["provider3"] = testProvider3

	mp, err := NewMultiProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)

	attributes := map[string]any{
		"foo": "bar",
	}
	evalCtx := of.NewTargetlessEvaluationContext(attributes)
	err = mp.Init(evalCtx)
	require.Errorf(t, err, "Provider provider3: test error")
	assert.Equal(t, of.ErrorState, mp.totalStatus)
}

func TestMultiProvider_Shutdown_WithoutInit(t *testing.T) {
	ctrl := gomock.NewController(t)

	testProvider1 := of.NewMockFeatureProvider(ctrl)
	testProvider1.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider1.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	testProvider2 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
	testProvider3 := of.NewMockFeatureProvider(ctrl)
	testProvider3.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider3.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2
	providers["provider3"] = testProvider3
	mp, err := NewMultiProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)

	mp.Shutdown()
}

func TestMultiProvider_Shutdown_WithInit(t *testing.T) {
	ctrl := gomock.NewController(t)

	testProvider1 := of.NewMockFeatureProvider(ctrl)
	testProvider1.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	testProvider1.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	testProvider2 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
	handlingProvider := of.NewMockFeatureProvider(ctrl)
	handlingProvider.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	handlingProvider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	handledHandler := of.NewMockStateHandler(ctrl)
	handledHandler.EXPECT().Init(gomock.Any()).Return(nil)
	handledHandler.EXPECT().Shutdown()
	testProvider3 := struct {
		of.FeatureProvider
		of.StateHandler
	}{
		handlingProvider,
		handledHandler,
	}

	providers := make(ProviderMap)
	providers["provider1"] = testProvider1
	providers["provider2"] = testProvider2
	providers["provider3"] = testProvider3
	mp, err := NewMultiProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)
	evalCtx := of.NewTargetlessEvaluationContext(map[string]any{
		"foo": "bar",
	})
	eventChan := make(chan of.Event)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case e := <-mp.EventChannel():
			eventChan <- e
		case <-ctx.Done():
			return
		}
	}()
	err = mp.Init(evalCtx)
	require.NoError(t, err)
	assert.Equal(t, of.ReadyState, mp.Status())
	cancel()
	mp.Shutdown()
	assert.Equal(t, of.NotReadyState, mp.Status())
}
