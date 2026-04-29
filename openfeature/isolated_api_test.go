package openfeature

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

// Requirement 1.8.1: The API MUST provide a factory function that creates new, independent API instances.
func TestRequirement_1_8_1(t *testing.T) {
	instance := NewAPI()
	if instance == nil {
		t.Fatal("NewAPI() returned nil")
	}
}

// Requirement 1.8.2: Isolated instances MUST conform to the same API contract as the singleton (IEvaluation).
func TestRequirement_1_8_2(t *testing.T) {
	// compile-time check: *EvaluationAPI must satisfy IEvaluation
	var _ IEvaluation = NewAPI()
}

// Requirement 1.8.1 (independence): State set on an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_IndependentFromSingleton(t *testing.T) {
	t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	instanceProvider := NewMockFeatureProvider(ctrl)
	instanceProvider.EXPECT().Metadata().Return(Metadata{Name: "instance-provider"}).AnyTimes()

	instance := NewAPI()
	if err := instance.SetProviderAndWait(instanceProvider); err != nil {
		t.Fatalf("SetProviderAndWait on isolated instance: %v", err)
	}

	// Singleton should still have the default NoopProvider
	if api.GetProviderMetadata().Name == "instance-provider" {
		t.Error("provider set on isolated instance leaked into the singleton")
	}
}

// Requirement 1.8.1 (independence): State set on the singleton MUST NOT affect isolated instances.
func TestIsolatedAPI_SingletonDoesNotAffectInstance(t *testing.T) {
	t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	singletonProvider := NewMockFeatureProvider(ctrl)
	singletonProvider.EXPECT().Metadata().Return(Metadata{Name: "singleton-provider"}).AnyTimes()

	if err := SetProviderAndWait(singletonProvider); err != nil {
		t.Fatalf("SetProviderAndWait on singleton: %v", err)
	}

	instance := NewAPI()
	if instance.GetProviderMetadata().Name == "singleton-provider" {
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

	instance1 := NewAPI()
	instance2 := NewAPI()

	if err := instance1.SetProvider(sharedProvider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	// Registering the same provider on a different instance must return an error.
	if err := instance2.SetProvider(sharedProvider); err == nil {
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

	instance1 := NewAPI()
	instance2 := NewAPI()

	if err := instance1.SetProvider(sharedProvider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	// Replace the shared provider on instance1, which releases the binding.
	if err := instance1.SetProvider(replacementProvider); err != nil {
		t.Fatalf("SetProvider (replacement) on instance1: %v", err)
	}

	// Now the shared provider should be free to bind to instance2.
	if err := instance2.SetProvider(sharedProvider); err != nil {
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

	instance := NewAPI()

	if err := instance.SetProvider(provider); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}
	// Registering the same provider as a named provider on the same instance must succeed.
	if err := instance.SetNamedProvider("domain-a", provider, true); err != nil {
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

	instance1 := NewAPI()
	instance2 := NewAPI()

	if err := instance1.SetProvider(provider); err != nil {
		t.Fatalf("SetProvider on instance1: %v", err)
	}

	instance1.Shutdown()

	// After shutdown, the provider should be free to bind to another instance.
	if err := instance2.SetProvider(provider); err != nil {
		t.Errorf("expected no error binding provider after instance1 shutdown, got: %v", err)
	}
}

// GetClient / GetNamedClient on an isolated instance must return clients bound to that instance.
func TestIsolatedAPI_GetClientBoundToInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "instance-provider"}).AnyTimes()

	instance := NewAPI()
	if err := instance.SetProvider(provider); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}

	client := instance.GetClient()
	if client == nil {
		t.Fatal("GetClient() returned nil")
	}

	namedClient := instance.GetNamedClient("my-domain")
	if namedClient == nil {
		t.Fatal("GetNamedClient() returned nil")
	}
}

// Requirement 1.8.1 (independence): Hooks added to an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_HooksIndependence(t *testing.T) {
	t.Cleanup(initSingleton)

	instance := NewAPI()

	hook := UnimplementedHook{}
	instance.AddHooks(hook)

	singletonAPI := api.(*EvaluationAPI)
	singletonAPI.mu.RLock()
	singletonHookCount := len(singletonAPI.hks)
	singletonAPI.mu.RUnlock()

	if singletonHookCount != 0 {
		t.Errorf("hook added to isolated instance leaked into singleton: got %d hooks, want 0", singletonHookCount)
	}
}

// Requirement 1.8.1 (independence): Hooks added to the singleton MUST NOT affect an isolated instance.
func TestIsolatedAPI_SingletonHooksDoNotAffectInstance(t *testing.T) {
	t.Cleanup(initSingleton)

	AddHooks(UnimplementedHook{})

	instance := NewAPI()
	instance.mu.RLock()
	instanceHookCount := len(instance.hks)
	instance.mu.RUnlock()

	if instanceHookCount != 0 {
		t.Errorf("singleton hooks leaked into isolated instance: got %d hooks, want 0", instanceHookCount)
	}
}

// Requirement 1.8.1 (independence): EvaluationContext set on an isolated instance MUST NOT affect the singleton.
func TestIsolatedAPI_EvalContextIndependence(t *testing.T) {
	t.Cleanup(initSingleton)

	instance := NewAPI()
	instance.SetEvaluationContext(EvaluationContext{
		attributes: map[string]any{"tenant": "isolated"},
	})

	singletonAPI := api.(*EvaluationAPI)
	singletonAPI.mu.RLock()
	singletonCtx := singletonAPI.evalCtx
	singletonAPI.mu.RUnlock()

	if _, ok := singletonCtx.attributes["tenant"]; ok {
		t.Error("evaluation context set on isolated instance leaked into singleton")
	}
}

// Requirement 1.8.1 (independence): EvaluationContext set on the singleton MUST NOT affect an isolated instance.
func TestIsolatedAPI_SingletonEvalContextDoesNotAffectInstance(t *testing.T) {
	t.Cleanup(initSingleton)

	SetEvaluationContext(EvaluationContext{
		attributes: map[string]any{"tenant": "singleton"},
	})

	instance := NewAPI()
	instance.mu.RLock()
	instanceCtx := instance.evalCtx
	instance.mu.RUnlock()

	if _, ok := instanceCtx.attributes["tenant"]; ok {
		t.Error("singleton evaluation context leaked into isolated instance")
	}
}

// Requirement 1.8.1 (independence): Events on an isolated instance MUST NOT fire handlers on the singleton.
func TestIsolatedAPI_EventsIndependence(t *testing.T) {
	t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	provider := NewMockFeatureProvider(ctrl)
	provider.EXPECT().Metadata().Return(Metadata{Name: "event-test-provider"}).AnyTimes()

	singletonFired := false
	singletonHandler := func(details EventDetails) {
		singletonFired = true
	}
	AddHandler(ProviderReady, &singletonHandler)

	instance := NewAPI()
	if err := instance.SetProviderAndWait(provider); err != nil {
		t.Fatalf("SetProviderAndWait on isolated instance: %v", err)
	}

	// Give a short window for any erroneous event propagation.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
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

	instance1 := NewAPI()
	instance2 := NewAPI()

	var mu sync.Mutex
	var instance1Events []string

	cb := func(details EventDetails) {
		mu.Lock()
		instance1Events = append(instance1Events, details.ProviderName)
		mu.Unlock()
	}
	instance1.AddHandler(ProviderReady, &cb)

	// Setting a provider on instance2 should NOT trigger instance1's handler.
	if err := instance2.SetProviderAndWait(provider2); err != nil {
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
