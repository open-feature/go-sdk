package multi

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
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

// eventingProvider implements both [of.StateHandler] and [of.EventHandler] with a scriptable
// initialization and an event channel the test drives, which is what the retry and recovery paths
// after a failed initialization need (#561).
type eventingProvider struct {
	of.FeatureProvider
	events    chan of.Event
	init      func() error
	initCalls atomic.Int32
	shutdowns atomic.Int32
}

var (
	_ of.StateHandler = (*eventingProvider)(nil)
	_ of.EventHandler = (*eventingProvider)(nil)
)

func newEventingProvider(ctrl *gomock.Controller, name string, init func() error) *eventingProvider {
	mock := of.NewMockFeatureProvider(ctrl)
	mock.EXPECT().Metadata().Return(of.Metadata{Name: name}).AnyTimes()
	mock.EXPECT().Hooks().Return([]of.Hook{}).AnyTimes()

	return &eventingProvider{
		FeatureProvider: mock,
		events:          make(chan of.Event, 10),
		init:            init,
	}
}

func (p *eventingProvider) Init(of.EvaluationContext) error {
	p.initCalls.Add(1)
	if p.init == nil {
		return nil
	}
	return p.init()
}

// Shutdown deliberately leaves the event channel open: the forwarder's listener exits on the
// worker context, and a retry has to be able to keep reading from the same provider.
func (p *eventingProvider) Shutdown() { p.shutdowns.Add(1) }

func (p *eventingProvider) EventChannel() <-chan of.Event { return p.events }

// contextAwareEventingProvider is an eventingProvider that also implements
// [of.ContextAwareStateHandler], so the multi-provider takes the ShutdownWithContext branch and
// the context handed to a provider during cleanup can be asserted on.
type contextAwareEventingProvider struct {
	*eventingProvider
	shutdownCtxErr chan error
}

var _ of.ContextAwareStateHandler = (*contextAwareEventingProvider)(nil)

func newContextAwareEventingProvider(ctrl *gomock.Controller, name string, init func() error) *contextAwareEventingProvider {
	return &contextAwareEventingProvider{
		eventingProvider: newEventingProvider(ctrl, name, init),
		shutdownCtxErr:   make(chan error, 1),
	}
}

func (p *contextAwareEventingProvider) InitWithContext(_ context.Context, evalCtx of.EvaluationContext) error {
	return p.Init(evalCtx)
}

func (p *contextAwareEventingProvider) ShutdownWithContext(ctx context.Context) error {
	p.shutdowns.Add(1)
	select {
	case p.shutdownCtxErr <- ctx.Err():
	default:
	}
	return nil
}

func (p *eventingProvider) emitReady() {
	p.events <- of.Event{
		ProviderName: p.Metadata().Name,
		EventType:    of.ProviderReady,
		ProviderEventDetails: of.ProviderEventDetails{
			EventMetadata: make(map[string]any),
		},
	}
}

func (p *eventingProvider) emitError() {
	p.events <- of.Event{
		ProviderName: p.Metadata().Name,
		EventType:    of.ProviderError,
		ProviderEventDetails: of.ProviderEventDetails{
			EventMetadata: make(map[string]any),
		},
	}
}

