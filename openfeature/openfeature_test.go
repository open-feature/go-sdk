package openfeature

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

// The `API`, and any state it maintains SHOULD exist as a global singleton,
// even in cases wherein multiple versions of the `API` are present at runtime.
func TestRequirement_1_1_1(t *testing.T) {
	t.Cleanup(resetSingleton)

	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()

	ofAPI := api()

	// set through instance level
	err := ofAPI.SetProvider(t.Context(), mockProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	// validate through global level
	if api().getProvider() != mockProvider {
		t.Error("func SetProvider hasn't set the provider to the singleton")
	}
}

// The `API` MUST provide a function to set the default `provider`,
// which accepts an API-conformant `provider` implementation.
func TestRequirement_1_1_2_1(t *testing.T) {
	t.Cleanup(resetSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProviderName := "mock-provider"
	mockProvider.EXPECT().Metadata().Return(Metadata{Name: mockProviderName}).AnyTimes()

	err := SetProvider(mockProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	if ProviderMetadata() != mockProvider.Metadata() {
		t.Error("globally set provider's metadata doesn't match the mock provider's metadata")
	}
}

// The provider mutator function MUST invoke the initialize function on the newly registered provider before using it
// to resolve flag values.
func TestRequirement_1_1_2_2(t *testing.T) {
	t.Run("default provider", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		provider, initSem, _ := setupProviderWithSemaphores()

		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		expectChannelReceive(t, initSem, "initialization not invoked with provider registration")

		if !reflect.DeepEqual(provider, api().getProvider()) {
			t.Errorf("provider not updated to the one set")
		}
	})

	t.Run("named provider", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		provider, initSem, _ := setupProviderWithSemaphores()

		client := "client"

		err := SetNamedProvider(client, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		expectChannelReceive(t, initSem, "initialization not invoked with provider registration")

		if !reflect.DeepEqual(provider, api().getDomainProviders()[client]) {
			t.Errorf("provider not updated to the one set")
		}
	})
}

// The provider mutator function MUST invoke the shutdown function on the previously registered provider once it's no
// longer being used to resolve flag values.
func TestRequirement_1_1_2_3(t *testing.T) {
	t.Run("default provider", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		provider, initSem, shutdownSem := setupProviderWithSemaphores()

		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		expectChannelReceive(t, initSem, "initialization not invoked with provider registration")

		providerOverride, overrideInitSem, _ := setupProviderWithSemaphores()

		err = SetProvider(providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the override provider's initialization semaphore to prevent goroutine leak
		expectChannelReceive(t, overrideInitSem, "override provider initialization not invoked")
		expectChannelReceive(t, shutdownSem, "shutdown not invoked for old default provider when registering new provider")
	})

	t.Run("named provider", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		provider, initSem, shutdownSem := setupProviderWithSemaphores()

		client := "client"

		err := SetNamedProvider(client, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		expectChannelReceive(t, initSem, "initialization not invoked with provider registration")

		providerOverride, overrideInitSem, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(client, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the override provider's initialization semaphore to prevent goroutine leak
		expectChannelReceive(t, overrideInitSem, "override provider initialization not invoked")
		expectChannelReceive(t, shutdownSem, "shutdown not invoked for old named provider when registering new provider")
	})

	t.Run("ignore shutdown for multiple references - default bound", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		// setup
		provider, initSem, shutdownSem := setupProviderWithSemaphores()

		// register provider multiple times
		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the initialization semaphore
		expectChannelReceive(t, initSem, "provider initialization not invoked")

		clientName := "clientA"

		err = SetNamedProvider(clientName, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		providerOverride, overrideInitSem, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(clientName, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the override provider's initialization semaphore to prevent goroutine leak
		expectChannelReceive(t, overrideInitSem, "override provider initialization not invoked")

		// validate
		expectTimeout(t, shutdownSem, "shutdown called on the provider with multiple references")
	})

	t.Run("ignore shutdown for multiple references - domain client bound", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		// setup
		providerA, initSemA, shutdownSemA := setupProviderWithSemaphores()

		// register provider multiple times

		clientA := "clientA"
		clientB := "clientB"

		err := SetNamedProvider(clientA, providerA)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the initialization semaphore for providerA
		expectChannelReceive(t, initSemA, "providerA initialization not invoked")

		err = SetNamedProvider(clientB, providerA)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		providerOverride, overrideInitSem, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(clientA, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// Consume the override provider's initialization semaphore to prevent goroutine leak
		expectChannelReceive(t, overrideInitSem, "override provider initialization not invoked")

		// validate
		expectTimeout(t, shutdownSemA, "shutdown called on the provider with multiple references")
	})
}

// The API SHOULD provide functions to set a provider and wait for the initialize function to return or throw.
func TestRequirement_1_1_2_4(t *testing.T) {
	t.Cleanup(resetSingleton)

	t.Run("default provider", func(t *testing.T) {
		// given - a provider with state handling capability, with substantial initializing delay
		initialized := false

		provider := struct {
			FeatureProvider
			StateHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					<-time.After(200 * time.Millisecond)
					initialized = true
					return nil
				},
			},
		}

		// when - registered
		err := SetProviderAndWait(provider)
		if err != nil {
			t.Fatal(err)
		}

		// then - must update variable as call is blocking
		if initialized != true {
			t.Errorf("expected initialization, but got false")
		}
	})

	t.Run("named provider", func(t *testing.T) {
		// given - a provider with state handling capability, with substantial initializing delay
		initialized := false

		provider := struct {
			FeatureProvider
			StateHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					<-time.After(200 * time.Millisecond)
					initialized = true
					return nil
				},
			},
		}

		// when - registered
		err := SetNamedProviderAndWait("someName", provider)
		if err != nil {
			t.Fatal(err)
		}

		// then - must update variable as call is blocking
		if initialized != true {
			t.Errorf("expected initialization, but got false")
		}
	})

	t.Run("error return and eventing", func(t *testing.T) {
		// given - provider with initialization error & error handlers registered
		provider := struct {
			FeatureProvider
			StateHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					<-time.After(200 * time.Millisecond)
					return errors.New("some initialization error")
				},
			},
		}

		errChan := make(chan EventDetails, 1)
		errHandler := func(details EventDetails) {
			errChan <- details
		}

		AddHandler(ProviderError, &errHandler)

		// when
		err := SetProviderAndWait(provider)

		// then
		if err == nil {
			t.Fatal("expected error to be non-nil, but got nil")
		}

		var errEvent EventDetails

		select {
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected error event, but time out waiting for event")
		case errEvent = <-errChan:
			break
		}

		if errEvent.Message == "" {
			t.Fatal("expected non empty event message, but got empty")
		}
	})

	t.Run("async registration to validate by contradiction", func(t *testing.T) {
		// given - a provider with state handling capability, with substantial initializing delay
		initialized := false

		s := make(chan struct{}) // to block the initialization
		provider := struct {
			FeatureProvider
			StateHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					s <- struct{}{} // initialization is blocked until read from the channel
					initialized = true
					return nil
				},
			},
		}

		// when - registered async
		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		// then - must not update variable as registration is async
		if initialized != false {
			t.Errorf("expected uninitialized as async, but got true")
		}
		<-s
	})
}

