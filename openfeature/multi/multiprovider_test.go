package multi

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

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

	mp, err := NewProvider(providers, StrategyFirstSuccess)
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
		_, err := NewProvider(nil, StrategyFirstMatch)
		require.Errorf(t, err, "providerMap cannot be nil or empty")
	})

	t.Run("naming a provider the empty string returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers[""] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewProvider(providers, StrategyFirstMatch)
		require.Errorf(t, err, "provider name cannot be the empty string")
	})

	t.Run("nil provider within map returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = nil
		_, err := NewProvider(providers, StrategyFirstMatch)
		require.Errorf(t, err, "provider provider1 cannot be nil")
	})

	t.Run("unknown evaluation strategyFunc returns an error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewProvider(providers, "unknown")
		require.Errorf(t, err, "unknown is an unknown evaluation strategyFunc")
	})

	t.Run("setting custom strategyFunc without custom strategyFunc option returns error", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		_, err := NewProvider(providers, StrategyCustom)
		require.Errorf(t, err, "A custom strategyFunc must be set via an option if StrategyCustom is set")
	})

	t.Run("success", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		mp, err := NewProvider(providers, StrategyComparison)
		require.NoError(t, err)
		assert.NotZero(t, mp)
	})

	t.Run("success with custom provider", func(t *testing.T) {
		providers := make(ProviderMap)
		providers["provider1"] = imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})
		mp, err := NewProvider(providers, StrategyCustom, WithCustomStrategy(func(providers []*NamedProvider) StrategyFn[FlagTypes] {
			return func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
				return of.GenericResolutionDetail[FlagTypes]{
					Value:                    defaultValue,
					ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.UnknownReason},
				}
			}
		}))
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

	mp, err := NewProvider(providers, StrategyFirstMatch)
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

		mp, err := NewProvider(providers, StrategyFirstSuccess)
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

		mp, err := NewProvider(providers, StrategyFirstSuccess)
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

	mp, err := NewProvider(providers, StrategyFirstMatch)
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

	mp, err := NewProvider(providers, StrategyFirstMatch)
	require.NoError(t, err)

	attributes := map[string]any{
		"foo": "bar",
	}
	evalCtx := of.NewTargetlessEvaluationContext(attributes)
	err = mp.Init(evalCtx)
	require.Errorf(t, err, "Provider provider3: test error")
	assert.Equal(t, of.ErrorState, mp.overallStatus)
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
	mp, err := NewProvider(providers, StrategyFirstMatch)
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
	mp, err := NewProvider(providers, StrategyFirstMatch)
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

func TestMultiProvider_statusEvaluation(t *testing.T) {
	multiProvider := &Provider{
		overallStatus:  of.NotReadyState,
		providerStatus: make(map[string]of.State),
	}

	t.Run("empty state is ready", func(t *testing.T) {
		assert.Equal(t, of.ReadyState, multiProvider.evaluateState())
	})

	t.Run("all states ready is ready", func(t *testing.T) {
		multiProvider.providerStatus["provider1"] = of.ReadyState
		multiProvider.providerStatus["provider2"] = of.ReadyState
		multiProvider.providerStatus["provider3"] = of.ReadyState
		assert.Equal(t, of.ReadyState, multiProvider.evaluateState())
	})

	t.Run("one state stale is stale", func(t *testing.T) {
		multiProvider.providerStatus["provider1"] = of.ReadyState
		multiProvider.providerStatus["provider2"] = of.ReadyState
		multiProvider.providerStatus["provider3"] = of.StaleState
		assert.Equal(t, of.StaleState, multiProvider.evaluateState())
	})

	t.Run("one state error is error", func(t *testing.T) {
		multiProvider.providerStatus["provider1"] = of.ReadyState
		multiProvider.providerStatus["provider2"] = of.StaleState
		multiProvider.providerStatus["provider3"] = of.ErrorState
		assert.Equal(t, of.ErrorState, multiProvider.evaluateState())
	})
}