// TestMultiProvider_ShutdownAfterFailedInitCleansUp pins where cleanup happens after a failed
// initialization. Init leaves everything running, because a provider in ERROR can still recover
// (spec §1.7, 5.3.2); Shutdown is what releases EventChannel() consumers and shuts the inner
// providers down, and the SDK calls it regardless of state (1.6.1). See #561.
func TestMultiProvider_ShutdownAfterFailedInitCleansUp(t *testing.T) {
	ctrl := gomock.NewController(t)

	initialized := make(chan struct{})
	// Context-aware, so cleanup goes through ShutdownWithContext and the context it is handed
	// can be checked: the errgroup's context is already cancelled by the failure, and shutting a
	// provider down on a dead context would silently skip its teardown.
	healthy := newContextAwareEventingProvider(ctrl, "healthy", func() error {
		close(initialized)
		return nil
	})
	// Fail only once the healthy provider is up, so "it did initialize" is ordered, not raced.
	failing := newEventingProvider(ctrl, "failing", func() error {
		<-initialized
		return errors.New("injected init failure")
	})

	mp, err := NewProvider(StrategyFirstMatch,
		WithProvider("healthy", healthy),
		WithProvider("failing", failing))
	require.NoError(t, err)

	require.Error(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	assert.Equal(t, of.ErrorState, mp.Status())
	assert.Zero(t, healthy.shutdowns.Load(),
		"a failed init must leave the providers that did initialize running, so they can recover")

	released := make(chan bool, 1)
	go func() {
		_, open := <-mp.EventChannel()
		released <- open
	}()

	require.NoError(t, mp.ShutdownWithContext(t.Context()))

	select {
	case open := <-released:
		assert.False(t, open, "shutdown should close the outbound event channel")
	case <-time.After(time.Second):
		t.Fatal("a consumer of EventChannel() was never released by shutdown after a failed init")
	}
	assert.Equal(t, int32(1), healthy.shutdowns.Load(), "shutdown must reach the provider that initialized")
	assert.Equal(t, int32(1), failing.shutdowns.Load(), "shutdown must reach the provider whose init failed")
	assert.Equal(t, of.NotReadyState, mp.Status())

	select {
	case ctxErr := <-healthy.shutdownCtxErr:
		require.NoError(t, ctxErr, "a context-aware provider must be shut down on a live context")
	default:
		t.Fatal("the context-aware provider was never shut down")
	}
}

// TestMultiProvider_RetryAfterFailedInit covers re-initializing after a failure. The outbound
// channel is created once in NewProvider, so closing it on the failure path made this second
// attempt panic with "close of closed channel" (#561).
func TestMultiProvider_RetryAfterFailedInit(t *testing.T) {
	ctrl := gomock.NewController(t)

	var attempts atomic.Int32
	flaky := newEventingProvider(ctrl, "flaky", func() error {
		if attempts.Add(1) == 1 {
			return errors.New("injected init failure")
		}
		return nil
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("flaky", flaky))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	require.Error(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	assert.Equal(t, of.ErrorState, mp.Status())

	require.NoError(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	assert.Equal(t, of.ReadyState, mp.Status())
	assert.Equal(t, int32(2), flaky.initCalls.Load())
}

// TestMultiProvider_RecoveryAfterFailedInit covers a provider that recovers on its own after its
// initialization failed. Its event handler is registered before init can fail, so the READY it
// emits is still heard and the aggregate status follows it back (#561).
func TestMultiProvider_RecoveryAfterFailedInit(t *testing.T) {
	ctrl := gomock.NewController(t)

	failing := newEventingProvider(ctrl, "failing", func() error {
		return errors.New("injected init failure")
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("failing", failing))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	require.Error(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	assert.Equal(t, of.ErrorState, mp.Status())

	failing.emitReady()

	select {
	case e := <-mp.EventChannel():
		assert.Equal(t, of.ProviderReady, e.EventType)
	case <-time.After(time.Second):
		t.Fatal("a READY event from a recovered provider was never forwarded")
	}
	assert.Equal(t, of.ReadyState, mp.Status())
}

// emitConfigChange sends a ProviderConfigChange, which forwardProviderEvents passes through
// unconditionally — unlike a state event, it needs outbound buffer space every time.
func (p *eventingProvider) emitConfigChange() {
	p.events <- of.Event{
		ProviderName: p.Metadata().Name,
		EventType:    of.ProviderConfigChange,
		ProviderEventDetails: of.ProviderEventDetails{
			EventMetadata: make(map[string]any),
		},
	}
}

// fillOutboundBuffer pushes more pass-through events than the outbound channel can hold, with
// nothing reading EventChannel(), so the forwarder is parked on a send.
func fillOutboundBuffer(t *testing.T, p *eventingProvider, n int) {
	t.Helper()
	for range n {
		p.emitConfigChange()
	}
	time.Sleep(200 * time.Millisecond)
}

// TestMultiProvider_ShutdownDoesNotHangOnABlockedForwarder covers a forwarder parked on a send
// because nobody is reading EventChannel(). ShutdownWithContext waits on workerGroup, so an
// unguarded send there waits forever.
func TestMultiProvider_ShutdownDoesNotHangOnABlockedForwarder(t *testing.T) {
	ctrl := gomock.NewController(t)

	// A single provider, so the outbound buffer holds exactly one event.
	provider := newEventingProvider(ctrl, "provider", nil)

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)
	require.NoError(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))

	fillOutboundBuffer(t, provider, 4)

	done := make(chan error, 1)
	go func() { done <- mp.ShutdownWithContext(t.Context()) }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("ShutdownWithContext hung on a forwarder parked on an unguarded send")
	}
}

// TestMultiProvider_RetryDoesNotHangOnABlockedForwarder is the same hazard reached through
// InitWithContext, which waits on the previous attempt's forwarder before starting a new one.
func TestMultiProvider_RetryDoesNotHangOnABlockedForwarder(t *testing.T) {
	ctrl := gomock.NewController(t)

	var attempts atomic.Int32
	flaky := newEventingProvider(ctrl, "flaky", func() error {
		if attempts.Add(1) == 1 {
			return errors.New("injected init failure")
		}
		return nil
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("flaky", flaky))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	require.Error(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))

	fillOutboundBuffer(t, flaky, 4)

	done := make(chan error, 1)
	go func() { done <- mp.InitWithContext(t.Context(), of.EvaluationContext{}) }()
	select {
	case err := <-done:
		require.NoError(t, err)
		assert.Equal(t, of.ReadyState, mp.Status())
	case <-time.After(5 * time.Second):
		t.Fatal("a retry hung waiting on a forwarder parked on an unguarded send")
	}
}

// TestMultiProvider_ShutdownTwiceIsNoOp guards the close that moved into ShutdownWithContext: a
// second shutdown must neither shut the inner providers down again nor close the outbound channel
// twice (#561).
func TestMultiProvider_ShutdownTwiceIsNoOp(t *testing.T) {
	ctrl := gomock.NewController(t)

	provider := newEventingProvider(ctrl, "provider", nil)

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)

	require.NoError(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	require.NoError(t, mp.ShutdownWithContext(t.Context()))
	require.NoError(t, mp.ShutdownWithContext(t.Context()))

	assert.Equal(t, int32(1), provider.shutdowns.Load())
	assert.Equal(t, of.NotReadyState, mp.Status())
}

func TestMultiProvider_InitAndShutdownAreSerialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	started := make(chan struct{})
	release := make(chan struct{})
	provider := newEventingProvider(ctrl, "provider", func() error {
		close(started)
		<-release
		return nil
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)

	initDone := make(chan error, 1)
	go func() { initDone <- mp.InitWithContext(t.Context(), of.EvaluationContext{}) }()
	<-started

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- mp.ShutdownWithContext(t.Context()) }()
	select {
	case err := <-shutdownDone:
		require.NoError(t, err)
		t.Fatal("ShutdownWithContext returned while InitWithContext was still running")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	require.NoError(t, <-initDone)
	require.NoError(t, <-shutdownDone)
	assert.Equal(t, int32(1), provider.shutdowns.Load())
	assert.Equal(t, of.NotReadyState, mp.Status())
}

func TestMultiProvider_ShutdownLifecycleWaitObservesContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	started := make(chan struct{})
	release := make(chan struct{})
	provider := newEventingProvider(ctrl, "provider", func() error {
		close(started)
		<-release
		return nil
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)

	initDone := make(chan error, 1)
	go func() { initDone <- mp.InitWithContext(t.Context(), of.EvaluationContext{}) }()
	<-started

	ctx, cancel := context.WithCancel(t.Context())
	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- mp.ShutdownWithContext(ctx) }()
	cancel()
	select {
	case err := <-shutdownDone:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("ShutdownWithContext did not observe cancellation while waiting for initialization")
	}

	close(release)
	require.NoError(t, <-initDone)
	require.NoError(t, mp.ShutdownWithContext(t.Context()))
}

func TestMultiProvider_QueuedInitEventOverridesInitialStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	var provider *eventingProvider
	provider = newEventingProvider(ctrl, "provider", func() error {
		provider.emitError()
		return nil
	})

	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	require.NoError(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
	select {
	case event := <-mp.EventChannel():
		require.Equal(t, of.ProviderError, event.EventType)
	case <-time.After(time.Second):
		t.Fatal("queued provider error was never forwarded")
	}
	assert.Equal(t, of.ErrorState, mp.Status())
}

// TestMultiProvider_InitDoesNotReportReadyWhileProvidersRemain guards the seeding order in
// InitWithContext: the aggregate status must not leave NOT_READY while any provider is still
// initializing (#580).
func TestMultiProvider_InitDoesNotReportReadyWhileProvidersRemain(t *testing.T) {
	const fastProviders = 128

	ctrl := gomock.NewController(t)
	blocker := of.NewMockFeatureProvider(ctrl)
	blocker.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
	blocker.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
	blockerState := of.NewMockStateHandler(ctrl)
	started := make(chan struct{})
	release := make(chan struct{})
	blockerState.EXPECT().Init(gomock.Any()).DoAndReturn(func(of.EvaluationContext) error {
		close(started)
		<-release
		return nil
	})
	blockerState.EXPECT().Shutdown().MaxTimes(1)
	wrapped := struct {
		of.FeatureProvider
		of.StateHandler
	}{blocker, blockerState}

	opts := make([]Option, 0, fastProviders+1)
	for i := range fastProviders {
		opts = append(opts, WithProvider(fmt.Sprintf("fast-%d", i), imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
	}
	opts = append(opts, WithProvider("blocker", wrapped))

	mp, err := NewProvider(StrategyFirstMatch, opts...)
	require.NoError(t, err)
	t.Cleanup(mp.Shutdown)

	// The blocking provider never reaches ready, so only close(release) may end NOT_READY.
	violation := make(chan of.State, 1)
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		for {
			if s := mp.Status(); s != of.NotReadyState {
				select {
				case <-release:
				default:
					violation <- s
				}
				return
			}
			select {
			case <-release:
				return
			default:
			}
		}
	}()

	initDone := make(chan error, 1)
	go func() {
		initDone <- mp.InitWithContext(t.Context(), of.EvaluationContext{})
	}()

	<-started
	assert.Equal(t, of.NotReadyState, mp.Status())
	close(release)
	<-pollDone
	require.NoError(t, <-initDone)
	select {
	case s := <-violation:
		t.Fatalf("Status() reported %s while providers were still initializing", s)
	default:
	}
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

func TestMultiProvider_InitErrorUpdatesStatus(t *testing.T) {
	tests := map[string]struct {
		initError     error
		expectedState of.State
	}{
		"fatal provider init error": {
			initError:     &of.ProviderInitError{ErrorCode: of.ProviderFatalCode},
			expectedState: of.FatalState,
		},
		"wrapped fatal provider init error": {
			initError:     fmt.Errorf("wrapped: %w", &of.ProviderInitError{ErrorCode: of.ProviderFatalCode}),
			expectedState: of.FatalState,
		},
		"non-fatal provider init error": {
			initError:     &of.ProviderInitError{ErrorCode: of.GeneralCode},
			expectedState: of.ErrorState,
		},
		"plain init error": {
			initError:     errors.New("init failed"),
			expectedState: of.ErrorState,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := of.NewMockFeatureProvider(ctrl)
			provider.EXPECT().Metadata().Return(of.Metadata{Name: "MockProvider"})
			provider.EXPECT().Hooks().Return([]of.Hook{}).MinTimes(1)
			stateHandler := of.NewMockStateHandler(ctrl)
			stateHandler.EXPECT().Init(gomock.Any()).Return(test.initError)
			wrapped := struct {
				of.FeatureProvider
				of.StateHandler
			}{provider, stateHandler}

			mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", wrapped))
			require.NoError(t, err)

			require.Error(t, mp.Init(of.EvaluationContext{}))
			assert.Equal(t, test.expectedState, mp.Status())
		})
	}
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

func TestMultiProvider_UnknownEventDefaultsToNotReady(t *testing.T) {
	mp, err := NewProvider(
		StrategyFirstMatch,
		WithProvider("provider", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})),
	)
	require.NoError(t, err)

	mp.updateProviderState("provider", of.ErrorState)
	mp.updateProviderStateFromEvent(namedEvent{
		Event:        of.Event{EventType: of.EventType("PROVIDER_CUSTOM")},
		providerName: "provider",
	})

	assert.Equal(t, of.NotReadyState, mp.Status())
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
