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
	providerShutdownHook map[string]chan interface{}
	apiRegistry          map[EventType][]EventCallBack
	scopedRegistry       map[string]scopedCallback
	logger               logr.Logger
	mu                   sync.Mutex
}

func newEventHandler(logger logr.Logger) eventExecutor {
	return eventExecutor{
		providerShutdownHook: map[string]chan interface{}{},
		apiRegistry:          map[EventType][]EventCallBack{},
		scopedRegistry:       map[string]scopedCallback{},
		logger:               logger,
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

func (e *eventExecutor) registerEventingProvider(provider FeatureProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := provider.(EventHandler)
	if !ok {
		return
	}

	// register shutdown semaphore
	sem := make(chan interface{})
	e.providerShutdownHook[provider.Metadata().Name] = sem

	go func() {
		for {
			select {
			case event := <-v.EventChannel():
				err := e.triggerEvent(event)
				if err != nil {
					e.logger.Error(err, fmt.Sprintf("error handling event type: %s", event.EventType))
				}
			case <-sem:
				return
			}
		}
	}()
}

func (e *eventExecutor) triggerEvent(event Event) error {
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

		// then run Client handlers
		// Note - we must only run associated client handlers of the provider
		associateClientRegistry := e.scopedRegistry[event.ProviderName]
		for _, c := range associateClientRegistry.callbacks[event.EventType] {
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
		client: clientName,
		ProviderEventDetails: ProviderEventDetails{
			Message:       event.Message,
			FlagChanges:   event.FlagChanges,
			EventMetadata: event.EventMetadata,
		},
	})
}

func (e *eventExecutor) unregisterEventingProvider(provider FeatureProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := provider.(EventHandler)
	if !ok {
		return nil
	}

	sem := e.providerShutdownHook[provider.Metadata().Name]
	if sem == nil {
		return nil
	}

	delete(e.providerShutdownHook, provider.Metadata().Name)

	// wait for completion or timeout
	select {
	case sem <- "":
		return nil
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("event handler of provider %s timeout with exiting", provider.Metadata().Name)
	}
}
