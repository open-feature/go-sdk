package multi

import (
	"context"
	"errors"
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

	mp, err := NewProvider(StrategyFirstSuccess, WithProvider("provider1", testProvider1), WithProvider("provider2", testProvider2))
	require.NoError(t, err)

	p := mp.Providers()
	assert.Len(t, p, 2)
	assert.NotNil(t, p[0])
	assert.Implements(t, (*of.FeatureProvider)(nil), p[0])
	assert.Equal(t, "provider1", p[0].Name())
	assert.NotNil(t, p[1])
	assert.Implements(t, (*of.FeatureProvider)(nil), p[1])
	assert.Equal(t, "provider2", p[1].Name())
}

func TestMultiProvider_NewMultiProvider(t *testing.T) {
	t.Run("nil providerMap returns an error", func(t *testing.T) {
		_, err := NewProvider(StrategyFirstMatch)
		require.Errorf(t, err, "providerMap cannot be nil or empty")
	})

	t.Run("naming a provider the empty string returns an error", func(t *testing.T) {
		_, err := NewProvider(StrategyFirstMatch, WithProvider("", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
		require.Errorf(t, err, "provider name cannot be the empty string")
	})

	t.Run("nil provider within map returns an error", func(t *testing.T) {
		_, err := NewProvider(StrategyFirstMatch, WithProvider("provider1", nil))
		require.Errorf(t, err, "provider provider1 cannot be nil")
	})

	t.Run("unknown evaluation strategyFunc returns an error", func(t *testing.T) {
		_, err := NewProvider("unknown", WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
		require.Errorf(t, err, "unknown is an unknown evaluation strategyFunc")
	})

	t.Run("setting custom strategyFunc without custom strategyFunc option returns error", func(t *testing.T) {
		_, err := NewProvider(StrategyCustom, WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
		require.Errorf(t, err, "A custom strategyFunc must be set via an option if StrategyCustom is set")
	})

	t.Run("success", func(t *testing.T) {
		mp, err := NewProvider(StrategyComparison, WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
		require.NoError(t, err)
		assert.NotZero(t, mp)
	})

	t.Run("success with custom provider", func(t *testing.T) {
		mp, err := NewProvider(StrategyCustom, WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
			return func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
				return of.GenericResolutionDetail[FlagTypes]{
					Value:                    defaultValue,
					ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.UnknownReason},
				}
			}
		}),
			WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})),
		)
		require.NoError(t, err)
		assert.NotZero(t, mp)
	})
}

// A custom strategy may return a zero-value result, and the typed accessors asserted on it
// without checking, so an unset Value took the calling process down.
func TestMultiProvider_TypedAccessorsWithUnsetStrategyValue(t *testing.T) {
	mp, err := NewProvider(StrategyCustom, WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
		return func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
			return of.GenericResolutionDetail[FlagTypes]{}
		}
	}),
		WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})),
	)
	require.NoError(t, err)

	assert.True(t, mp.BooleanEvaluation(t.Context(), "flag", true, of.FlattenedContext{}).Value)
	assert.Equal(t, "fallback", mp.StringEvaluation(t.Context(), "flag", "fallback", of.FlattenedContext{}).Value)
	assert.Equal(t, 1.5, mp.FloatEvaluation(t.Context(), "flag", 1.5, of.FlattenedContext{}).Value)
	assert.Equal(t, int64(7), mp.IntEvaluation(t.Context(), "flag", 7, of.FlattenedContext{}).Value)
}

// An inner provider may resolve an object flag to a nil value, and nil has no type to assert
// against, so the strategy's object branch panicked rather than resolving.
func TestMultiProvider_ObjectEvaluationWithNilProviderValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(of.Metadata{Name: "nil-object"}).AnyTimes()
	provider.EXPECT().Hooks().Return(nil).AnyTimes()
	provider.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(of.InterfaceResolutionDetail{
			ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.StaticReason},
		})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("nil-object", provider))
	require.NoError(t, err)

	defaultValue := map[string]any{"enabled": true}
	res := mp.ObjectEvaluation(t.Context(), "flag", defaultValue, of.FlattenedContext{})
	assert.Equal(t, defaultValue, res.Value)
}

// The object accessor asserted nothing at all, so a strategy that resolved no value
// returned nil to the caller in place of their default.
func TestMultiProvider_ObjectEvaluationWithUnsetStrategyValue(t *testing.T) {
	mp, err := NewProvider(StrategyCustom, WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
		return func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
			return of.GenericResolutionDetail[FlagTypes]{}
		}
	}),
		WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})),
	)
	require.NoError(t, err)

	defaultValue := map[string]any{"enabled": true}
	assert.Equal(t, defaultValue, mp.ObjectEvaluation(t.Context(), "flag", defaultValue, of.FlattenedContext{}).Value)
}

