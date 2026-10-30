package openfeature

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/open-feature/go-sdk/openfeature/internal/factory"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
)

// Requirement 1.8.1: The API MUST provide a factory function that creates new, independent API instances.
func TestRequirement_1_8_1(t *testing.T) {
	instance := newAPI()
	if instance == nil {
		t.Fatal("newAPI() returned nil")
	}
}

// Requirement 1.8.2: Isolated instances MUST conform to the same API contract as the singleton.
func TestRequirement_1_8_2(t *testing.T) {
	if _, ok := factory.NewAPI().(*EvaluationAPI); !ok {
		t.Error("factory.NewAPI() did not return *EvaluationAPI")
	}
}

// Requirement 1.8.1 (independence): State set on an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_IndependentFromSingleton(t *testing.T) {
	t.Cleanup(resetSingleton)

	ctrl := gomock.NewController(t)
	instanceProvider := NewMockFeatureProvider(ctrl)
	instanceProvider.EXPECT().Metadata().Return(Metadata{Name: "instance-provider"}).AnyTimes()

	instance := newAPI()
	if err := instance.SetProviderAndWait(t.Context(), instanceProvider); err != nil {
		t.Fatalf("SetProviderAndWait on isolated instance: %v", err)
	}

	// Singleton should still have the default NoopProvider
	if api().getProviderMetadata().Name == "instance-provider" {
		t.Error("provider set on isolated instance leaked into the singleton")
	}
}

// Requirement 1.8.1 (independence): State set on the singleton MUST NOT affect isolated instances.
func TestIsolatedAPI_SingletonDoesNotAffectInstance(t *testing.T) {
	t.Cleanup(resetSingleton)

	ctrl := gomock.NewController(t)
	singletonProvider := NewMockFeatureProvider(ctrl)
	singletonProvider.EXPECT().Metadata().Return(Metadata{Name: "singleton-provider"}).AnyTimes()

	if err := SetProviderAndWait(singletonProvider); err != nil {
		t.Fatalf("SetProviderAndWait on singleton: %v", err)
	}

	instance := newAPI()
	if instance.getProviderMetadata().Name == "singleton-provider" {
		t.Error("singleton provider leaked into newly created isolated instance")
	}
}

// Requirement 1.8.4: A provider instance SHOULD NOT be bound to more than one API instance at a time.
func TestRequirement_1_8_4_CrossInstanceBinding(t *testing.T) {
	t.Cleanup(func() {
		providerBindingsMu.Lock()
		providerBindings = make(map[uintptr]*providerBindingEntry)
		providerBindingsMu.Unlock()
	})

	ctrl := gomock.NewController(t)
	sharedProvider := NewMockFeatureProvider(ctrl)
	sharedProvider.EXPECT().Metadata().Return(Metadata{Name: "shared-provider"}).AnyTimes()

	instance1 := newAPI()
	instance2 := newAPI()

	if err := instance1.SetProvider(t.Context(), sharedProvider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	// Registering the same provider on a different instance must return an error.
	if err := instance2.SetProvider(t.Context(), sharedProvider); err == nil {
		t.Error("expected error when binding a provider already bound to another instance, got nil")
	}
}

// Requirement 1.8.4 (release): After a provider is released from one instance it CAN be registered on another.
func TestRequirement_1_8_4_ProviderReleasedAfterReplacement(t *testing.T) {
	t.Cleanup(func() {
		providerBindingsMu.Lock()
		providerBindings = make(map[uintptr]*providerBindingEntry)
		providerBindingsMu.Unlock()
	})

	ctrl := gomock.NewController(t)
	sharedProvider := NewMockFeatureProvider(ctrl)
	sharedProvider.EXPECT().Metadata().Return(Metadata{Name: "shared-provider"}).AnyTimes()

	replacementProvider := NewMockFeatureProvider(ctrl)
	replacementProvider.EXPECT().Metadata().Return(Metadata{Name: "replacement-provider"}).AnyTimes()

	instance1 := newAPI()
	instance2 := newAPI()

	if err := instance1.SetProvider(t.Context(), sharedProvider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	// Replace the shared provider on instance1, which releases the binding.
	if err := instance1.SetProvider(t.Context(), replacementProvider); err != nil {
		t.Fatalf("SetProvider (replacement) on instance1: %v", err)
	}

	// Now the shared provider should be free to bind to instance2.
	if err := instance2.SetProvider(t.Context(), sharedProvider); err != nil {
		t.Errorf("expected no error binding released provider to instance2, got: %v", err)
	}
}

// Requirement 1.8.4: The same provider CAN be bound to default and named domains within the same instance.
func TestRequirement_1_8_4_SameInstanceMultipleDomains(t *testing.T) {
	t.Cleanup(func() {
		providerBindingsMu.Lock()
		providerBindings = make(map[uintptr]*providerBindingEntry)
		providerBindingsMu.Unlock()
	})

	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "multi-domain-provider"}).AnyTimes()

	instance := newAPI()

	if err := instance.SetProvider(t.Context(), provider); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}
	// Registering the same provider as a named provider on the same instance must succeed.
	if err := instance.SetProvider(t.Context(), provider, WithDomain("domain-a")); err != nil {
		t.Errorf("expected no error binding same provider to a second domain on same instance, got: %v", err)
	}
}

