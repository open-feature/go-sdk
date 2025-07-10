package openfeature

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
)

func init() {
}

// Requirement 5.1.1 The provider MAY define a mechanism for signaling the occurrence of one of a set of events,
// including PROVIDER_READY, PROVIDER_ERROR, PROVIDER_CONFIGURATION_CHANGED and PROVIDER_STALE,
// with a provider event details payload.
func TestEventHandler_RegisterUnregisterEventProvider(t *testing.T) {
	t.Run("Accepts addition of eventing providers", func(t *testing.T) {
		eventingImpl := &ProviderEventing{}

		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		executor := newEventExecutor()
		err := executor.registerDefaultProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if executor.defaultProviderReference.featureProvider != eventingProvider {
			t.Error("implementation should register default eventing provider")
		}

		err = executor.registerNamedEventingProvider("domain", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := executor.namedProviderReference["domain"]; !ok {
			t.Errorf("implementation should register named eventing provider")
		}
	})
}

// Requirement 5.1.2 When a provider signals the occurrence of a particular event,
// the associated client and API event handlers MUST run.
func TestEventHandler_Eventing(t *testing.T) {
	t.Run("Simple API level event", func(t *testing.T) {
		t.Cleanup(initSingleton)

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

		err := SetProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		eventType := ProviderConfigChange
		AddHandler(eventType, &callBack)

		fCh := []string{"flagA"}
		meta := map[string]any{
			"key": "value",
		}

		// trigger event from provider implementation
		eventingImpl.Invoke(Event{
			EventType: eventType,
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
		t.Cleanup(initSingleton)

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

		// associated to client domain
		associatedName := "providerForClient"

		err := SetNamedProviderAndWait(associatedName, eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails)
		callBack := func(details EventDetails) {
			rsp <- details
		}

		client := NewClient(associatedName)
		client.AddHandler(ProviderError, &callBack)

		fCh := []string{"flagA"}
		meta := map[string]any{
			"key": "value",
		}

		// trigger event from provider implementation
		eventingImpl.Invoke(Event{
			ProviderName: eventingProvider.Metadata().Name,
			EventType:    ProviderError,
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

		if result.ProviderName != eventingProvider.Metadata().Name {
			t.Errorf("expected %s, but got %s", eventingProvider.Metadata().Name, result.ProviderName)
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
	t.Cleanup(initSingleton)

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

	// default provider
	err := SetProviderAndWait(defaultProvider)
	if err != nil {
		t.Fatal(err)
	}

	// named provider(associated to domain someClient)
	err = SetNamedProviderAndWait("someClient", struct {
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

	event := ProviderError
	client := NewClient("someClient")
	client.AddHandler(event, &callBack)

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
	t.Cleanup(initSingleton)

	eventing := &ProviderEventing{
		c: make(chan Event, 1),
	}

	provider := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		eventing,
	}

	failingCallback := func(e EventDetails) {
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

	err := SetProvider(provider)
	if err != nil {
		t.Fatal(err)
	}

	successEventType := ProviderStale

	// api level handlers
	AddHandler(ProviderConfigChange, &failingCallback)
	AddHandler(successEventType, &successAPICallback)

	// provider association
	providerName := "providerA"

	client := NewClient(providerName)

	// client level handlers
	client.AddHandler(ProviderConfigChange, &failingCallback)
	client.AddHandler(successEventType, &successClientCallback)

	// trigger events manually
	go func() {
		eventing.Invoke(Event{
			ProviderName:         providerName,
			EventType:            successEventType,
			ProviderEventDetails: ProviderEventDetails{},
		})
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

// Requirement 5.3.1 If the provider's initialize function terminates normally, PROVIDER_READY handlers MUST run.
func TestEventHandler_InitOfProvider(t *testing.T) {
	t.Run("for default provider in global handler scope", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderReady, &callback)
		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("for default provider with unassociated client handler", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("clientWithNoProviderAssociation")
		client.AddHandler(ProviderReady, &callback)
		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("for named provider in client scope", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("someClient")
		client.AddHandler(ProviderReady, &callback)

		err := SetNamedProvider("someClient", provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("no callback for named provider with no associations", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderConfigChange})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("someClient")
		client.AddHandler(ProviderReady, &callback)

		err := SetNamedProvider("providerWithoutClient", provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			t.Errorf("event must have not emitted")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})
}

// Requirement 5.3.2 If the provider's initialize function terminates abnormally, PROVIDER_ERROR handlers MUST run.
func TestEventHandler_InitOfProviderError(t *testing.T) {
	t.Run("for default provider in global scope", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderError, &callback)
		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("for default provider with unassociated client handler", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("clientWithNoProviderAssociation")
		client.AddHandler(ProviderError, &callback)

		err := SetProvider(provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("for named provider in client scope", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("someClient")
		client.AddHandler(ProviderError, &callback)

		err := SetNamedProvider("someClient", provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("no callback for named provider with no associations", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		// callback
		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("provider")
		client.AddHandler(ProviderError, &callback)

		err := SetNamedProvider("someClient", provider)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-rsp:
			t.Errorf("event must have not emitted")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})
}

// Requirement 5.3.3 PROVIDER_READY handlers attached after the provider is already in a ready state MUST run immediately.
func TestEventHandler_ProviderReadiness(t *testing.T) {
	t.Run("for api level under default provider", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event),
		}

		readyEventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		err := SetProvider(readyEventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderReady, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback")
		}
	})

	t.Run("for domain associated handler", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event),
		}

		readyEventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		clientAssociation := "clientA"
		err := SetNamedProviderAndWait(clientAssociation, readyEventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := api.GetNamedClient(clientAssociation)
		client.AddHandler(ProviderReady, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback")
		}
	})

	t.Run("for unassociated handler from default", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event),
		}

		readyEventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		err := SetProviderAndWait(readyEventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := api.GetNamedClient("someClient")
		client.AddHandler(ProviderReady, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timedout waiting for ready state callback, but got none")
		}
	})

	t.Run("no event if provider is not ready", func(t *testing.T) {
		t.Cleanup(initSingleton)

		notReadyEventingProvider := struct {
			FeatureProvider
			StateHandler
			EventHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					return errors.New("some error from initialization")
				},
			},
			&ProviderEventing{},
		}

		err := SetProvider(notReadyEventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := api.GetNamedClient("someClient")
		client.AddHandler(ProviderReady, &callback)

		select {
		case <-rsp:
			t.Errorf("event must not emit for non ready provider")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})

	t.Run("no event if subscribed for some other event", func(t *testing.T) {
		t.Cleanup(initSingleton)

		readyEventingProvider := struct {
			FeatureProvider
			StateHandler
			EventHandler
		}{
			NoopProvider{},
			&stateHandlerForTests{
				initF: func(e EvaluationContext) error {
					return nil
				},
			},
			&ProviderEventing{},
		}

		err := SetProvider(readyEventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		client := NewClient("someClient")
		client.AddHandler(ProviderError, &callback)

		select {
		case <-rsp:
			t.Errorf("event must not emit for this handler")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})
}

// Requirement 5.3.3, Spec version 0.7.0: Handlers attached after the
// provider is already in the associated state, MUST run immediately
func TestEventHandler_HandlersRunImmediately(t *testing.T) {
	t.Run("ready handler runs when provider ready", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProvider(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderReady, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timed out waiting for callback")
		}
	})

	t.Run("error handler runs when provider error", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProvider(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderError, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timed out waiting for callback")
		}
	})

	t.Run("stale handler runs when provider stale", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderStale})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProvider(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 1)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderStale, &callback)

		select {
		case <-rsp:
			break
		case <-time.After(200 * time.Millisecond):
			t.Errorf("timed out waiting for callback")
		}
	})

	t.Run("non-ready handler does not run when provider ready", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProvider(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 3)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderError, &callback)
		AddHandler(ProviderStale, &callback)
		AddHandler(ProviderConfigChange, &callback)

		select {
		case <-rsp:
			t.Errorf("event must not emit for this handler")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})

	t.Run("non-error handler does not run when provider error", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderError})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProviderAndWait(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 3)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderReady, &callback)
		<-rsp // ignore first READY event which gets emitted after registration
		AddHandler(ProviderStale, &callback)
		AddHandler(ProviderConfigChange, &callback)

		// assert client transitioned to ERROR
		eventually(t, func() bool {
			return GetApiInstance().GetClient().State() == ErrorState
		}, time.Second, time.Millisecond*100, "")

		select {
		case <-rsp:
			t.Errorf("event must not emit for this handler")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})

	t.Run("non-stale handler does not run when provider stale", func(t *testing.T) {
		t.Cleanup(initSingleton)

		eventingImpl := &ProviderEventing{
			c: make(chan Event, 1),
		}
		eventingImpl.Invoke(Event{EventType: ProviderStale})

		provider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			eventingImpl,
		}

		if err := SetProviderAndWait(provider); err != nil {
			t.Fatal(err)
		}

		rsp := make(chan EventDetails, 3)
		callback := func(e EventDetails) {
			rsp <- e
		}

		AddHandler(ProviderReady, &callback)
		<-rsp // ignore first READY event which gets emitted after registration
		AddHandler(ProviderError, &callback)
		AddHandler(ProviderConfigChange, &callback)

		// assert client transitioned to STALE
		eventually(t, func() bool {
			return GetApiInstance().GetClient().State() == StaleState
		}, time.Second, time.Millisecond*100, "")

		select {
		case <-rsp:
			t.Errorf("event must not emit for this handler")
		case <-time.After(200 * time.Millisecond):
			break
		}
	})
}