// A strategy that resolves the wrong type is a bug in the strategy, and reporting it as a
// successful resolution of the default leaves the caller nothing to diagnose it with.
func TestMultiProvider_TypedAccessorsWithWrongTypedStrategyValue(t *testing.T) {
	mp, err := NewProvider(StrategyCustom, WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
		return func(ctx context.Context, flag string, defaultValue FlagTypes, evalCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
			return of.GenericResolutionDetail[FlagTypes]{
				Value:                    "not-a-bool",
				ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.StaticReason},
			}
		}
	}),
		WithProvider("provider1", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})),
	)
	require.NoError(t, err)

	res := mp.BooleanEvaluation(t.Context(), "flag", true, of.FlattenedContext{})
	assert.True(t, res.Value)
	assert.Equal(t, of.ErrorReason, res.Reason)
	assert.Equal(t,
		of.NewTypeMismatchResolutionError("strategy resolved string, expected bool").Error(),
		res.ResolutionError.Error())
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

		mp, err := NewProvider(
			StrategyFirstSuccess,
			WithProvider("provider1", testProvider1),
			WithProvider("provider2", testProvider2),
		)
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

		mp, err := NewProvider(
			StrategyFirstSuccess,
			WithProvider("provider1", testProvider1),
			WithProvider("provider2", testProvider2),
			WithProvider("provider3", testProvider3),
		)
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

	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("provider1", testProvider1),
		WithProvider("provider2", testProvider2),
		WithProvider("provider3", testProvider3),
	)
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

func TestMultiProvider_InitReportsNotReady(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := of.NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	stateHandler := of.NewMockStateHandler(ctrl)
	started := make(chan struct{})
	release := make(chan struct{})
	stateHandler.EXPECT().Init(gomock.Any()).DoAndReturn(func(of.EvaluationContext) error {
		close(started)
		<-release
		return nil
	})
	stateHandler.EXPECT().Shutdown().MaxTimes(1)
	wrapped := struct {
		of.FeatureProvider
		of.StateHandler
	}{provider, stateHandler}

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", wrapped))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	initDone := make(chan error, 1)
	go func() {
		initDone <- mp.InitWithContext(t.Context(), of.EvaluationContext{})
	}()
	<-started
	assert.Equal(t, of.NotReadyState, mp.Status())
	close(release)
	require.NoError(t, <-initDone)
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

	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("provider1", testProvider1),
		WithProvider("provider2", testProvider2),
		WithProvider("provider3", testProvider3),
	)
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

	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("provider1", testProvider1),
		WithProvider("provider2", testProvider2),
		WithProvider("provider3", testProvider3),
	)
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

	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("provider1", testProvider1),
		WithProvider("provider2", testProvider2),
		WithProvider("provider3", testProvider3),
	)
	require.NoError(t, err)
	evalCtx := of.NewTargetlessEvaluationContext(map[string]any{
		"foo": "bar",
	})
	eventChan := make(chan of.Event)
	ctx, cancel := context.WithCancel(t.Context())
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

	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("primary", primaryProvider),
		WithProvider("secondary", secondaryProvider),
	)
	if err != nil {
		t.Fatalf("failed to create multi-provider: %v", err)
	}
	t.Cleanup(mp.Shutdown)

	// Initialize the provider
	ctx := of.NewEvaluationContext("test", nil)
	if err := mp.Init(ctx); err != nil {
		t.Fatalf("failed to initialize multi-provider: %v", err)
	}

	primaryProvider.EmitEvent(of.ProviderError, "fail to fetch data")
	secondaryProvider.EmitEvent(of.ProviderReady, "rev 1")
	// wait for processing
	require.Eventually(t, func() bool {
		select {
		case <-mp.outboundEvents:
			return true
		default:
			return false
		}
	}, time.Second, 50*time.Millisecond, "expected event was not emitted within timeout")

	// Check the state after the error event
	mp.providerStatusLock.Lock()
	primaryState := mp.providerStatus["primary"]
	secondaryState := mp.providerStatus["secondary"]
	numProviders := len(mp.providerStatus)
	mp.providerStatusLock.Unlock()

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

