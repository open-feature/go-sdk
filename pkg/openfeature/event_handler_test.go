package openfeature

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"reflect"
	"testing"
	"time"
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
		handler := newEventHandler(logger)
		handler.registerEventingProvider(NoopProvider{})

		if len(handler.providerShutdownHook) != 0 {
			t.Errorf("implementation should ignore non eventing provider")
		}

		err := handler.unregisterEventingProvider(NoopProvider{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Accepts addition and removal of eventing providers", func(t *testing.T) {
		eventingImpl := &ProviderEventing{}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		handler := newEventHandler(logger)
		handler.registerEventingProvider(eventingProvider)

		if len(handler.providerShutdownHook) != 1 {
			t.Errorf("implementation should register eventing provider")
		}

		err := handler.unregisterEventingProvider(eventingProvider)
		if err != nil {
			t.Error(err)
		}

		if len(handler.providerShutdownHook) != 0 {
			t.Errorf("implementation should allow removal of eventing provider")
		}

		// double removal check for nil check & avoid panic
		err = handler.unregisterEventingProvider(eventingProvider)
		if err != nil {
			t.Error(err)
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

		handler := newEventHandler(logger)
		handler.registerEventingProvider(eventingProvider)

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		handler.registerApiHandler(ProviderReady, &callBack)

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
		case <-time.After(handlerExecutionTime):
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

		handler := newEventHandler(logger)
		handler.registerEventingProvider(eventingProvider)

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		// associated to NoopProvider
		handler.registerClientHandler(eventingProvider.Metadata().Name, ProviderReady, &callBack)

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
		case <-time.After(handlerExecutionTime):
			t.Fatalf("timeout - event did not trigger")
		}

		if result.client != eventingProvider.Metadata().Name {
			t.Errorf("expected %s, but got %s", eventingProvider.Metadata().Name, result.client)
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

	// NoopProvider backed event supported provider
	eventingProvider := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		eventingImpl,
	}

	// event handler & provider registration
	handler := newEventHandler(logger)
	handler.registerEventingProvider(eventingProvider)

	rsp := make(chan EventDetails)
	callBack := func(details EventDetails) {
		rsp <- details
	}

	event := ProviderReady

	// register a random client - no association with registered provider
	handler.registerClientHandler("someClient", event, &callBack)

	// invoke provider event
	eventingImpl.Invoke(Event{
		ProviderName:         eventingProvider.Metadata().Name,
		EventType:            event,
		ProviderEventDetails: ProviderEventDetails{},
	})

	select {
	case <-rsp:
		t.Fatalf("incorrect association - handler must not have been invoked")
	case <-time.After(handlerExecutionTime):
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

	handler := newEventHandler(logger)

	// api level handlers
	handler.registerApiHandler(ProviderReady, &errorCallback)
	handler.registerApiHandler(ProviderReady, &successAPICallback)

	// provider association
	provider := "providerA"

	// client level handlers
	handler.registerClientHandler(provider, ProviderReady, &errorCallback)
	handler.registerClientHandler(provider, ProviderReady, &successClientCallback)

	// trigger events manually
	go func() {
		_ = handler.triggerEvent(Event{
			ProviderName:         provider,
			EventType:            ProviderReady,
			ProviderEventDetails: ProviderEventDetails{},
		})
	}()

	select {
	case <-rsp:
		break
	case <-time.After(handlerExecutionTime):
		t.Error("API level callback timeout - handler recovery was not successful")
	}

	select {
	case <-rspClient:
		break
	case <-time.After(handlerExecutionTime):
		t.Error("client callback timeout - handler recovery was not successful")
	}
}

// Make sure event handler cannot block
func TestEventHandler_Timeout(t *testing.T) {
	timeoutCallback := func(e EventDetails) {
		time.Sleep(handlerExecutionTime * 10)
	}

	handler := newEventHandler(logger)
	handler.registerApiHandler(ProviderReady, &timeoutCallback)

	group, ctx := errgroup.WithContext(context.Background())

	group.Go(func() error {
		return handler.triggerEvent(Event{
			ProviderName:         "provider",
			EventType:            ProviderReady,
			ProviderEventDetails: ProviderEventDetails{},
		})
	})

	select {
	case <-ctx.Done():
		break
	case <-time.After(handlerExecutionTime * 2):
		t.Fatalf("timeout while waiting for condition")
	}

	err := group.Wait()
	if err == nil {
		t.Errorf("expected timeout error, but got none")
	}
}

// Contract tests - registration & removal

func TestEventHandler_Registration(t *testing.T) {
	t.Run("API handlers", func(t *testing.T) {
		handler := newEventHandler(logger)

		// Add multiple - ProviderReady
		handler.registerApiHandler(ProviderReady, &h1)
		handler.registerApiHandler(ProviderReady, &h2)
		handler.registerApiHandler(ProviderReady, &h3)
		handler.registerApiHandler(ProviderReady, &h4)

		// Add multiple - ProviderError
		handler.registerApiHandler(ProviderError, &h2)
		handler.registerApiHandler(ProviderError, &h2)

		// Add single types
		handler.registerApiHandler(ProviderStale, &h3)
		handler.registerApiHandler(ProviderConfigChange, &h4)

		readyLen := len(handler.apiRegistry[ProviderReady])
		if readyLen != 4 {
			t.Errorf("expected %d events, but got %d", 4, readyLen)
		}

		errLen := len(handler.apiRegistry[ProviderError])
		if errLen != 2 {
			t.Errorf("expected %d events, but got %d", 2, errLen)
		}

		staleLen := len(handler.apiRegistry[ProviderStale])
		if staleLen != 1 {
			t.Errorf("expected %d events, but got %d", 1, staleLen)
		}

		cfgLen := len(handler.apiRegistry[ProviderConfigChange])
		if cfgLen != 1 {
			t.Errorf("expected %d events, but got %d", 1, cfgLen)
		}
	})

	t.Run("Client handlers", func(t *testing.T) {
		handler := newEventHandler(logger)

		// Add multiple - client a
		handler.registerClientHandler("a", ProviderReady, &h1)
		handler.registerClientHandler("a", ProviderReady, &h2)
		handler.registerClientHandler("a", ProviderReady, &h3)
		handler.registerClientHandler("a", ProviderReady, &h4)

		// Add single for rest of the client
		handler.registerClientHandler("b", ProviderError, &h2)
		handler.registerClientHandler("c", ProviderStale, &h3)
		handler.registerClientHandler("d", ProviderConfigChange, &h4)

		readyLen := len(handler.clientRegistry["a"].holder[ProviderReady])
		if readyLen != 4 {
			t.Errorf("expected %d events in client a, but got %d", 4, readyLen)
		}

		errLen := len(handler.clientRegistry["b"].holder[ProviderError])
		if errLen != 1 {
			t.Errorf("expected %d events in client b, but got %d", 1, errLen)
		}

		staleLen := len(handler.clientRegistry["c"].holder[ProviderStale])
		if staleLen != 1 {
			t.Errorf("expected %d events in client c, but got %d", 1, staleLen)
		}

		cfgLen := len(handler.clientRegistry["d"].holder[ProviderConfigChange])
		if cfgLen != 1 {
			t.Errorf("expected %d events in client d, but got %d", 1, cfgLen)
		}
	})
}

func TestEventHandler_APIRemoval(t *testing.T) {
	t.Run("API level removal", func(t *testing.T) {
		handler := newEventHandler(logger)

		// Add multiple - ProviderReady
		handler.registerApiHandler(ProviderReady, &h1)
		handler.registerApiHandler(ProviderReady, &h2)
		handler.registerApiHandler(ProviderReady, &h3)
		handler.registerApiHandler(ProviderReady, &h4)

		// Add single types
		handler.registerApiHandler(ProviderError, &h2)
		handler.registerApiHandler(ProviderStale, &h3)
		handler.registerApiHandler(ProviderConfigChange, &h4)

		// removal
		handler.removeApiHandler(ProviderReady, &h1)
		handler.removeApiHandler(ProviderError, &h2)
		handler.removeApiHandler(ProviderStale, &h3)
		handler.removeApiHandler(ProviderConfigChange, &h4)

		readyLen := len(handler.apiRegistry[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events, but got %d", 3, readyLen)
		}

		if !slices.Contains(handler.apiRegistry[ProviderReady], &h2) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(handler.apiRegistry[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(handler.apiRegistry[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		errLen := len(handler.apiRegistry[ProviderError])
		if errLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, errLen)
		}

		staleLen := len(handler.apiRegistry[ProviderStale])
		if staleLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, staleLen)
		}

		cfgLen := len(handler.apiRegistry[ProviderConfigChange])
		if cfgLen != 0 {
			t.Errorf("expected %d events, but got %d", 0, cfgLen)
		}
	})

	t.Run("Client level removal", func(t *testing.T) {
		handler := newEventHandler(logger)

		// Add multiple - client a
		handler.registerClientHandler("a", ProviderReady, &h1)
		handler.registerClientHandler("a", ProviderReady, &h2)
		handler.registerClientHandler("a", ProviderReady, &h3)
		handler.registerClientHandler("a", ProviderReady, &h4)

		// Add single
		handler.registerClientHandler("b", ProviderError, &h2)
		handler.registerClientHandler("c", ProviderStale, &h3)
		handler.registerClientHandler("d", ProviderConfigChange, &h4)

		// removal
		handler.removeClientHandler("a", ProviderReady, &h1)
		handler.removeClientHandler("b", ProviderError, &h2)
		handler.removeClientHandler("c", ProviderStale, &h3)
		handler.removeClientHandler("d", ProviderConfigChange, &h4)

		readyLen := len(handler.clientRegistry["a"].holder[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events in client a, but got %d", 3, readyLen)
		}

		if !slices.Contains(handler.clientRegistry["a"].holder[ProviderReady], &h2) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(handler.clientRegistry["a"].holder[ProviderReady], &h3) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(handler.clientRegistry["a"].holder[ProviderReady], &h4) {
			t.Errorf("expected callback to be present")
		}

		errLen := len(handler.clientRegistry["b"].holder[ProviderError])
		if errLen != 0 {
			t.Errorf("expected %d events in client b, but got %d", 0, errLen)
		}

		staleLen := len(handler.clientRegistry["c"].holder[ProviderStale])
		if staleLen != 0 {
			t.Errorf("expected %d events in client c, but got %d", 0, staleLen)
		}

		cfgLen := len(handler.clientRegistry["d"].holder[ProviderConfigChange])
		if cfgLen != 0 {
			t.Errorf("expected %d events in client d, but got %d", 0, cfgLen)
		}

		// removal referenced to non-existing clients does nothing & no panics
		handler.removeClientHandler("non-existing", ProviderReady, &h1)
		handler.removeClientHandler("b", ProviderReady, &h1)
	})
}
