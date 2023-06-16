package openfeature

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

// event handler is registry to connect API and Client event handlers to Providers

type clientHolder struct {
	name   string
	holder map[EventType][]EventCallBack
}

func newClientHolder(client string) clientHolder {
	return clientHolder{
		name:   client,
		holder: map[EventType][]EventCallBack{},
	}
}

type eventHandler struct {
	providerShutdownHook map[string]chan interface{}
	apiRegistry          map[EventType][]EventCallBack
	clientRegistry       map[string]clientHolder
	mu                   sync.Mutex
}

func newEventHandler() eventHandler {
	return eventHandler{
		providerShutdownHook: map[string]chan interface{}{},
		apiRegistry:          map[EventType][]EventCallBack{},
		clientRegistry:       map[string]clientHolder{},
	}
}

func (e *eventHandler) registerApiHandler(t EventType, c EventCallBack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.apiRegistry[t] == nil {
		e.apiRegistry[t] = []EventCallBack{c}
	} else {
		e.apiRegistry[t] = append(e.apiRegistry[t], c)
	}
}

func (e *eventHandler) removeApiHandler(t EventType, c EventCallBack) {
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

func (e *eventHandler) registerClientHandler(name string, t EventType, c EventCallBack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.clientRegistry[name]
	if !ok {
		e.clientRegistry[name] = newClientHolder(name)
	}

	v := e.clientRegistry[name]

	if v.holder[t] == nil {
		v.holder[t] = []EventCallBack{c}
	} else {
		v.holder[t] = append(v.holder[t], c)
	}
}

func (e *eventHandler) removeClientHandler(name string, t EventType, c EventCallBack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.clientRegistry[name]
	if !ok {
		// nothing to remove
		return
	}

	entrySlice := e.clientRegistry[name].holder[t]
	if entrySlice == nil {
		// nothing to remove
		return
	}

	for i, f := range entrySlice {
		if f == c {
			entrySlice = append(entrySlice[:i], entrySlice[i+1:]...)
		}
	}

	e.clientRegistry[name].holder[t] = entrySlice
}

func (e *eventHandler) registerEventingProvider(provider FeatureProvider) {
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
				// todo loggign or ignore errors here
				err := e.triggerEvent(event)
				if err != nil {
					return
				}
			case <-sem:
				return
			}
		}
	}()
}

func (e *eventHandler) triggerEvent(t Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		// first run API handlers
		for _, c := range e.apiRegistry[t.EventType] {
			f := *c
			f(EventDetails{
				ProviderEventDetails: ProviderEventDetails{
					Message:       t.Message,
					FlagChanges:   t.FlagChanges,
					EventMetadata: t.EventMetadata,
				},
			})
		}

		// then run Client handlers
		for _, client := range e.clientRegistry {
			for _, c := range client.holder[t.EventType] {
				f := *c
				f(EventDetails{
					client: client.name,
					ProviderEventDetails: ProviderEventDetails{
						Message:       t.Message,
						FlagChanges:   t.FlagChanges,
						EventMetadata: t.EventMetadata,
					},
				})
			}
		}

		return nil
	})

	// wait for completion or timeout
	select {
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("event handlers timeout")
	case <-gCtx.Done():
		return nil
	}
}

func (e *eventHandler) unregisterEventingProvider(provider FeatureProvider) error {
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
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("event handler of provider %s timeout with exiting", provider.Metadata().Name)
	}
}