// The `API` MUST provide a function to bind a given `provider` to one or more client `domain`s.
// If the client-domain already has a bound provider, it is overwritten with the new mapping.
func TestRequirement_1_1_3(t *testing.T) {
	t.Cleanup(resetSingleton)

	// Setup

	ctrl := gomock.NewController(t)
	providerA := NewMockFeatureProvider(ctrl)
	providerA.EXPECT().Metadata().Return(Metadata{Name: "providerA"}).AnyTimes()

	providerB := NewMockFeatureProvider(ctrl)
	providerB.EXPECT().Metadata().Return(Metadata{Name: "providerB"}).AnyTimes()

	err := SetNamedProvider("clientA", providerA)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	err = SetNamedProvider("clientB", providerB)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	namedProviders := api().getDomainProviders()

	// Validate binding

	if len(namedProviders) != 2 {
		t.Errorf("expected %d providers, but got %d", 2, len(namedProviders))
	}

	if namedProviders["clientA"] != providerA {
		t.Errorf("invalid provider binding")
	}

	if namedProviders["clientB"] != providerB {
		t.Errorf("invalid provider binding")
	}

	// Validate provider retrieval by client evaluation. This uses forTransaction("clientName")

	provider, _, _ := api().resolveBinding("clientA")
	if provider.Metadata().Name != "providerA" {
		t.Errorf("expected %s, but got %s", "providerA", providerA.Metadata().Name)
	}

	provider, _, _ = api().resolveBinding("clientB")
	if provider.Metadata().Name != "providerB" {
		t.Errorf("expected %s, but got %s", "providerB", provider.Metadata().Name)
	}

	// Validate overriding: If the client-domain already has a bound provider, it is overwritten with the new mapping.

	providerB2 := NewMockFeatureProvider(ctrl)
	providerB2.EXPECT().Metadata().Return(Metadata{Name: "providerB2"}).AnyTimes()

	err = SetNamedProvider("clientB", providerB2)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	namedProviders = api().getDomainProviders()
	if namedProviders["clientB"] != providerB2 {
		t.Errorf("named provider overriding failed")
	}

	// Validate provider retrieval by client evaluation. This uses forTransaction("clientName")

	provider, _, _ = api().resolveBinding("clientB")
	if provider.Metadata().Name != "providerB2" {
		t.Errorf("expected %s, but got %s", "providerB2", provider.Metadata().Name)
	}
}

