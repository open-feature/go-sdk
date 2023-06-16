package openfeature

import (
	"golang.org/x/exp/slices"
	"reflect"
	"testing"
	"time"
)

var h1 func(details EventDetails)
var h2 func(details EventDetails)
var h3 func(details EventDetails)
var h4 func(details EventDetails)

func init() {
	h1 = func(details EventDetails) {
		// noop
	}

	h2 = func(details EventDetails) {
		// noop
	}

	h3 = func(details EventDetails) {
		// noop
	}

	h4 = func(details EventDetails) {
		// noop
	}
}

func TestEventHandler_Registration(t *testing.T) {
	t.Run("API handlers", func(t *testing.T) {
		handler := newEventHandler()

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
		handler := newEventHandler()

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
		handler := newEventHandler()

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
		handler := newEventHandler()

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

func TestEventHandler_RegisterUnregisterEventProvider(t *testing.T) {

	t.Run("Ignored non-eventing providers", func(t *testing.T) {
		handler := newEventHandler()
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

		handler := newEventHandler()
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

func TestEventHandler_Eventing(t *testing.T) {
	t.Run("Simple API level event", func(t *testing.T) {
		eventingImpl := &ProviderEventing{
			c: make(chan Event),
		}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		handler := newEventHandler()
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

		// trigger event
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
		case <-time.After(100 * time.Second):
			t.Errorf("timeout - event did not trigger")
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
			c: make(chan Event),
		}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		handler := newEventHandler()
		handler.registerEventingProvider(eventingProvider)

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		handler.registerClientHandler("clientA", ProviderReady, &callBack)

		fCh := []string{"flagA"}
		meta := map[string]interface{}{
			"key": "value",
		}

		// trigger event
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
		case <-time.After(100 * time.Second):
			t.Errorf("timeout - event did not trigger")
		}

		if result.client != "clientA" {
			t.Errorf("expected %s, but got %s", "clientA", result.client)
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
