package openfeature

import (
	"testing"

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
	// compile-time check: NewAPI() must return IEvaluation
	_ = NewAPI()
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

	instance1 := NewAPI().(*evaluationAPI)
	instance2 := NewAPI().(*evaluationAPI)

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

	instance1 := NewAPI().(*evaluationAPI)
	instance2 := NewAPI().(*evaluationAPI)

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

	instance := NewAPI().(*evaluationAPI)

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

	instance1 := NewAPI().(*evaluationAPI)
	instance2 := NewAPI().(*evaluationAPI)

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