func TestMultiProvider_FatalEventUpdatesStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := newMockProviderWithEvents(ctrl, "fatal-provider")
	mp, err := NewProvider(StrategyFirstMatch, WithProvider("fatal-provider", provider))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)
	require.NoError(t, mp.Init(of.EvaluationContext{}))

	provider.eventChannel <- of.Event{
		EventType: of.ProviderError,
		ProviderEventDetails: of.ProviderEventDetails{
			ErrorCode:     of.ProviderFatalCode,
			EventMetadata: map[string]any{},
		},
	}

	require.Eventually(t, func() bool {
		return mp.Status() == of.FatalState
	}, time.Second, 10*time.Millisecond)
}

func TestMultiProvider_ConfigurationChangedEventForwarding(t *testing.T) {
	// awaitEvent drains the outbound channel until an event of the given type is found.
	awaitEvent := func(t *testing.T, ch <-chan of.Event, eventType of.EventType) of.Event {
		t.Helper()
		var found of.Event
		require.Eventually(t, func() bool {
			for {
				select {
				case e, ok := <-ch:
					if !ok {
						return false
					}
					if e.EventType == eventType {
						found = e
						return true
					}
				default:
					return false
				}
			}
		}, time.Second, 10*time.Millisecond, "expected %s event was not received", eventType)
		return found
	}

	// setup creates a two-provider multi-provider, initializes it, and waits for READY.
	setup := func(t *testing.T) (*Provider, *mockProviderWithEvents, *mockProviderWithEvents) {
		t.Helper()
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		p1 := newMockProviderWithEvents(ctrl, "provider1")
		p2 := newMockProviderWithEvents(ctrl, "provider2")

		mp, err := NewProvider(
			StrategyFirstMatch,
			WithProvider("provider1", p1),
			WithProvider("provider2", p2),
		)
		require.NoError(t, err)
		t.Cleanup(mp.Shutdown)

		require.NoError(t, mp.Init(of.NewEvaluationContext("test", nil)))
		require.Eventually(t, func() bool {
			return mp.Status() == of.ReadyState
		}, time.Second, 10*time.Millisecond)

		return mp, p1, p2
	}

	t.Run("event is forwarded with correct payload and does not corrupt provider state", func(t *testing.T) {
		mp, provider1, _ := setup(t)

		provider1.EmitEvent(of.ProviderConfigChange, "flags updated")
		e := awaitEvent(t, mp.outboundEvents, of.ProviderConfigChange)

		assert.Equal(t, "flags updated", e.Message)
		assert.Equal(t, "provider1", e.EventMetadata[MetadataProviderName])

		// Provider state should remain READY (not corrupted to empty string)
		mp.providerStatusLock.Lock()
		assert.Equal(t, of.ReadyState, mp.providerStatus["provider1"])
		assert.Equal(t, of.ReadyState, mp.providerStatus["provider2"])
		mp.providerStatusLock.Unlock()

		assert.Equal(t, of.ReadyState, mp.Status())
	})

	t.Run("does not affect aggregate state when another provider is degraded", func(t *testing.T) {
		mp, provider1, provider2 := setup(t)

		// Put provider2 into STALE state
		provider2.EmitEvent(of.ProviderStale, "stale")
		awaitEvent(t, mp.outboundEvents, of.ProviderStale)
		require.Equal(t, of.StaleState, mp.Status())

		// ConfigurationChanged from provider1 should not affect aggregate
		provider1.EmitEvent(of.ProviderConfigChange, "flags updated")
		awaitEvent(t, mp.outboundEvents, of.ProviderConfigChange)

		assert.Equal(t, of.StaleState, mp.Status())

		mp.providerStatusLock.Lock()
		assert.Equal(t, of.ReadyState, mp.providerStatus["provider1"])
		mp.providerStatusLock.Unlock()
	})
}