// Requirement 1.8.4: Provider bindings are released when an isolated instance is shut down.
func TestRequirement_1_8_4_ReleasedOnShutdown(t *testing.T) {
	t.Cleanup(func() {
		providerBindingsMu.Lock()
		providerBindings = make(map[uintptr]*providerBindingEntry)
		providerBindingsMu.Unlock()
	})

	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "shutdown-provider"}).AnyTimes()

	instance1 := newAPI()
	instance2 := newAPI()

	if err := instance1.SetProvider(t.Context(), provider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	if err := instance1.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown: %v", err)
	}

	// After shutdown, the provider should be free to bind to another instance.
	if err := instance2.SetProvider(t.Context(), provider); err != nil {
		t.Errorf("expected no error binding provider after instance1 shutdown, got: %v", err)
	}
}

// NewClient on an isolated instance must return clients bound to that instance.
func TestIsolatedAPI_NewClientBoundToInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "instance-provider"}).AnyTimes()

	instance := newAPI()
	if err := instance.SetProvider(t.Context(), provider); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}

	client := instance.NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	namedClient := instance.NewClient(WithDomain("my-domain"))
	if namedClient == nil {
		t.Fatal("NewClient() returned nil")
	}
}

// Requirement 1.8.1 (independence): Hooks added to an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_HooksIndependence(t *testing.T) {
	t.Cleanup(resetSingleton)

	instance := newAPI()

	hook := UnimplementedHook{}
	instance.AddHooks(hook)

	singletonAPI := api()
	singletonAPI.mu.RLock()
	singletonHookCount := len(singletonAPI.hks)
	singletonAPI.mu.RUnlock()

	if singletonHookCount != 0 {
		t.Errorf("hook added to isolated instance leaked into singleton: got %d hooks, want 0", singletonHookCount)
	}
}

// Requirement 1.8.1 (independence): Hooks added to the singleton MUST NOT affect an isolated instance.
func TestIsolatedAPI_SingletonHooksDoNotAffectInstance(t *testing.T) {
	t.Cleanup(resetSingleton)

	AddHooks(UnimplementedHook{})

	instance := newAPI()
	instance.mu.RLock()
	instanceHookCount := len(instance.hks)
	instance.mu.RUnlock()

	if instanceHookCount != 0 {
		t.Errorf("singleton hooks leaked into isolated instance: got %d hooks, want 0", instanceHookCount)
	}
}

// Requirement 1.8.1 (independence): EvaluationContext set on an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_EvalContextIndependence(t *testing.T) {
	t.Cleanup(resetSingleton)

	instance := newAPI()
	instance.SetEvaluationContext(EvaluationContext{
		attributes: map[string]any{"tenant": "isolated"},
	})

	singletonAPI := api()
	singletonAPI.mu.RLock()
	singletonCtx := singletonAPI.evalCtx
	singletonAPI.mu.RUnlock()

	if _, ok := singletonCtx.attributes["tenant"]; ok {
		t.Error("evaluation context set on isolated instance leaked into singleton")
	}
}

