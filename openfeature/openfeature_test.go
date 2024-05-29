package openfeature

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/open-feature/go-sdk/openfeature/internal"
)

// The `API`, and any state it maintains SHOULD exist as a global singleton,
// even in cases wherein multiple versions of the `API` are present at runtime.
func TestRequirement_1_1_1(t *testing.T) {
	defer t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()

	ofAPI := GetApiInstance()

	// set through instance level
	err := ofAPI.SetProvider(mockProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	// validate through global level
	if api.GetProvider() != mockProvider {
		t.Error("func SetProvider hasn't set the provider to the singleton")
	}
}

// The `API` MUST provide a function to set the default `provider`,
// which accepts an API-conformant `provider` implementation.
func TestRequirement_1_1_2_1(t *testing.T) {
	defer t.Cleanup(initSingleton)
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
		defer t.Cleanup(initSingleton)

		provider, initSem, _ := setupProviderWithSemaphores()

		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("initialization not invoked with provider registration")
		case <-initSem:
			break
		}

		if !reflect.DeepEqual(provider, api.GetProvider()) {
			t.Errorf("provider not updated to the one set")
		}
	})

	t.Run("named provider", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		provider, initSem, _ := setupProviderWithSemaphores()

		var client = "client"

		err := SetNamedProvider(client, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("initialization not invoked with provider registration")
		case <-initSem:
			break
		}

		if !reflect.DeepEqual(provider, api.GetNamedProviders()[client]) {
			t.Errorf("provider not updated to the one set")
		}
	})
}

// The provider mutator function MUST invoke the shutdown function on the previously registered provider once it's no
// longer being used to resolve flag values.
func TestRequirement_1_1_2_3(t *testing.T) {
	t.Run("default provider", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		provider, initSem, shutdownSem := setupProviderWithSemaphores()

		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("initialization not invoked with provider registration")
		case <-initSem:
			break
		}

		providerOverride, _, _ := setupProviderWithSemaphores()

		err = SetProvider(providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("shutdown not invoked for old default provider when registering new provider")
		case <-shutdownSem:
			break
		}

	})

	t.Run("named provider", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		provider, initSem, shutdownSem := setupProviderWithSemaphores()

		var client = "client"

		err := SetNamedProvider(client, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("initialization not invoked with provider registration")
		case <-initSem:
			break
		}

		providerOverride, _, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(client, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			t.Errorf("shutdown not invoked for old named provider when registering new provider")
		case <-shutdownSem:
			break
		}
	})

	t.Run("ignore shutdown for multiple references - default bound", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		// setup
		provider, _, shutdownSem := setupProviderWithSemaphores()

		// register provider multiple times
		err := SetProvider(provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		clientName := "clientA"

		err = SetNamedProvider(clientName, provider)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		providerOverride, _, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(clientName, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// validate
		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			break
		case <-shutdownSem:
			t.Errorf("shutdown called on the provider with multiple references")
		}
	})

	t.Run("ignore shutdown for multiple references - domain client bound", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		// setup
		providerA, _, shutdownSemA := setupProviderWithSemaphores()

		// register provider multiple times

		clientA := "clientA"
		clientB := "clientB"

		err := SetNamedProvider(clientA, providerA)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		err = SetNamedProvider(clientB, providerA)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		providerOverride, _, _ := setupProviderWithSemaphores()

		err = SetNamedProvider(clientA, providerOverride)
		if err != nil {
			t.Errorf("error setting up provider %v", err)
		}

		// validate
		select {
		// short enough wait time, but not too long
		case <-time.After(100 * time.Millisecond):
			break
		case <-shutdownSemA:
			t.Errorf("shutdown called on the provider with multiple references")
		}
	})
}

// The API SHOULD provide functions to set a provider and wait for the initialize function to return or throw.
func TestRequirement_1_1_2_4(t *testing.T) {
	defer t.Cleanup(initSingleton)

	t.Run("default provider", func(t *testing.T) {
		// given - a provider with state handling capability, with substantial initializing delay
		var initialized = false

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
		var initialized = false

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
		var initialized = false

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

		// when - registered async
		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		// then - must not update variable as registration is async
		if initialized != false {
			t.Errorf("expected uninitialized as async, but got true")
		}
	})
}

// The `API` MUST provide a function to bind a given `provider` to one or more client `domain`s.
// If the client-domain already has a bound provider, it is overwritten with the new mapping.
func TestRequirement_1_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)

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

	namedProviders := api.GetNamedProviders()

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

	provider, _, _ := api.ForEvaluation("clientA")
	if provider.Metadata().Name != "providerA" {
		t.Errorf("expected %s, but got %s", "providerA", providerA.Metadata().Name)
	}

	provider, _, _ = api.ForEvaluation("clientB")
	if provider.Metadata().Name != "providerB" {
		t.Errorf("expected %s, but got %s", "providerB", providerA.Metadata().Name)
	}

	// Validate overriding: If the client-domain already has a bound provider, it is overwritten with the new mapping.

	providerB2 := NewMockFeatureProvider(ctrl)
	providerB2.EXPECT().Metadata().Return(Metadata{Name: "providerB2"}).AnyTimes()

	err = SetNamedProvider("clientB", providerB2)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	namedProviders = api.GetNamedProviders()
	if namedProviders["clientB"] != providerB2 {
		t.Errorf("named provider overriding failed")
	}

	// Validate provider retrieval by client evaluation. This uses forTransaction("clientName")

	provider, _, _ = api.ForEvaluation("clientB")
	if provider.Metadata().Name != "providerB2" {
		t.Errorf("expected %s, but got %s", "providerB2", providerA.Metadata().Name)
	}
}