// non-spec bound validations

func TestEventHandler_multiSubs(t *testing.T) {
	t.Cleanup(initSingleton)

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

	eventingImplOther := &ProviderEventing{
		c: make(chan Event, 1),
	}

	eventingProvideOther := struct {
		FeatureProvider
		EventHandler
	}{
		NoopProvider{},
		eventingImplOther,
	}

	// register for default and named providers
	err := SetProvider(eventingProvider)
	if err != nil {
		t.Fatal(err)
	}

	err = SetNamedProvider("clientA", eventingProvideOther)
	if err != nil {
		t.Fatal(err)
	}

	err = SetNamedProvider("clientB", eventingProvideOther)
	if err != nil {
		t.Fatal(err)
	}

	// register global and scoped callbacks
	rspGlobal := make(chan EventDetails, 1)
	globalF := func(e EventDetails) {
		rspGlobal <- e
	}

	AddHandler(ProviderStale, &globalF)

	rspClientA := make(chan EventDetails, 1)
	callbackA := func(e EventDetails) {
		rspClientA <- e
	}

	clientA := NewClient("clientA")
	clientA.AddHandler(ProviderStale, &callbackA)

	rspClientB := make(chan EventDetails, 1)
	callbackB := func(e EventDetails) {
		rspClientB <- e
	}

	clientB := NewClient("clientB")
	clientB.AddHandler(ProviderStale, &callbackB)

	emitDone := make(chan any)
	eventCount := 5

	// invoke events
	go func() {
		for range eventCount {
			eventingImpl.Invoke(Event{
				ProviderName:         "provider",
				EventType:            ProviderStale,
				ProviderEventDetails: ProviderEventDetails{},
			})

			eventingImplOther.Invoke(Event{
				ProviderName:         "providerOther",
				EventType:            ProviderStale,
				ProviderEventDetails: ProviderEventDetails{},
			})

			time.Sleep(100 * time.Millisecond)
		}

		emitDone <- ""
	}()

	// make sure events are received and count them
	globalEvents := make(chan string, 10)
	go func() {
		for rsp := range rspGlobal {
			globalEvents <- rsp.ProviderName
		}
	}()

	clientAEvents := make(chan string, 10)
	go func() {
		for rsp := range rspClientA {
			if rsp.ProviderName != "providerOther" {
				t.Errorf("incorrect event provider association, expected %s, got %s", "providerOther", rsp.ProviderName)
			}

			clientAEvents <- rsp.ProviderName
		}
	}()

	clientBEvents := make(chan string, 10)
	go func() {
		for rsp := range rspClientB {
			if rsp.ProviderName != "providerOther" {
				t.Errorf("incorrect event provider association, expected %s, got %s", "providerOther", rsp.ProviderName)
			}

			clientBEvents <- rsp.ProviderName
		}
	}()

	select {
	case <-time.After(1 * time.Minute):
		t.Errorf("test timedout")
	case <-emitDone:
		break
	}

	if len(globalEvents) != eventCount*2 || len(clientAEvents) != eventCount || len(clientBEvents) != eventCount {
		t.Error("event counts does not match with emitted event count")
	}
}