// Requirement 1.8.1 (independence): EvaluationContext set on the singleton MUST NOT affect an isolated instance.
func TestIsolatedAPI_SingletonEvalContextDoesNotAffectInstance(t *testing.T) {
	t.Cleanup(resetSingleton)

	SetEvaluationContext(EvaluationContext{
		attributes: map[string]any{"tenant": "singleton"},
	})

	instance := newAPI()
	instance.mu.RLock()
	instanceCtx := instance.evalCtx
	instance.mu.RUnlock()

	if _, ok := instanceCtx.attributes["tenant"]; ok {
		t.Error("singleton evaluation context leaked into isolated instance")
	}
}

// Requirement 1.8.1 (independence): Events on an isolated instance MUST NOT fire handlers on the singleton.
func TestIsolatedAPI_EventsIndependence(t *testing.T) {
	t.Cleanup(resetSingleton)

	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "event-test-provider"}).AnyTimes()

	singletonFired := false
	singletonHandler := func(details EventDetails) {
		singletonFired = true
	}
	AddHandler(ProviderReady, &singletonHandler)

	instance := newAPI()
	if err := instance.SetProviderAndWait(t.Context(), provider); err != nil {
		t.Fatalf("SetProviderAndWait on isolated instance: %v", err)
	}

	// Give a short window for any erroneous event propagation.
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	<-ctx.Done()

	if singletonFired {
		t.Error("provider ready event on isolated instance fired singleton handler")
	}
}

// Requirement 1.8.1 (independence): Events between two isolated instances do not interfere.
func TestIsolatedAPI_EventsBetweenInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider2 := NewMockFeatureProvider(ctrl)
	provider2.EXPECT().Metadata().Return(Metadata{Name: "instance2-provider"}).AnyTimes()

	instance1 := newAPI()
	instance2 := newAPI()

	var mu sync.Mutex
	var instance1Events []string

	cb := func(details EventDetails) {
		mu.Lock()
		instance1Events = append(instance1Events, details.ProviderName)
		mu.Unlock()
	}
	instance1.AddHandler(ProviderReady, &cb)

	// Setting a provider on instance2 should NOT trigger instance1's handler.
	if err := instance2.SetProviderAndWait(t.Context(), provider2); err != nil {
		t.Fatalf("SetProviderAndWait on instance2: %v", err)
	}

	// Give a short window for any erroneous event propagation.
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	for _, name := range instance1Events {
		if name == "instance2-provider" {
			t.Error("provider event from instance2 fired handler on instance1")
		}
	}
}

func TestIsolatedAPI_ShutdownStopsEventExecutor(t *testing.T) {
	startLeakTest(t)

	defer goleak.VerifyNone(t)

	instance := newAPI()

	defaultProvider := struct {
		FeatureProvider
		EventHandler
	}{NoopProvider{}, &ProviderEventing{c: make(chan Event, 1)}}

	if err := instance.SetProvider(t.Context(), defaultProvider); err != nil {
		t.Fatalf("SetProvider on isolated instance: %v", err)
	}

	namedProvider := struct {
		FeatureProvider
		EventHandler
	}{NoopProvider{}, &ProviderEventing{c: make(chan Event, 1)}}

	if err := instance.SetProvider(t.Context(), namedProvider, WithDomain("test-domain")); err != nil {
		t.Fatalf("SetNamedProvider on isolated instance: %v", err)
	}

	// Allow the executor goroutines to start handling events.
	time.Sleep(50 * time.Millisecond)

	if err := instance.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown: %v", err)
	}

	// goleak will verify no goroutines leak from the isolated instance's
	// per-instance event executor.
}

