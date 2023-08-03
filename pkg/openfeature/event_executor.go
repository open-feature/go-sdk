package openfeature

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
)

// event executor is a registry to connect API and Client event handlers to Providers

// eventExecutor handles events emitted from FeatureProvider. It follows a pub-sub model based on channels.
// Emitted events are written to eventChan. This model is chosen so that events can be triggered from subscribed
// feature provider as well as from API(ex:- for initialization events).
// Usage of channels help with concurrency and adhere to the principal of sharing memory by communication.
type eventExecutor struct {
	defaultProviderReference providerReference
	namedProviderReference   map[string]providerReference
	activeSubscriptions      []providerReference
	apiRegistry              map[EventType][]EventCallback
	scopedRegistry           map[string]scopedCallback
	logger                   logr.Logger
	eventChan                chan eventPayload
	once                     sync.Once
	mu                       sync.Mutex
}

func newEventExecutor(logger logr.Logger) *eventExecutor {
	executor := eventExecutor{
		namedProviderReference: map[string]providerReference{},
		activeSubscriptions:    []providerReference{},
		apiRegistry:            map[EventType][]EventCallback{},
		scopedRegistry:         map[string]scopedCallback{},
		logger:                 logger,
		eventChan:              make(chan eventPayload, 1),
	}

	executor.startEventListener()
	return &executor
}

// scopedCallback is a helper struct to hold client name associated callbacks.
// Here, the scope correlates to the client and provider name
type scopedCallback struct {
	scope     string
	callbacks map[EventType][]EventCallback
}

type eventPayload struct {
	event   Event
	handler FeatureProvider
}

func newScopedCallback(client string) scopedCallback {
	return scopedCallback{
		scope:     client,
		callbacks: map[EventType][]EventCallback{},
	}
}

// providerReference is a helper struct to store FeatureProvider with EventHandler capability along with their
// shutdown semaphore
type providerReference struct {
	featureProvider   FeatureProvider
	shutdownSemaphore chan interface{}
}

// updateLogger updates the executor's logger
func (e *eventExecutor) updateLogger(l logr.Logger) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger = l
}

// registerApiHandler adds an API(global) level handler
func (e *eventExecutor) registerApiHandler(t EventType, c EventCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.apiRegistry[t] == nil {
		e.apiRegistry[t] = []EventCallback{c}
	} else {
		e.apiRegistry[t] = append(e.apiRegistry[t], c)
	}

	e.emitOnRegistration(e.defaultProviderReference, t, c)
}

// removeApiHandler removes an API(global) level handler
func (e *eventExecutor) removeApiHandler(t EventType, c EventCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entrySlice, ok := e.apiRegistry[t]
	if !ok {
		// nothing to remove
		return
	}

	for i, f := range entrySlice {
		if f == c {
			entrySlice = append(entrySlice[:i], entrySlice[i+1:]...)
		}
	}

	e.apiRegistry[t] = entrySlice
}

// registerClientHandler registers a client level handler
func (e *eventExecutor) registerClientHandler(clientName string, t EventType, c EventCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.scopedRegistry[clientName]
	if !ok {
		e.scopedRegistry[clientName] = newScopedCallback(clientName)
	}

	registry := e.scopedRegistry[clientName]

	if registry.callbacks[t] == nil {
		registry.callbacks[t] = []EventCallback{c}
	} else {
		registry.callbacks[t] = append(registry.callbacks[t], c)
	}

	reference, ok := e.namedProviderReference[clientName]
	if !ok {
		// fallback to default
		reference = e.defaultProviderReference
	}

	e.emitOnRegistration(reference, t, c)
}

// removeClientHandler removes a client level handler
func (e *eventExecutor) removeClientHandler(name string, t EventType, c EventCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.scopedRegistry[name]
	if !ok {
		// nothing to remove
		return
	}

	entrySlice := e.scopedRegistry[name].callbacks[t]
	if entrySlice == nil {
		// nothing to remove
		return
	}

	for i, f := range entrySlice {
		if f == c {
			entrySlice = append(entrySlice[:i], entrySlice[i+1:]...)
		}
	}

	e.scopedRegistry[name].callbacks[t] = entrySlice
}

// emitOnRegistration fulfils the spec requirement to fire ready events if associated provider is ready
func (e *eventExecutor) emitOnRegistration(providerReference providerReference, t EventType, c EventCallback) {
	if t != ProviderReady {
		return
	}

	s, ok := (providerReference.featureProvider).(StateHandler)
	if !ok {
		// not a state handler, hence ignore state emitting
		return
	}

	if s.Status() == ReadyState {
		(*c)(EventDetails{
			providerName: (providerReference.featureProvider).Metadata().Name,
			ProviderEventDetails: ProviderEventDetails{
				Message: "provider is at ready state",
			},
		})
	}
}