func TestEventHandler_1ToNMapping(t *testing.T) {
	t.Run("provider eventing must be subscribed only once", func(t *testing.T) {
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

		executor := newEventExecutor()

		err := executor.registerDefaultProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.activeSubscriptions) != 1 &&
			executor.activeSubscriptions[0].featureProvider != eventingProvider.FeatureProvider {
			t.Fatal("provider not added to active provider subscriptions")
		}

		err = executor.registerNamedEventingProvider("clientA", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		err = executor.registerNamedEventingProvider("clientB", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.activeSubscriptions) != 1 {
			t.Fatal("multiple provided in active subscriptions")
		}
	})

	t.Run("avoid unsubscribe from active providers - default and named", func(t *testing.T) {
		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		executor := newEventExecutor()

		err := executor.registerDefaultProvider(eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		err = executor.registerNamedEventingProvider("clientA", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		overridingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		err = executor.registerNamedEventingProvider("clientA", overridingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.activeSubscriptions) != 2 {
			t.Fatal("error with active provider subscriptions")
		}
	})

	t.Run("avoid unsubscribe from active providers - named only", func(t *testing.T) {
		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		executor := newEventExecutor()

		err := executor.registerNamedEventingProvider("clientA", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		err = executor.registerNamedEventingProvider("clientB", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		overridingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		err = executor.registerNamedEventingProvider("clientA", overridingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.activeSubscriptions) != 2 {
			t.Fatal("error with active provider subscriptions")
		}
	})

	t.Run("unbound providers must be removed from active subscriptions", func(t *testing.T) {
		eventingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		executor := newEventExecutor()

		err := executor.registerNamedEventingProvider("clientA", eventingProvider)
		if err != nil {
			t.Fatal(err)
		}

		overridingProvider := struct {
			FeatureProvider
			EventHandler
		}{
			NoopProvider{},
			&ProviderEventing{
				c: make(chan Event, 1),
			},
		}

		err = executor.registerNamedEventingProvider("clientA", overridingProvider)
		if err != nil {
			t.Fatal(err)
		}

		if len(executor.activeSubscriptions) != 1 &&
			executor.activeSubscriptions[0].featureProvider != overridingProvider.FeatureProvider {
			t.Fatal("error with active provider subscriptions")
		}
	})
}

// Contract tests - registration & removal
func TestEventHandler_Registration(t *testing.T) {
	t.Run("API handlers", func(t *testing.T) {
		executor := newEventExecutor()

		// Add multiple - ProviderReady
		executor.AddHandler(ProviderReady, &h1)
		executor.AddHandler(ProviderReady, &h2)
		executor.AddHandler(ProviderReady, &h3)
		executor.AddHandler(ProviderReady, &h4)

		// Add multiple - ProviderError
		executor.AddHandler(ProviderError, &h2)
		executor.AddHandler(ProviderError, &h2)

		// Add single types
		executor.AddHandler(ProviderStale, &h3)
		executor.AddHandler(ProviderConfigChange, &h4)

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
		executor := newEventExecutor()

		// Add multiple - client a
		executor.AddClientHandler("a", ProviderReady, &h1)
		executor.AddClientHandler("a", ProviderReady, &h2)
		executor.AddClientHandler("a", ProviderReady, &h3)
		executor.AddClientHandler("a", ProviderReady, &h4)

		// Add single for rest of the client
		executor.AddClientHandler("b", ProviderError, &h2)
		executor.AddClientHandler("c", ProviderStale, &h3)
		executor.AddClientHandler("d", ProviderConfigChange, &h4)

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
		executor := newEventExecutor()

		// Add multiple - ProviderReady
		executor.AddHandler(ProviderReady, &h1)
		executor.AddHandler(ProviderReady, &h2)
		executor.AddHandler(ProviderReady, &h3)
		executor.AddHandler(ProviderReady, &h4)

		// Add single types
		executor.AddHandler(ProviderError, &h2)
		executor.AddHandler(ProviderStale, &h3)
		executor.AddHandler(ProviderConfigChange, &h4)

		// removal
		executor.RemoveHandler(ProviderReady, &h1)
		executor.RemoveHandler(ProviderError, &h2)
		executor.RemoveHandler(ProviderStale, &h3)
		executor.RemoveHandler(ProviderConfigChange, &h4)

		readyLen := len(executor.apiRegistry[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events, but got %d", 3, readyLen)
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], EventCallback(&h2)) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], EventCallback(&h3)) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.apiRegistry[ProviderReady], EventCallback(&h4)) {
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
		executor := newEventExecutor()

		// Add multiple - client a
		executor.AddClientHandler("a", ProviderReady, &h1)
		executor.AddClientHandler("a", ProviderReady, &h2)
		executor.AddClientHandler("a", ProviderReady, &h3)
		executor.AddClientHandler("a", ProviderReady, &h4)

		// Add single
		executor.AddClientHandler("b", ProviderError, &h2)
		executor.AddClientHandler("c", ProviderStale, &h3)
		executor.AddClientHandler("d", ProviderConfigChange, &h4)

		// removal
		executor.RemoveClientHandler("a", ProviderReady, &h1)
		executor.RemoveClientHandler("b", ProviderError, &h2)
		executor.RemoveClientHandler("c", ProviderStale, &h3)
		executor.RemoveClientHandler("d", ProviderConfigChange, &h4)

		readyLen := len(executor.scopedRegistry["a"].callbacks[ProviderReady])
		if readyLen != 3 {
			t.Errorf("expected %d events in client a, but got %d", 3, readyLen)
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], EventCallback(&h2)) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], EventCallback(&h3)) {
			t.Errorf("expected callback to be present")
		}

		if !slices.Contains(executor.scopedRegistry["a"].callbacks[ProviderReady], EventCallback(&h4)) {
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
		executor.RemoveClientHandler("non-existing", ProviderReady, &h1)
		executor.RemoveClientHandler("b", ProviderReady, &h1)
	})

	t.Run("remove handlers that were not added", func(t *testing.T) {
		executor := newEventExecutor()

		// removal of non-added handlers shall not panic
		executor.RemoveHandler(ProviderReady, &h1)
		executor.RemoveClientHandler("a", ProviderReady, &h1)
	})
}