// Requirement 1.8.1 (lifecycle): ShutdownWithContext on an isolated instance
// must also stop the per-instance event executor goroutine.
func TestIsolatedAPI_ShutdownWithContextStopsEventExecutor(t *testing.T) {
	startLeakTest(t)

	defer goleak.VerifyNone(t)

	instance := newAPI()

	eventingProvider := struct {
		FeatureProvider
		EventHandler
	}{NoopProvider{}, &ProviderEventing{c: make(chan Event, 1)}}

	if err := instance.SetProvider(t.Context(), eventingProvider); err != nil {
		t.Fatalf("SetProvider on isolated instance: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := instance.Shutdown(t.Context()); err != nil {
		t.Fatalf("ShutdownWithContext on isolated instance: %v", err)
	}
}

// A domain whose own provider is still initializing MUST report NOT_READY rather than inheriting
// the default provider's status, per 1.7.1 / 1.7.3 and the section 1.7 lifecycle diagram.
func TestDomainStateWhileProviderInitializing(t *testing.T) {
	api := newAPI()
	t.Cleanup(func() {
		_ = api.Shutdown(context.Background()) //nolint:usetesting
	})

	if err := api.SetProviderAndWait(t.Context(), NoopProvider{}); err != nil {
		t.Fatalf("failed to set up default provider: %v", err)
	}
	if got := api.eventExecutor.State(defaultDomain); got != ReadyState {
		t.Fatalf("expected the default provider to be READY, got %q", got)
	}

	initializing := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			initF: func(EvaluationContext) error {
				// Block until test cleanup so the provider never leaves NOT_READY.
				<-t.Context().Done()
				return nil
			},
		},
		&ProviderEventing{},
	}

	if _, err := api.setDomainProvider(t.Context(), "foo", initializing); err != nil {
		t.Fatalf("failed to set up the domain provider: %v", err)
	}

	if got := api.eventExecutor.State("foo"); got != NotReadyState {
		t.Errorf("expected domain \"foo\" to report NOT_READY while its provider initializes, got %q", got)
	}
}

// Requirement 1.6.2 and 1.8.2: Shutdown on an isolated instance MUST reset the
// state of that instance, removing all hooks, event handlers, evaluation context
// and providers, just as Shutdown does for the global singleton.
func TestRequirement_1_8_2_ShutdownResetsState(t *testing.T) {
	instance := newAPI()
	t.Cleanup(func() {
		_ = instance.Shutdown(context.Background()) //nolint:usetesting
	})

	defaultProvider := newResetTestProvider("default-provider")
	domainProvider := newResetTestProvider("domain-provider")
	if err := instance.SetProviderAndWait(t.Context(), defaultProvider); err != nil {
		t.Fatalf("SetProviderAndWait: %v", err)
	}
	if err := instance.SetProviderAndWait(t.Context(), domainProvider, WithDomain("domain")); err != nil {
		t.Fatalf("SetProviderAndWait with domain: %v", err)
	}

	hook := &countingHook{}
	instance.AddHooks(hook)
	instance.SetEvaluationContext(NewEvaluationContext("global-key", map[string]any{"global": "yes"}))

	handler := func(EventDetails) {}
	instance.AddHandler(ProviderReady, &handler)
	client := instance.NewClient(WithDomain("domain"))
	client.AddHandler(ProviderReady, &handler)

	if err := instance.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	// A second Shutdown must not shut down the providers again: they are no longer registered.
	if err := instance.Shutdown(t.Context()); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
	if got := defaultProvider.shutdowns.Load(); got != 1 {
		t.Errorf("expected the default provider to be shut down once, got %d", got)
	}
	if got := domainProvider.shutdowns.Load(); got != 1 {
		t.Errorf("expected the domain provider to be shut down once, got %d", got)
	}

	// Providers: every domain resolves to the no-op provider again.
	for _, domain := range []string{"", "domain"} {
		provider, hooks, evalCtx := instance.resolveBinding(domain)
		if _, ok := provider.(NoopProvider); !ok {
			t.Errorf("domain %q: expected NoopProvider after Shutdown, got %q", domain, provider.Metadata().Name)
		}
		// Hooks and evaluation context.
		if len(hooks) != 0 {
			t.Errorf("domain %q: expected no API hooks after Shutdown, got %d", domain, len(hooks))
		}
		if evalCtx.TargetingKey() != "" || len(evalCtx.Attributes()) != 0 {
			t.Errorf("domain %q: expected an empty evaluation context after Shutdown, got %v", domain, evalCtx)
		}
	}
	if got := len(instance.getDomainProviders()); got != 0 {
		t.Errorf("expected no domain providers after Shutdown, got %d", got)
	}

	// Event handlers.
	if got := len(instance.eventExecutor.GetAPIRegistry()[ProviderReady]); got != 0 {
		t.Errorf("expected no API handlers after Shutdown, got %d", got)
	}
	if got := len(instance.eventExecutor.GetClientRegistry("domain").callbacks[ProviderReady]); got != 0 {
		t.Errorf("expected no client handlers after Shutdown, got %d", got)
	}

	// A client created before Shutdown sees the reset instance.
	if got := client.State(); got != NotReadyState {
		t.Errorf("expected client state %q after Shutdown, got %q", NotReadyState, got)
	}
	before := hook.calls.Load()
	details, err := client.BooleanValueDetails(t.Context(), "flag", false, EvaluationContext{})
	if err != nil {
		t.Fatalf("BooleanValueDetails: %v", err)
	}
	if details.Value || details.Reason != DefaultReason {
		t.Errorf("expected the default value from the no-op provider, got %v (reason %q)", details.Value, details.Reason)
	}
	if hook.calls.Load() != before {
		t.Error("API hook added before Shutdown still runs after Shutdown")
	}
}

