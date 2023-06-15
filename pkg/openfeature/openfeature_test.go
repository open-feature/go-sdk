package openfeature

import (
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
	"reflect"
	"testing"
	"time"
)

// The `API`, and any state it maintains SHOULD exist as a global singleton,
// even in cases wherein multiple versions of the `API` are present at runtime.
func TestRequirement_1_1_1(t *testing.T) {
	defer t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()

	err := SetProvider(mockProvider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	if getProvider() != mockProvider {
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

		if !reflect.DeepEqual(provider, getProvider()) {
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

		if !reflect.DeepEqual(provider, getNamedProviders()[client]) {
			t.Errorf("provider not updated to the one set")
		}
	})
}

// The provider mutator function MUST invoke the shutdown function on the previously registered provider once it's no
// longer being used to resolve flag values.

// The `API` MUST provide a function to bind a given `provider` to one or more client `name`s.
// If the client-name already has a bound provider, it is overwritten with the new mapping.
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

	namedProviders := getNamedProviders()

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

	provider, _, _ := forTransaction("clientA")
	if provider.Metadata().Name != "providerA" {
		t.Errorf("expected %s, but got %s", "providerA", providerA.Metadata().Name)
	}

	provider, _, _ = forTransaction("clientB")
	if provider.Metadata().Name != "providerB" {
		t.Errorf("expected %s, but got %s", "providerB", providerA.Metadata().Name)
	}

	// Validate overriding: If the client-name already has a bound provider, it is overwritten with the new mapping.

	providerB2 := NewMockFeatureProvider(ctrl)
	providerB2.EXPECT().Metadata().Return(Metadata{Name: "providerB2"}).AnyTimes()

	err = SetNamedProvider("clientB", providerB2)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	namedProviders = getNamedProviders()
	if namedProviders["clientB"] != providerB2 {
		t.Errorf("named provider overriding failed")
	}

	// Validate provider retrieval by client evaluation. This uses forTransaction("clientName")

	provider, _, _ = forTransaction("clientB")
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

	if len(getHooks()) != 3 {
		t.Error("func AddHooks didn't append the list of hooks to the existing collection of hooks")
	}
}

// The API MUST provide a function for retrieving the metadata field of the configured `provider`.
func TestRequirement_1_1_5(t *testing.T) {
	defer t.Cleanup(initSingleton)
	defaultProvider := NoopProvider{}
	if ProviderMetadata() != defaultProvider.Metadata() {
		t.Error("default global provider's metadata isn't NoopProvider's metadata")
	}
}

// The `API` MUST provide a function for creating a `client` which accepts the following options:
// - name (optional): A logical string identifier for the client.
func TestRequirement_1_1_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	NewClient("test-client")
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

	// Setup provider and wait for initialization done - minor delay
	err := SetProvider(provider)
	if err != nil {
		t.Errorf("error setting up provider %v", err)
	}

	select {
	// short enough wait time, but not too long
	case <-time.After(100 * time.Millisecond):
		t.Errorf("intialization delayed")
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

// Non-spec bound validations

// If there is no client name bound provider, then return the default provider
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
	provider, _, _ := forTransaction("ClientName")

	if provider.Metadata().Name != "defaultClientReplacement" {
		t.Errorf("expected %s, but got %s", "defaultClientReplacement", provider.Metadata().Name)
	}
}

// Ability to override default logger
func TestLoggerOverride(t *testing.T) {
	defer t.Cleanup(initSingleton)

	newOverride := internal.Logger{}
	SetLogger(logr.New(newOverride))

	if !reflect.DeepEqual(globalLogger().GetSink(), newOverride) {
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
}, chan interface{}, chan interface{}) {
	intiSem := make(chan interface{}, 1)
	shutdownSem := make(chan interface{}, 1)

	sh := &StateHandlerForTests{
		initF: func(e EvaluationContext) {
			intiSem <- ""
		},
		// Semaphore must be invoked
		shutdownF: func() {
			shutdownSem <- ""
		},
		State: ReadyState,
	}

	// Derive provider with initialize & shutdown support
	provider := struct {
		FeatureProvider
		StateHandler
	}{
		FeatureProvider: NoopProvider{},
		StateHandler:    sh,
	}

	return provider, intiSem, shutdownSem
}