func TestMultiProvider_StateUpdateWithSameTypeProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Create two mock providers with EventHandler support
	primaryProvider := newMockProviderWithEvents(ctrl, "MockProvider")
	secondaryProvider := newMockProviderWithEvents(ctrl, "MockProvider")

	providers := ProviderMap{
		"primary":   primaryProvider,
		"secondary": secondaryProvider,
	}

	multiProvider, err := NewProvider(providers, StrategyFirstMatch)
	if err != nil {
		t.Fatalf("failed to create multi-provider: %v", err)
	}
	t.Cleanup(multiProvider.Shutdown)

	// Initialize the provider
	ctx := of.NewEvaluationContext("test", nil)
	if err := multiProvider.Init(ctx); err != nil {
		t.Fatalf("failed to initialize multi-provider: %v", err)
	}

	primaryProvider.EmitEvent(of.ProviderError, "fail to fetch data")
	secondaryProvider.EmitEvent(of.ProviderReady, "rev 1")

	time.Sleep(200 * time.Millisecond)

	// Check the state after the error event
	multiProvider.providerStatusLock.Lock()
	primaryState := multiProvider.providerStatus["primary"]
	secondaryState := multiProvider.providerStatus["secondary"]
	numProviders := len(multiProvider.providerStatus)
	multiProvider.providerStatusLock.Unlock()

	if primaryState != of.ErrorState {
		t.Errorf("Expected primary-mock state to be ERROR after emitting error event, got %s", primaryState)
	}

	if secondaryState != of.ReadyState {
		t.Errorf("Expected secondary-mock state to be READY, got %s", secondaryState)
	}

	if numProviders != 2 {
		t.Errorf("Expected 2 providers in status map, got %d", numProviders)
	}
}

var _ of.StateHandler = (*mockProviderWithEvents)(nil)

// mockProviderWithEvents wraps a mock provider to add EventHandler capability
type mockProviderWithEvents struct {
	*of.MockFeatureProvider
	*of.MockStateHandler
	eventChannel chan of.Event
	metadata     of.Metadata
}

func newMockProviderWithEvents(ctrl *gomock.Controller, name string) *mockProviderWithEvents {
	mockProvider := of.NewMockFeatureProvider(ctrl)
	mockStateHandler := of.NewMockStateHandler(ctrl)
	eventChan := make(chan of.Event, 10)

	metadata := of.Metadata{Name: name}

	// Set up expectations
	mockProvider.EXPECT().Metadata().Return(metadata).AnyTimes()
	mockProvider.EXPECT().Hooks().Return([]of.Hook{}).AnyTimes()
	mockStateHandler.EXPECT().Init(gomock.Any()).DoAndReturn(func(ctx of.EvaluationContext) error {
		// Emit READY event on init
		eventChan <- of.Event{
			ProviderName: name,
			EventType:    of.ProviderReady,
			ProviderEventDetails: of.ProviderEventDetails{
				EventMetadata: make(map[string]any),
			},
		}
		return nil
	}).AnyTimes()
	mockStateHandler.EXPECT().Shutdown()

	return &mockProviderWithEvents{
		MockFeatureProvider: mockProvider,
		MockStateHandler:    mockStateHandler,
		eventChannel:        eventChan,
		metadata:            metadata,
	}
}

func (m *mockProviderWithEvents) Init(evalCtx of.EvaluationContext) error {
	return m.MockStateHandler.Init(evalCtx)
}

func (m *mockProviderWithEvents) Shutdown() {
	m.MockStateHandler.Shutdown()
	close(m.eventChannel)
}

func (m *mockProviderWithEvents) EventChannel() <-chan of.Event {
	return m.eventChannel
}

func (m *mockProviderWithEvents) EmitEvent(eventType of.EventType, message string) {
	m.eventChannel <- of.Event{
		ProviderName: m.metadata.Name,
		EventType:    eventType,
		ProviderEventDetails: of.ProviderEventDetails{
			Message:       message,
			EventMetadata: make(map[string]any),
		},
	}
}