// Requirement 1.6.2 and 1.8.2: Shutdown resets an isolated instance rather than
// closing it, so the instance can be used again. Providers registered after
// Shutdown must work and their events must reach handlers.
func TestRequirement_1_8_2_ReuseAfterShutdown(t *testing.T) {
	domain := []APIOption{WithDomain("domain")}
	tests := map[string]struct {
		first, second []APIOption
	}{
		"default provider after default provider": {},
		"domain provider after default provider":  {second: domain},
		"domain provider after domain provider":   {first: domain, second: domain},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			instance := newAPI()
			t.Cleanup(func() {
				_ = instance.Shutdown(context.Background()) //nolint:usetesting
			})
			// Created before Shutdown, so it must keep working with the reset instance.
			client := instance.NewClient(tt.second...)

			if err := instance.SetProviderAndWait(t.Context(), newResetTestProvider("first"), tt.first...); err != nil {
				t.Fatalf("SetProviderAndWait: %v", err)
			}
			if err := instance.Shutdown(t.Context()); err != nil {
				t.Fatalf("Shutdown: %v", err)
			}

			second := newResetTestProvider("second")
			if err := instance.SetProviderAndWait(t.Context(), second, tt.second...); err != nil {
				t.Fatalf("SetProviderAndWait after Shutdown: %v", err)
			}

			if !client.Boolean(t.Context(), "flag", false, EvaluationContext{}) {
				t.Error("expected the provider set after Shutdown to be used")
			}

			received := make(chan EventDetails, 1)
			handler := func(details EventDetails) { received <- details }
			client.AddHandler(ProviderConfigChange, &handler)

			second.events <- Event{ProviderName: "second", EventType: ProviderConfigChange}
			select {
			case details := <-received:
				if details.ProviderName != "second" {
					t.Errorf("expected an event from %q, got one from %q", "second", details.ProviderName)
				}
			case <-time.After(time.Second):
				t.Fatal("event from the provider set after Shutdown never reached the handler")
			}
		})
	}
}

// resetTestProvider is a pointer-type provider that returns the opposite of the
// default value for boolean flags, counts Shutdown calls and emits events.
type resetTestProvider struct {
	NoopProvider
	name      string
	events    chan Event
	shutdowns atomic.Int32
}

func newResetTestProvider(name string) *resetTestProvider {
	return &resetTestProvider{name: name, events: make(chan Event, 1)}
}

func (p *resetTestProvider) Metadata() Metadata {
	return Metadata{Name: p.name}
}

func (p *resetTestProvider) BooleanEvaluation(_ context.Context, _ string, defaultValue bool, _ FlattenedContext) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value:                    !defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: StaticReason},
	}
}

func (p *resetTestProvider) Init(EvaluationContext) error { return nil }

func (p *resetTestProvider) Shutdown() { p.shutdowns.Add(1) }

func (p *resetTestProvider) EventChannel() <-chan Event { return p.events }

// countingHook counts how many times its Before stage runs.
type countingHook struct {
	UnimplementedHook
	calls atomic.Int32
}

func (h *countingHook) Before(context.Context, HookContext, HookHints) (*EvaluationContext, error) {
	h.calls.Add(1)
	return nil, nil
}