// registerDefaultProvider registers the default FeatureProvider and remove the old default provider if available
func (e *eventExecutor) registerDefaultProvider(provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// register shutdown semaphore for new default provider
	sem := make(chan interface{})

	newProvider := providerReference{
		featureProvider:   provider,
		shutdownSemaphore: sem,
	}

	oldProvider := e.defaultProviderReference
	e.defaultProviderReference = newProvider

	return e.startListeningAndShutdownOld(newProvider, oldProvider)
}

// registerNamedEventingProvider registers a named FeatureProvider and remove event listener for old named provider
func (e *eventExecutor) registerNamedEventingProvider(associatedClient string, provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// register shutdown semaphore for new named provider
	sem := make(chan interface{})

	newProvider := providerReference{
		featureProvider:   provider,
		shutdownSemaphore: sem,
	}

	oldProvider := e.namedProviderReference[associatedClient]
	e.namedProviderReference[associatedClient] = newProvider

	return e.startListeningAndShutdownOld(newProvider, oldProvider)
}

// startListeningAndShutdownOld is a helper to start concurrent listening to new provider events and  invoke shutdown
// hook of the old provider if it's not bound by another subscription
func (e *eventExecutor) startListeningAndShutdownOld(newProvider providerReference, oldReference providerReference) error {

	// check if this provider already actively handled - 1:N binding capability
	if !isRunning(newProvider, e.activeSubscriptions) {
		e.activeSubscriptions = append(e.activeSubscriptions, newProvider)

		go func() {
			v, ok := newProvider.featureProvider.(EventHandler)
			if !ok {
				return
			}

			// event handling of the new feature provider
			for {
				select {
				case event := <-v.EventChannel():
					e.eventChan <- eventPayload{
						event:   event,
						handler: newProvider.featureProvider,
					}
				case <-newProvider.shutdownSemaphore:
					return
				}
			}
		}()
	}

	// shutdown old provider handling

	// check if this provider is still bound - 1:N binding capability
	if isBound(oldReference, e.defaultProviderReference, maps.Values(e.namedProviderReference)) {
		return nil
	}

	// drop from active references
	for i, r := range e.activeSubscriptions {
		if reflect.DeepEqual(oldReference.featureProvider, r.featureProvider) {
			e.activeSubscriptions = append(e.activeSubscriptions[:i], e.activeSubscriptions[i+1:]...)
		}
	}

	_, ok := oldReference.featureProvider.(EventHandler)
	if !ok {
		// no shutdown for non event handling provider
		return nil
	}

	// avoid shutdown lockouts
	select {
	case oldReference.shutdownSemaphore <- "":
		return nil
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("old event handler %s timeout waiting for handler shutdown",
			oldReference.featureProvider.Metadata().Name)
	}
}

// startEventListener trigger the event listening of this executor
func (e *eventExecutor) startEventListener() {
	e.once.Do(func() {
		go func() {
			for payload := range e.eventChan {
				e.triggerEvent(payload.event, payload.handler)
			}
		}()
	})
}

// triggerEvent performs the actual event handling
func (e *eventExecutor) triggerEvent(event Event, handler FeatureProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// first run API handlers
	for _, c := range e.apiRegistry[event.EventType] {
		e.executeHandler(*c, event)
	}

	// then run client handlers
	for name, reference := range e.namedProviderReference {
		if reference.featureProvider != handler {
			// unassociated client, continue to next
			continue
		}

		for _, c := range e.scopedRegistry[name].callbacks[event.EventType] {
			e.executeHandler(*c, event)
		}
	}

	if e.defaultProviderReference.featureProvider != handler {
		return
	}

	// handling the default provider - invoke default provider bound (no provider associated) handlers by filtering
	for clientName, registry := range e.scopedRegistry {
		if _, ok := e.namedProviderReference[clientName]; ok {
			// association exist, skip and check next
			continue
		}

		for _, c := range registry.callbacks[event.EventType] {
			e.executeHandler(*c, event)
		}
	}

}

// executeHandler is a helper which performs the actual invocation of the callback
func (e *eventExecutor) executeHandler(f func(details EventDetails), event Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				e.logger.Info("recovered from a panic")
			}
		}()

		f(EventDetails{
			providerName: event.ProviderName,
			ProviderEventDetails: ProviderEventDetails{
				Message:       event.Message,
				FlagChanges:   event.FlagChanges,
				EventMetadata: event.EventMetadata,
			},
		})
	}()
}

// isRunning is a helper till we bump to the latest go version with slices.contains support
func isRunning(provider providerReference, activeProviders []providerReference) bool {
	for _, activeProvider := range activeProviders {
		if reflect.DeepEqual(activeProvider.featureProvider, provider.featureProvider) {
			return true
		}
	}

	return false
}

// isRunning is a helper to check if given provider is already in use
func isBound(provider providerReference, defaultProvider providerReference, namedProviders []providerReference) bool {
	if reflect.DeepEqual(provider.featureProvider, defaultProvider.featureProvider) {
		return true
	}

	for _, namedProvider := range namedProviders {
		if reflect.DeepEqual(provider.featureProvider, namedProvider.featureProvider) {
			return true
		}
	}

	return false
}
