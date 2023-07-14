package openfeature

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
	"golang.org/x/exp/slices"
)

var logger logr.Logger

func init() {
	logger = logr.New(internal.Logger{})
}

// Requirement 5.1.1 The provider MAY define a mechanism for signaling the occurrence of one of a set of events,
// including PROVIDER_READY, PROVIDER_ERROR, PROVIDER_CONFIGURATION_CHANGED and PROVIDER_STALE,
// with a provider event details payload.
func TestEventHandler_RegisterUnregisterEventProvider(t *testing.T) {

	t.Run("Ignored non-eventing providers", func(t *testing.T) {
		executor := newEventExecutor(logger)
		err := executor.registerDefaultProvider(NoopProvider{})
		if err != nil {
			t.Fatal(err)
		}

		if executor.defaultProviderReference != nil {
			t.Errorf("implementation should ignore non eventing provider")
		}

		err = executor.registerNamedEventingProvider("name", NoopProvider{})
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.namedProviderReference) != 0 {
			t.Fatalf("implementation should ignore non eventing provider")
		}
	})

	t.Run("Accepts addition of eventing providers", func(t *testing.T) {
		eventingImpl := &ProviderEventing{}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		executor := newEventExecutor(logger)
		err := executor.registerDefaultProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if executor.defaultProviderReference == nil {
			t.Errorf("implementation should register eventing provider")
		}

		err = executor.registerNamedEventingProvider("name", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := executor.namedProviderReference["name"]; !ok {
			t.Errorf("implementation should register named eventing provider")
		}
	})
}