func TestMultiProvider_Track(t *testing.T) {
	t.Run("forwards tracking to all ready providers that implement Tracker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		provider1 := newMockProviderWithEvents(ctrl, "provider1")
		provider2 := newMockProviderWithEvents(ctrl, "provider2")
		provider3 := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{}) // Does not implement Tracker

		mp, err := NewProvider(
			StrategyFirstSuccess,
			WithProvider("provider1", provider1),
			WithProvider("provider2", provider2),
			WithProvider("provider3", provider3),
		)
		require.NoError(t, err)
		t.Cleanup(mp.Shutdown)

		evalCtx := of.NewEvaluationContext("user-123", map[string]any{"plan": "premium"})
		err = mp.Init(evalCtx)
		require.NoError(t, err)

		trackingEventName := "button-clicked"
		details := of.NewTrackingEventDetails(42.0).Add("currency", "USD")

		ctx := t.Context()
		// Expect Track to be called on providers that implement Tracker
		provider1.MockTracker.EXPECT().Track(ctx, trackingEventName, evalCtx, details).Times(1)
		provider2.MockTracker.EXPECT().Track(ctx, trackingEventName, evalCtx, details).Times(1)

		mp.Track(ctx, trackingEventName, evalCtx, details)
	})

	t.Run("does not track when provider is not initialized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		provider1 := newMockProviderWithEvents(ctrl, "provider1")
		// manual shutdown on cleanup because multi-provider won't be initialized
		t.Cleanup(provider1.Shutdown)

		mp, err := NewProvider(StrategyFirstSuccess, WithProvider("provider1", provider1))
		require.NoError(t, err)
		t.Cleanup(mp.Shutdown)

		// Don't initialize the multi-provider
		ctx := t.Context()
		trackingEventName := "button-clicked"
		evalCtx := of.NewEvaluationContext("user-123", map[string]any{})
		details := of.TrackingEventDetails{}

		// Should not call Track on provider
		provider1.MockTracker.EXPECT().Track(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		mp.Track(ctx, trackingEventName, evalCtx, details)
	})

	t.Run("only tracks on providers in ready state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		readyProvider := newMockProviderWithEvents(ctrl, "ready-provider")
		errorProvider := newMockProviderWithEvents(ctrl, "error-provider")

		mp, err := NewProvider(
			StrategyFirstSuccess,
			WithProvider("ready-provider", readyProvider),
			WithProvider("error-provider", errorProvider),
		)
		require.NoError(t, err)
		t.Cleanup(mp.Shutdown)

		evalCtx := of.NewEvaluationContext("user-456", map[string]any{})
		err = mp.Init(evalCtx)
		require.NoError(t, err)

		// Simulate error state for one provider
		errorProvider.EmitEvent(of.ProviderError, "error")

		// wait for event processing
		require.Eventually(t, func() bool {
			select {
			case e := <-mp.outboundEvents:
				return e.ProviderName == "error-provider" && e.EventType == of.ProviderError
			default:
				return false
			}
		}, time.Second, 50*time.Millisecond, "expected event was not emitted within timeout")

		trackingEventName := "page-view"
		details := of.TrackingEventDetails{}

		ctx := t.Context()
		readyProvider.MockTracker.EXPECT().Track(ctx, trackingEventName, evalCtx, details).Times(1)
		errorProvider.MockTracker.EXPECT().Track(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		mp.Track(ctx, trackingEventName, evalCtx, details)
	})

	t.Run("handles providers that don't implement Tracker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		trackerProvider := newMockProviderWithEvents(ctrl, "tracker-provider")
		nonTrackerProvider := imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})

		mp, err := NewProvider(
			StrategyFirstSuccess,
			WithProvider("tracker-provider", trackerProvider),
			WithProvider("non-tracker", nonTrackerProvider),
		)
		require.NoError(t, err)
		t.Cleanup(mp.Shutdown)

		evalCtx := of.NewEvaluationContext("user-789", map[string]any{})
		err = mp.Init(evalCtx)
		require.NoError(t, err)

		trackingEventName := "conversion"
		details := of.NewTrackingEventDetails(99.99)

		ctx := t.Context()
		trackerProvider.MockTracker.EXPECT().Track(ctx, trackingEventName, evalCtx, details).Times(1)
		mp.Track(ctx, trackingEventName, evalCtx, details)
	})
}

var _ of.StateHandler = (*mockProviderWithEvents)(nil)

// mockProviderWithEvents wraps a mock provider to add EventHandler and optional Tracker capability
type mockProviderWithEvents struct {
	*of.MockFeatureProvider
	*of.MockStateHandler
	*of.MockTracker
	eventChannel chan of.Event
	metadata     of.Metadata
}

func newMockProviderWithEvents(ctrl *gomock.Controller, name string) *mockProviderWithEvents {
	mockProvider := of.NewMockFeatureProvider(ctrl)
	mockStateHandler := of.NewMockStateHandler(ctrl)
	mockTracker := of.NewMockTracker(ctrl)
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
		MockTracker:         mockTracker,
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

func (m *mockProviderWithEvents) Track(ctx context.Context, trackingEventName string, evaluationContext of.EvaluationContext, details of.TrackingEventDetails) {
	if m.MockTracker != nil {
		m.MockTracker.Track(ctx, trackingEventName, evaluationContext, details)
	}
}
