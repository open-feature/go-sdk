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
	apiRegistry              map[EventType][]EventCallBack
	scopedRegistry           map[string]scopedCallback
	logger                   logr.Logger
	mu                       sync.Mutex
}

func newEventExecutor(logger logr.Logger) eventExecutor {
	return eventExecutor{
		namedProviderReference: map[string]*providerReference{},
		apiRegistry:            map[EventType][]EventCallBack{},
		scopedRegistry:         map[string]scopedCallback{},
		logger:                 logger,
	}
}

// scopedCallback is a helper struct to hold scope specific callbacks.
// Here, the scope correlates to the client and provider name
type scopedCallback struct {
	scope     string
	callbacks map[EventType][]EventCallBack
}

func newScopedCallback(client string) scopedCallback {
	return scopedCallback{
		scope:     client,
		callbacks: map[EventType][]EventCallBack{},
	}
}

type providerReference struct {
	eventHandler      *EventHandler
	name              string
	isDefault         bool
	shutdownSemaphore chan interface{}
}

func (e *eventExecutor) updateLogger(l logr.Logger) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger = l
}

func (e *eventExecutor) registerApiHandler(t EventType, c EventCallBack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.apiRegistry[t] == nil {
		e.apiRegistry[t] = []EventCallBack{c}
	} else {
		e.apiRegistry[t] = append(e.apiRegistry[t], c)
	}
}

func (e *eventExecutor) removeApiHandler(t EventType, c EventCallBack) {
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

func (e *eventExecutor) registerClientHandler(clientName string, t EventType, c EventCallBack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.scopedRegistry[clientName]
	if !ok {
		e.scopedRegistry[clientName] = newScopedCallback(clientName)
	}

	callback := e.scopedRegistry[clientName]

	if callback.callbacks[t] == nil {
		callback.callbacks[t] = []EventCallBack{c}
	} else {
		callback.callbacks[t] = append(callback.callbacks[t], c)
	}
}

func (e *eventExecutor) removeClientHandler(name string, t EventType, c EventCallBack) {
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

// registerDefaultProvider register the default FeatureProvider and remove event listener for old default provider
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
		isDefault:         true,
		shutdownSemaphore: sem,
	}

	e.defaultProviderReference = newProvider
	return e.listenAndShutdown(newProvider, oldProvider)
}

// registerNamedEventingProvider register a named FeatureProvider and remove event listener for old named provider
func (e *eventExecutor) registerNamedEventingProvider(name string, provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := provider.(EventHandler)
	if !ok {
		return nil
	}

	oldProvider := e.namedProviderReference[name]

	// register shutdown semaphore for new named provider
	sem := make(chan interface{})

	newProvider := &providerReference{
		eventHandler:      &v,
		name:              name,
		shutdownSemaphore: sem,
	}

	e.namedProviderReference[name] = newProvider
	return e.listenAndShutdown(newProvider, oldProvider)
}

func (e *eventExecutor) listenAndShutdown(newProvider *providerReference, oldReference *providerReference) error {
	go func(newProvider *providerReference, oldReference *providerReference) {
		for {
			select {
			case event := <-(*newProvider.eventHandler).EventChannel():
				err := e.triggerEvent(event, newProvider.name, newProvider.isDefault)
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

func (e *eventExecutor) triggerEvent(event Event, providerName string, isDefault bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		// first run API handlers
		for _, c := range e.apiRegistry[event.EventType] {
			e.executeHandler(*c, "", event)
		}

		// then run Client handlers for name association

		// first direct associates
		associateClientRegistry := e.scopedRegistry[providerName]
		for _, c := range associateClientRegistry.callbacks[event.EventType] {
			e.executeHandler(*c, associateClientRegistry.scope, event)
		}

		if !isDefault {
			return nil
		}

		// handling the default provider - invoke default provider bound handlers by filtering
		var defaultHandlers []EventCallBack

		for clientName, registry := range e.scopedRegistry {
			if _, ok := e.namedProviderReference[clientName]; !ok {
				defaultHandlers = append(defaultHandlers, registry.callbacks[event.EventType]...)
			}
		}

		for _, c := range defaultHandlers {
			e.executeHandler(*c, associateClientRegistry.scope, event)
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

func (e *eventExecutor) executeHandler(f func(details EventDetails), clientName string, event Event) {
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