// The `API` MUST provide a function to add `hooks` which accepts one or more API-conformant `hooks`,
// and appends them to the collection of any previously added hooks. When new hooks are added,
// previously added hooks are not removed.
func TestRequirement_1_1_4(t *testing.T) {
	t.Cleanup(resetSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	AddHooks(mockHook)
	AddHooks(mockHook, mockHook)

	if len(api().getHooks()) != 3 {
		t.Error("func AddHooks didn't append the list of hooks to the existing collection of hooks")
	}
}

// The API MUST provide a function for retrieving the metadata field of the configured `provider`.
func TestRequirement_1_1_5(t *testing.T) {
	t.Cleanup(resetSingleton)

	t.Run("default provider", func(t *testing.T) {
		defaultProvider := NoopProvider{}
		err := SetProvider(defaultProvider)
		if err != nil {
			t.Errorf("provider registration failed %v", err)
		}
		if ProviderMetadata() != defaultProvider.Metadata() {
			t.Error("default global provider's metadata isn't NoopProvider's metadata")
		}
	})

	t.Run("named provider", func(t *testing.T) {
		defaultProvider := NoopProvider{}
		name := "test-provider"

		err := SetNamedProvider(name, defaultProvider)
		if err != nil {
			t.Errorf("provider registration failed %v", err)
		}
		if NamedProviderMetadata(name) != defaultProvider.Metadata() {
			t.Error("default global provider's metadata isn't NoopProvider's metadata")
		}
	})
}

// The `API` MUST provide a function for creating a `client` which accepts the following options:
// - domain (optional): A logical string identifier for the client.
func TestRequirement_1_1_6(t *testing.T) {
	t.Cleanup(resetSingleton)

	t.Run("client from direct invocation", func(t *testing.T) {
		client := NewClient("test-client")
		if client == nil {
			t.Errorf("expected an Client instance, but got invalid")
		}
	})

	t.Run("client from api level - no domain", func(t *testing.T) {
		client := api().NewClient()
		if client == nil {
			t.Errorf("expected an IClient instance, but got invalid")
		}
	})

	t.Run("client from api level - with domain", func(t *testing.T) {
		client := api().NewClient(WithDomain("test-client"))
		if client == nil {
			t.Errorf("expected an IClient instance, but got invalid")
		}
	})
}

// The client creation function MUST NOT throw, or otherwise abnormally terminate.
func TestRequirement_1_1_7(t *testing.T) {
	t.Cleanup(resetSingleton)
	type clientCreationFunc func(name string) *Client

	// asserting that our NewClient method matches this signature is enough to deduce that no error is returned
	var f clientCreationFunc = NewClient

	use(f) // to avoid the declared and not used error
}

// The API MUST define a mechanism to propagate a shutdown request to active providers.
func TestRequirement_1_6_1(t *testing.T) {
	t.Cleanup(resetSingleton)

	provider, initSem, shutdownSem := setupProviderWithSemaphores()

	// Setup provider and wait for initialization done
	err := SetProvider(provider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	expectChannelReceive(t, initSem, "intialization timeout")

	Shutdown()

	expectChannelReceive(t, shutdownSem, "shutdown not invoked")
}

// The API's `shutdown` function MUST reset all state of the API, removing all
// hooks, event handlers, and providers.
func TestRequirement_1_6_2(t *testing.T) {
	t.Cleanup(resetSingleton)

	// TODO: test that hooks and event handlers are removed as well. This only
	// tests that providers are removed.

	provider1, _, shutdownSem1 := setupProviderWithSemaphores()

	// Setup provider and wait for initialization done
	err := SetProviderAndWait(provider1)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	Shutdown()

	// Shutdown should be synchronous. Try a non-blocking receive and fail
	// immediately if there is not a value in the channel.
	select {
	case <-shutdownSem1:
		break
	default:
		t.Fatalf("shutdown not invoked")
	}

	// Try shutting down again, and make sure that provider1 is not shut down
	// again, since it is now inactive.
	provider2, _, shutdownSem2 := setupProviderWithSemaphores()

	err = SetProviderAndWait(provider2)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	Shutdown()

	select {
	case <-shutdownSem2:
		break
	default:
		t.Fatalf("shutdown not invoked")
	}

	expectTimeout(t, shutdownSem1, "provider1 should not have been shut down again, since it is unregistered")
}

func TestRequirement_EventCompliance(t *testing.T) {
	// The client MUST provide a function for associating handler functions with a particular provider event type.
	// The API and client MUST provide a function allowing the removal of event handlers.
	t.Run("requirement_5_2_1 & requirement_5_2_1", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		clientName := "OFClient"

		client := NewClient(clientName)

		// adding handlers
		client.AddHandler(ProviderReady, &h1)
		client.AddHandler(ProviderError, &h1)
		client.AddHandler(ProviderStale, &h1)
		client.AddHandler(ProviderConfigChange, &h1)

		registry := api().eventExecutor.GetClientRegistry(clientName)

		if len(registry.eventCallbacks()[ProviderReady]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry.eventCallbacks()[ProviderError]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry.eventCallbacks()[ProviderStale]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry.eventCallbacks()[ProviderConfigChange]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		// removing handlers
		client.RemoveHandler(ProviderReady, &h1)
		client.RemoveHandler(ProviderError, &h1)
		client.RemoveHandler(ProviderStale, &h1)
		client.RemoveHandler(ProviderConfigChange, &h1)

		if len(registry.eventCallbacks()[ProviderReady]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry.eventCallbacks()[ProviderError]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry.eventCallbacks()[ProviderStale]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry.eventCallbacks()[ProviderConfigChange]) > 0 {
			t.Errorf("expected empty registrations")
		}
	})

	// The API MUST provide a function for associating handler functions with a particular provider event type.
	t.Run("requirement_5_2_2 & requirement_5_2_1", func(t *testing.T) {
		t.Cleanup(resetSingleton)

		// adding handlers
		AddHandler(ProviderReady, &h1)
		AddHandler(ProviderError, &h1)
		AddHandler(ProviderStale, &h1)
		AddHandler(ProviderConfigChange, &h1)

		registry := api().eventExecutor.GetAPIRegistry()

		if len(registry[ProviderReady]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry[ProviderError]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry[ProviderStale]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		if len(registry[ProviderConfigChange]) < 1 {
			t.Errorf("expected a registry regiration, but got none")
		}

		// removing handlers
		RemoveHandler(ProviderReady, &h1)
		RemoveHandler(ProviderError, &h1)
		RemoveHandler(ProviderStale, &h1)
		RemoveHandler(ProviderConfigChange, &h1)

		registry = api().eventExecutor.GetAPIRegistry()

		if len(registry[ProviderReady]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry[ProviderError]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry[ProviderStale]) > 0 {
			t.Errorf("expected empty registrations")
		}

		if len(registry[ProviderConfigChange]) > 0 {
			t.Errorf("expected empty registrations")
		}
	})
}

// Non-spec bound validations

// If there is no client domain bound provider, then return the default provider
func TestDefaultClientUsage(t *testing.T) {
	t.Cleanup(resetSingleton)

	ctrl := gomock.NewController(t)
	defaultProvider := NewMockFeatureProvider(ctrl)
	defaultProvider.EXPECT().Metadata().Return(Metadata{Name: "defaultClientReplacement"}).AnyTimes()

	err := SetProvider(defaultProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	// Validate provider retrieval by client evaluation
	provider, _, _ := api().resolveBinding("ClientName")

	if provider.Metadata().Name != "defaultClientReplacement" {
		t.Errorf("expected %s, but got %s", "defaultClientReplacement", provider.Metadata().Name)
	}
}

func TestLateBindingOfDefaultProvider(t *testing.T) {
	t.Cleanup(resetSingleton)
	// we are expecting
	expectedResultUnboundProvider := "default-value-from-unbound-provider"
	expectedResultFromLateDefaultProvider := "value-from-late-default-provider"

	ctrl := gomock.NewController(t)
	defaultProvider := NewMockFeatureProvider(ctrl)
	defaultProvider.EXPECT().Metadata().Return(Metadata{Name: "defaultClientReplacement"}).AnyTimes()
	defaultProvider.EXPECT().Hooks().AnyTimes().Return([]Hook{})
	defaultProvider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(StringResolutionDetail{Value: expectedResultFromLateDefaultProvider})

	client := NewClient("app")
	strResult, err := client.StringValue(t.Context(), "flag", expectedResultUnboundProvider, EvaluationContext{})
	if err != nil {
		t.Errorf("flag evaluation failed %v", err)
	}

	if strResult != expectedResultUnboundProvider {
		t.Errorf("expected %s, but got %s", expectedResultUnboundProvider, strResult)
	}

	err = SetProviderAndWait(defaultProvider)
	if err != nil {
		t.Errorf("provider registration failed %v", err)
	}

	strResult, err = client.StringValue(t.Context(), "flag", "default", EvaluationContext{})
	if err != nil {
		t.Errorf("flag evaluation failed %v", err)
	}

	if strResult != expectedResultFromLateDefaultProvider {
		t.Errorf("expected %s, but got %s", expectedResultFromLateDefaultProvider, strResult)
	}
}

// Nil providers are not accepted for default and named providers
func TestForNilProviders(t *testing.T) {
	t.Cleanup(resetSingleton)

	err := SetProvider(nil)
	if err == nil {
		t.Errorf("setting nil provider must result in an error")
	}

	err = SetNamedProvider("client", nil)
	if err == nil {
		t.Errorf("setting nil named provider must result in an error")
	}
}

func use(vals ...any) {
	for _, val := range vals {
		_ = val
	}
}

func setupProviderWithSemaphores() (struct {
	FeatureProvider
	StateHandler
	EventHandler
}, chan any, chan any,
) {
	intiSem := make(chan any, 1)
	shutdownSem := make(chan any, 1)

	sh := &stateHandlerForTests{
		// Semaphore must be invoked
		initF: func(e EvaluationContext) error {
			intiSem <- ""
			return nil
		},
		// Semaphore must be invoked
		shutdownF: func() {
			shutdownSem <- ""
		},
	}

	eventing := &ProviderEventing{
		c: make(chan Event, 1),
	}

	// Derive provider with initialize & shutdown support
	provider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		FeatureProvider: NoopProvider{},
		StateHandler:    sh,
		EventHandler:    eventing,
	}

	return provider, intiSem, shutdownSem
}

// expectChannelReceive waits for a channel to receive a value within the timeout.
// It fails the test with the provided message if the timeout occurs first.
func expectChannelReceive(t *testing.T, ch <-chan any, timeoutMsg string) {
	t.Helper()
	select {
	case <-time.After(100 * time.Millisecond):
		t.Error(timeoutMsg)
	case <-ch:
	}
}

// expectTimeout waits for a timeout to occur, expecting no value from the channel.
// It fails the test with the provided message if the channel receives a value.
func expectTimeout(t *testing.T, ch <-chan any, receiveMsg string) {
	t.Helper()
	select {
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	case <-ch:
		t.Error(receiveMsg)
	}
}
