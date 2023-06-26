package openfeature

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
)

// event executor is a registry to connect API and Client event handlers to Providers

// handlerExecutionTime defines the maximum time event handler will wait for its handlers to complete
const handlerExecutionTime = 500 * time.Millisecond

type eventExecutor struct {
	defaultProviderReference *providerReference
	namedProviderReference   map[string]*providerReference
	apiRegistry              map[EventType][]EventCallback
	scopedRegistry           map[string]scopedCallback
	logger                   logr.Logger
	mu                       sync.Mutex
}

func newEventExecutor(logger logr.Logger) eventExecutor {
	return eventExecutor{
		namedProviderReference: map[string]*providerReference{},
		apiRegistry:            map[EventType][]EventCallback{},
		scopedRegistry:         map[string]scopedCallback{},
		logger:                 logger,
	}
}

// scopedCallback is a helper struct to hold client name associated callbacks.
// Here, the scope correlates to the client and provider name
type scopedCallback struct {
	scope     string
	callbacks map[EventType][]EventCallback
}

func newScopedCallback(client string) scopedCallback {
	return scopedCallback{
		scope:     client,
		callbacks: map[EventType][]EventCallback{},
	}
}

type providerReference struct {
	eventHandler          *EventHandler
	metadata              Metadata
	clientNameAssociation string
	isDefault             bool
	shutdownSemaphore     chan interface{}
}

// updateLogger updates the executor's logger
func (e *eventExecutor) updateLogger(l logr.Logger) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger = l
}

// registerApiHandler add API(global) level handler
func (e *eventExecutor) registerApiHandler(t EventType, c EventCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.apiRegistry[t] == nil {
		e.apiRegistry[t] = []EventCallback{c}
	} else {
		e.apiRegistry[t] = append(e.apiRegistry[t], c)
	}
}

// removeApiHandler remove API(global) level handler
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

// registerClientHandler register client level handler
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

	// fulfil spec requirement to fire ready events if associated provider is ready
	if t != ProviderReady {
		return
	}

	provider, ok := e.namedProviderReference[clientName]
	if !ok {
		return
	}

	s, ok := (*provider.eventHandler).(StateHandler)
	if !ok {
		return
	}

	if s.Status() == ReadyState {
		(*c)(EventDetails{
			provider: provider.metadata.Name,
			ProviderEventDetails: ProviderEventDetails{
				Message: "provider is at ready state",
			},
		})
	}
}

// removeClientHandler remove client level handler
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

// registerDefaultProvider register the default FeatureProvider and remove the old default provider if available
func (e *eventExecutor) registerDefaultProvider(provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := provider.(EventHandler)
	if !ok {
		return nil
	}

	oldProvider := e.defaultProviderReference

	// register shutdown semaphore for new default provider
	sem := make(chan interface{})

	newProvider := &providerReference{
		eventHandler:      &v,
		metadata:          provider.Metadata(),
		isDefault:         true,
		shutdownSemaphore: sem,
	}

	e.defaultProviderReference = newProvider
	return e.listenAndShutdown(newProvider, oldProvider)
}

// registerNamedEventingProvider register a named FeatureProvider and remove event listener for old named provider
func (e *eventExecutor) registerNamedEventingProvider(associatedClient string, provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := provider.(EventHandler)
	if !ok {
		return nil
	}

	oldProvider := e.namedProviderReference[associatedClient]

	// register shutdown semaphore for new named provider
	sem := make(chan interface{})

	newProvider := &providerReference{
		eventHandler:          &v,
		clientNameAssociation: associatedClient,
		metadata:              provider.Metadata(),
		shutdownSemaphore:     sem,
	}

	e.namedProviderReference[associatedClient] = newProvider
	return e.listenAndShutdown(newProvider, oldProvider)
}

// listenAndShutdown is a helper to start concurrent listening to new provider events and invoke shutdown hook of the
// old provider
func (e *eventExecutor) listenAndShutdown(newProvider *providerReference, oldReference *providerReference) error {
	go func(newProvider *providerReference, oldReference *providerReference) {
		for {
			select {
			case event := <-(*newProvider.eventHandler).EventChannel():
				err := e.triggerEvent(event, newProvider.clientNameAssociation, newProvider.isDefault)
				if err != nil {
					e.logger.Error(err, fmt.Sprintf("error handling event type: %s", event.EventType))
				}
			case <-newProvider.shutdownSemaphore:
				return
			}
		}
	}(newProvider, oldReference)

	// shutdown old provider handling
	if oldReference == nil {
		return nil
	}

	select {
	case oldReference.shutdownSemaphore <- "":
		return nil
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("event handler timeout waiting for handler shutdown")
	}
}

// triggerEvent performs the actual event handling
func (e *eventExecutor) triggerEvent(event Event, clientNameAssociation string, isDefault bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		// first run API handlers
		for _, c := range e.apiRegistry[event.EventType] {
			e.executeHandler(*c, event)
		}

		// then run Client handlers for name association

		// first direct associates
		associateClientRegistry := e.scopedRegistry[clientNameAssociation]
		for _, c := range associateClientRegistry.callbacks[event.EventType] {
			e.executeHandler(*c, event)
		}

		if !isDefault {
			return nil
		}

		// handling the default provider - invoke default provider bound handlers by filtering

		var defaultHandlers []EventCallback

		for clientName, registry := range e.scopedRegistry {
			if _, ok := e.namedProviderReference[clientName]; !ok {
				defaultHandlers = append(defaultHandlers, registry.callbacks[event.EventType]...)
			}
		}

		for _, c := range defaultHandlers {
			e.executeHandler(*c, event)
		}

		return nil
	})

	// wait for completion or timeout
	select {
	case <-time.After(handlerExecutionTime):
		return fmt.Errorf("event handlers timeout")
	case <-gCtx.Done():
		return nil
	}
}

// executeHandler is a helper which performs the actual invocation of the callback
func (e *eventExecutor) executeHandler(f func(details EventDetails), event Event) {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Info("recovered from a panic")
		}
	}()

	f(EventDetails{
		provider: event.ProviderName,
		ProviderEventDetails: ProviderEventDetails{
			Message:       event.Message,
			FlagChanges:   event.FlagChanges,
			EventMetadata: event.EventMetadata,
		},
	})
}