// Requirement 5.1.2 When a provider signals the occurrence of a particular event,
// the associated client and API event handlers MUST run.
func TestEventHandler_Eventing(t *testing.T) {
	t.Run("Simple API level event", func(t *testing.T) {
		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		executor := newEventExecutor(logger)
		err := executor.registerDefaultProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		executor.registerApiHandler(ProviderReady, &callBack)

		fCh := []string{"flagA"}
		meta := map[string]interface{}{
			"key": "value",
		}

		// trigger event from provider implementation
		eventingImpl.Invoke(Event{
			EventType: ProviderReady,
			ProviderEventDetails: ProviderEventDetails{
				Message:       "ReadyMessage",
				FlagChanges:   fCh,
				EventMetadata: meta,
			},
		})

		// wait for execution
		var result EventDetails
		select {
		case result = <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout - event did not trigger")
		}
		if result.Message != "ReadyMessage" {
			t.Errorf("expected %s, but got %s", "EventMessage", result.Message)
		}

		if !slices.Equal(result.FlagChanges, fCh) {
			t.Errorf("flag changes are not equal")
		}

		if !reflect.DeepEqual(result.EventMetadata, meta) {
			t.Errorf("metadata are not equal")
		}
	})

	t.Run("Simple Client level event", func(t *testing.T) {
		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		// NoopProvider backed event supported provider
		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		executor := newEventExecutor(logger)

		// associated to client name
		associatedName := "providerForClient"

		err := executor.registerNamedEventingProvider(associatedName, eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		executor.registerClientHandler(associatedName, ProviderReady, &callBack)

		fCh := []string{"flagA"}
		meta := map[string]interface{}{
			"key": "value",
		}

		// trigger event from provider implementation
		eventingImpl.Invoke(Event{
			ProviderName: eventingProvider.Metadata().Name,
			EventType:    ProviderReady,
			ProviderEventDetails: ProviderEventDetails{
				Message:       "ReadyMessage",
				FlagChanges:   fCh,
				EventMetadata: meta,
			},
		})

		// wait for execution
		var result EventDetails
		select {
		case result = <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout - event did not trigger")
		}

		if result.providerName != eventingProvider.Metadata().Name {
			t.Errorf("expected %s, but got %s", eventingProvider.Metadata().Name, result.providerName)
		}

		if result.Message != "ReadyMessage" {
			t.Errorf("expected %s, but got %s", "EventMessage", result.Message)
		}

		if !slices.Equal(result.FlagChanges, fCh) {
			t.Errorf("flag changes are not equal")
		}

		if !reflect.DeepEqual(result.EventMetadata, meta) {
			t.Errorf("metadata are not equal")
		}
	})
}

// Requirement 5.1.3 When a provider signals the occurrence of a particular event,
// event handlers on clients which are not associated with that provider MUST NOT run.
func TestEventHandler_clientAssociation(t *testing.T) {
	eventingImpl := &ProviderEventing{
		c: make(chan Event, 1),
	}

	// eventing supported provider
	defaultProvider := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		eventingImpl,
	}

	executor := newEventExecutor(logger)

	// default provider
	err := executor.registerDefaultProvider(defaultProvider)
	if err != nil {
		t.Fatal(err)
	}

	// named provider(associated to name someClient)
	err = executor.registerNamedEventingProvider("someClient", struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		&ProviderEventing{
			c: make(chan Event, 1),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	rsp := make(chan EventDetails)
	callBack := func(details EventDetails) {
		rsp <- details
	}

	event := ProviderReady
	executor.registerClientHandler("someClient", event, &callBack)

	// invoke default provider
	eventingImpl.Invoke(Event{
		ProviderName:         defaultProvider.Metadata().Name,
		EventType:            event,
		ProviderEventDetails: ProviderEventDetails{},
	})

	select {
	case <-rsp:
		t.Fatalf("incorrect association - executor must not have been invoked")
	case <-time.After(200 * time.Millisecond):
		break
	}
}

// Requirement 5.2.5 If a handler function terminates abnormally, other handler functions MUST run.
func TestEventHandler_ErrorHandling(t *testing.T) {
	errorCallback := func(e EventDetails) {
		panic("callback panic")
	}

	rsp := make(chan EventDetails, 1)
	successAPICallback := func(e EventDetails) {
		rsp <- e
	}

	rspClient := make(chan EventDetails, 1)
	successClientCallback := func(e EventDetails) {
		rspClient <- e
	}

	executor := newEventExecutor(logger)

	// api level handlers
	executor.registerApiHandler(ProviderReady, &errorCallback)
	executor.registerApiHandler(ProviderReady, &successAPICallback)

	// provider association
	provider := "providerA"

	// client level handlers
	executor.registerClientHandler(provider, ProviderReady, &errorCallback)
	executor.registerClientHandler(provider, ProviderReady, &successClientCallback)

	// trigger events manually
	go func() {
		executor.triggerEvent(Event{
			ProviderName:         provider,
			EventType:            ProviderReady,
			ProviderEventDetails: ProviderEventDetails{},
		}, "", true)
	}()

	select {
	case <-rsp:
		break
	case <-time.After(200 * time.Millisecond):
		t.Error("API level callback timeout - executor recovery was not successful")
	}

	select {
	case <-rspClient:
		break
	case <-time.After(200 * time.Millisecond):
		t.Error("client callback timeout - executor recovery was not successful")
	}
}

// Requirement 5.3.3 PROVIDER_READY handlers attached after the provider is already in a ready state MUST run immediately.
func TestEventHandler_ProviderReadiness(t *testing.T) {
	readyEventingProvider := struct {
		FeatureProvider
		StateHandler
		EventHandler
	}{
		NoopProvider{},
		&stateHandlerForTests{
			State: ReadyState,
		},
		&ProviderEventing{},
	}

	executor := newEventExecutor(logger)

	clientAssociation := "clientA"
	err := executor.registerNamedEventingProvider(clientAssociation, readyEventingProvider)
	if err != nil {
		t.Fatal(err)
	}

	rsp := make(chan EventDetails, 1)
	successAPICallback := func(e EventDetails) {
		rsp <- e
	}

	executor.registerClientHandler(clientAssociation, ProviderReady, &successAPICallback)

	select {
	case <-rsp:
		break
	case <-time.After(200 * time.Millisecond):
		t.Errorf("timedout waiting for ready state callback, but got none")
	}
}

// Contract tests - registration & removal

func TestEventHandler_Registration(t *testing.T) {
	t.Run("API handlers", func(t *testing.T) {
		executor := newEventExecutor(logger)

		// Add multiple - ProviderReady
		executor.registerApiHandler(ProviderReady, &h1)
		executor.registerApiHandler(ProviderReady, &h2)
		executor.registerApiHandler(ProviderReady, &h3)
		executor.registerApiHandler(ProviderReady, &h4)

		// Add multiple - ProviderError
		executor.registerApiHandler(ProviderError, &h2)
		executor.registerApiHandler(ProviderError, &h2)

		// Add single types
		executor.registerApiHandler(ProviderStale, &h3)
		executor.registerApiHandler(ProviderConfigChange, &h4)

		readyLen := len(executor.apiRegistry[ProviderReady])
		if readyLen != 4 {
			t.Errorf("expected %d events, but got %d", 4, readyLen)
		}

		errLen := len(executor.apiRegistry[ProviderError])
		if errLen != 2 {
			t.Errorf("expected %d events, but got %d", 2, errLen)
		}

		staleLen := len(executor.apiRegistry[ProviderStale])
		if staleLen != 1 {
			t.Errorf("expected %d events, but got %d", 1, staleLen)
		}

		cfgLen := len(executor.apiRegistry[ProviderConfigChange])
		if cfgLen != 1 {
			t.Errorf("expected %d events, but got %d", 1, cfgLen)
		}
	})

	t.Run("Client handlers", func(t *testing.T) {
		executor := newEventExecutor(logger)

		// Add multiple - client a
		executor.registerClientHandler("a", ProviderReady, &h1)
		executor.registerClientHandler("a", ProviderReady, &h2)
		executor.registerClientHandler("a", ProviderReady, &h3)
		executor.registerClientHandler("a", ProviderReady, &h4)

		// Add single for rest of the client
		executor.registerClientHandler("b", ProviderError, &h2)
		executor.registerClientHandler("c", ProviderStale, &h3)
		executor.registerClientHandler("d", ProviderConfigChange, &h4)

		readyLen := len(executor.scopedRegistry["a"].callbacks[ProviderReady])
		if readyLen != 4 {
			t.Errorf("expected %d events in client a, but got %d", 4, readyLen)
		}

		errLen := len(executor.scopedRegistry["b"].callbacks[ProviderError])
		if errLen != 1 {
			t.Errorf("expected %d events in client b, but got %d", 1, errLen)
		}

		staleLen := len(executor.scopedRegistry["c"].callbacks[ProviderStale])
		if staleLen != 1 {
			t.Errorf("expected %d events in client c, but got %d", 1, staleLen)
		}

		cfgLen := len(executor.scopedRegistry["d"].callbacks[ProviderConfigChange])
		if cfgLen != 1 {
			t.Errorf("expected %d events in client d, but got %d", 1, cfgLen)
		}
	})
}

func TestEventHandler_APIRemoval(t *testing.T) {
	t.Run("API level removal", func(t *testing.T) {
		executor := newEventExecutor(logger)

		// Add multiple - ProviderReady
		executor.registerApiHandler(ProviderReady, &h1)
		executor.registerApiHandler(ProviderReady, &h2)
		executor.registerApiHandler(ProviderReady, &h3)
		executor.registerApiHandler(ProviderReady, &h4)

		// Add single types
		executor.registerApiHandler(ProviderError, &h2)
		executor.registerApiHandler(ProviderStale, &h3)
		executor.registerApiHandler(ProviderConfigChange, &h4)

		// removal
		executor.removeApiHandler(ProviderReady, &h1)
		executor.removeApiHandler(ProviderError, &h2)
		executor.removeApiHandler(ProviderStale, &h3)
		executor.removeApiHandler(ProviderConfigChange, &h4)

		readyLen := len(executor.apiRegistry[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events, but got %d", 3, readyLen)
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], &h2) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		errLen := len(executor.apiRegistry[ProviderError])
		if errLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, errLen)
		}

		staleLen := len(executor.apiRegistry[ProviderStale])
		if staleLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, staleLen)
		}

		cfgLen := len(executor.apiRegistry[ProviderConfigChange])
		if cfgLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, cfgLen)
		}
	})

	t.Run("Client level removal", func(t *testing.T) {
		executor := newEventExecutor(logger)

		// Add multiple - client a
		executor.registerClientHandler("a", ProviderReady, &h1)
		executor.registerClientHandler("a", ProviderReady, &h2)
		executor.registerClientHandler("a", ProviderReady, &h3)
		executor.registerClientHandler("a", ProviderReady, &h4)

		// Add single
		executor.registerClientHandler("b", ProviderError, &h2)
		executor.registerClientHandler("c", ProviderStale, &h3)
		executor.registerClientHandler("d", ProviderConfigChange, &h4)

		// removal
		executor.removeClientHandler("a", ProviderReady, &h1)
		executor.removeClientHandler("b", ProviderError, &h2)
		executor.removeClientHandler("c", ProviderStale, &h3)
		executor.removeClientHandler("d", ProviderConfigChange, &h4)

		readyLen := len(executor.scopedRegistry["a"].callbacks[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events in client a, but got %d", 3, readyLen)
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], &h2) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], &h4) {
			t.Errorf("expected callback to be present")
		}

		errLen := len(executor.scopedRegistry["b"].callbacks[ProviderError])
		if errLen != 0 {
			t.Errorf("expected %d events in client b, but got %d", 0, errLen)
		}

		staleLen := len(executor.scopedRegistry["c"].callbacks[ProviderStale])
		if staleLen != 0 {
			t.Errorf("expected %d events in client c, but got %d", 0, staleLen)
		}

		cfgLen := len(executor.scopedRegistry["d"].callbacks[ProviderConfigChange])
		if cfgLen != 0 {
			t.Errorf("expected %d events in client d, but got %d", 0, cfgLen)
		}

		// removal referenced to non-existing clients does nothing & no panics
		executor.removeClientHandler("non-existing", ProviderReady, &h1)
		executor.removeClientHandler("b", ProviderReady, &h1)
	})

	t.Run("remove handlers that were not added", func(t *testing.T) {
		executor := newEventExecutor(logger)

		// removal of non-added handlers shall not panic
		executor.removeApiHandler(ProviderReady, &h1)
		executor.removeClientHandler("a", ProviderReady, &h1)
	})
}