// The `API` MUST provide a function to add `hooks` which accepts one or more API-conformant `hooks`,
// and appends them to the collection of any previously added hooks. When new hooks are added,
// previously added hooks are not removed.
func TestRequirement_1_1_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	AddHooks(mockHook)
	AddHooks(mockHook, mockHook)

	if len(api.GetHooks()) != 3 {
		t.Error("func AddHooks didn't append the list of hooks to the existing collection of hooks")
	}
}

// The API MUST provide a function for retrieving the metadata field of the configured `provider`.
func TestRequirement_1_1_5(t *testing.T) {
	defer t.Cleanup(initSingleton)

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
	defer t.Cleanup(initSingleton)

	t.Run("client from direct invocation", func(t *testing.T) {
		client := NewClient("test-client")
		if client == nil {
			t.Errorf("expected an Client instance, but got invalid")
		}
	})

	t.Run("client from api level - no name", func(t *testing.T) {
		client := api.GetClient()
		if client == nil {
			t.Errorf("expected an IClient instance, but got invalid")
		}
	})

	t.Run("client from api level - with name", func(t *testing.T) {
		client := api.GetNamedClient("test-client")
		if client == nil {
			t.Errorf("expected an IClient instance, but got invalid")
		}
	})
}

// The client creation function MUST NOT throw, or otherwise abnormally terminate.
func TestRequirement_1_1_7(t *testing.T) {
	defer t.Cleanup(initSingleton)
	type clientCreationFunc func(name string) *Client

	// asserting that our NewClient method matches this signature is enough to deduce that no error is returned
	var f clientCreationFunc = NewClient

	use(f) // to avoid the declared and not used error
}

// The API MUST define a mechanism to propagate a shutdown request to active providers.
func TestRequirement_1_6_1(t *testing.T) {
	defer t.Cleanup(initSingleton)

	provider, initSem, shutdownSem := setupProviderWithSemaphores()

	// Setup provider and wait for initialization done
	err := SetProvider(provider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	select {
	// short enough wait time, but not too long
	case <-time.After(100 * time.Millisecond):
		t.Errorf("intialization timeout")
	case <-initSem:
		break
	}

	Shutdown()

	select {
	// short enough wait time, but not too long
	case <-time.After(100 * time.Millisecond):
		t.Errorf("shutdown not invoked")
	case <-shutdownSem:
		break
	}
}

func TestRequirement_EventCompliance(t *testing.T) {

	// The client MUST provide a function for associating handler functions with a particular provider event type.
	// The API and client MUST provide a function allowing the removal of event handlers.
	t.Run("requirement_5_2_1 & requirement_5_2_1", func(t *testing.T) {
		defer t.Cleanup(initSingleton)

		clientName := "OFClient"

		client := NewClient(clientName)

		// adding handlers
		client.AddHandler(ProviderReady, &h1)
		client.AddHandler(ProviderError, &h1)
		client.AddHandler(ProviderStale, &h1)
		client.AddHandler(ProviderConfigChange, &h1)

		registry := eventing.GetClientRegistry(clientName)

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
		defer t.Cleanup(initSingleton)

		// adding handlers
		AddHandler(ProviderReady, &h1)
		AddHandler(ProviderError, &h1)
		AddHandler(ProviderStale, &h1)
		AddHandler(ProviderConfigChange, &h1)

		registry := eventing.GetAPIRegistry()

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

		registry = eventing.GetAPIRegistry()

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
	defer t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)
	defaultProvider := NewMockFeatureProvider(ctrl)
	defaultProvider.EXPECT().Metadata().Return(Metadata{Name: "defaultClientReplacement"}).AnyTimes()

	err := SetProvider(defaultProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	// Validate provider retrieval by client evaluation
	provider, _, _ := api.ForEvaluation("ClientName")

	if provider.Metadata().Name != "defaultClientReplacement" {
		t.Errorf("expected %s, but got %s", "defaultClientReplacement", provider.Metadata().Name)
	}
}

// Ability to override default logger
func TestLoggerOverride(t *testing.T) {
	defer t.Cleanup(initSingleton)

	newOverride := internal.Logger{}
	SetLogger(logr.New(newOverride))

	if !reflect.DeepEqual(logger.GetSink(), newOverride) {
		t.Error("logger overriding failed")
	}
}

// Nil providers are not accepted for default and named providers
func TestForNilProviders(t *testing.T) {
	defer t.Cleanup(initSingleton)

	err := SetProvider(nil)
	if err == nil {
		t.Errorf("setting nil provider must result in an error")
	}

	err = SetNamedProvider("client", nil)
	if err == nil {
		t.Errorf("setting nil named provider must result in an error")
	}
}

func use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}

func setupProviderWithSemaphores() (struct {
	FeatureProvider
	StateHandler
	EventHandler
}, chan interface{}, chan interface{}) {
	intiSem := make(chan interface{}, 1)
	shutdownSem := make(chan interface{}, 1)

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
		State: NotReadyState,
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
