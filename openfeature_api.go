package openfeature

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"
)

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	defaultProvider FeatureProvider
	namedProviders  map[string]FeatureProvider
	hks             []Hook
	evalCtx         EvaluationContext
	eventExecutor   *eventExecutor
	mu              sync.RWMutex
}

// newEvaluationAPI is a helper to generate an API. Used internally
func newEvaluationAPI(eventExecutor *eventExecutor) *evaluationAPI {
	return &evaluationAPI{
		defaultProvider: NoopProvider{},
		namedProviders:  map[string]FeatureProvider{},
		hks:             []Hook{},
		evalCtx:         EvaluationContext{},
		mu:              sync.RWMutex{},
		eventExecutor:   eventExecutor,
	}
}

func (api *evaluationAPI) SetProvider(ctx context.Context, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}
	api.setProviderWithContext(ctx, provider)
	return nil
}

func (api *evaluationAPI) SetProviderAndWait(ctx context.Context, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}
	initCh := api.setProviderWithContext(ctx, provider)
	return <-initCh
}

// setProviderWithContext sets the default FeatureProvider of the evaluationAPI with context-aware initialization.
func (api *evaluationAPI) setProviderWithContext(ctx context.Context, provider FeatureProvider) <-chan error {
	api.mu.Lock()
	defer api.mu.Unlock()

	oldProvider := api.defaultProvider
	api.defaultProvider = provider

	api.eventExecutor.registerDefaultProvider(provider)

	api.shutdownOld(ctx, oldProvider)

	return api.initNew(ctx, "", provider)
}

func (api *evaluationAPI) setNamedProviderWithContext(ctx context.Context, clientName string, provider FeatureProvider) <-chan error {
	api.mu.Lock()
	defer api.mu.Unlock()

	oldProvider := api.namedProviders[clientName]
	api.namedProviders[clientName] = provider

	api.eventExecutor.registerNamedEventingProvider(clientName, provider)

	api.shutdownOld(ctx, oldProvider)

	return api.initNew(ctx, clientName, provider)
}

// SetNamedProvider sets a provider with client name. Returns an error if FeatureProvider is nil.
func (api *evaluationAPI) SetNamedProvider(ctx context.Context, clientName string, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}
	api.setNamedProviderWithContext(ctx, clientName, provider)
	return nil
}

// SetNamedProviderAndWait sets a provider with client name and waits for initialization to complete.
func (api *evaluationAPI) SetNamedProviderAndWait(ctx context.Context, clientName string, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}
	initCh := api.setNamedProviderWithContext(ctx, clientName, provider)
	return <-initCh
}

func (api *evaluationAPI) initNew(ctx context.Context, clientName string, newProvider FeatureProvider) <-chan error {
	errCh := make(chan error, 1)

	go func(executor *eventExecutor, evalCtx EvaluationContext, ctx context.Context, provider FeatureProvider, clientName string) {
		event, err := initializerWithContext(ctx, provider, evalCtx)
		executor.states.Store(clientName, stateFromEventOrError(event, err))
		executor.triggerEvent(event, provider)
		if err != nil {
			err = fmt.Errorf("failed to initialize named provider %q for domain %q: %w", provider.Metadata().Name, clientName, err)
		}
		errCh <- err
	}(api.eventExecutor, api.evalCtx, ctx, newProvider, clientName)

	return errCh
}

func (api *evaluationAPI) shutdownOld(ctx context.Context, oldProvider FeatureProvider) {
	v, ok := oldProvider.(StateHandler)

	// oldProvider can be nil or without state handling capability
	if oldProvider == nil || !ok {
		return
	}

	namedProviders := slices.Collect(maps.Values(api.namedProviders))

	// check for multiple bindings
	if oldProvider == api.defaultProvider || slices.Contains(namedProviders, oldProvider) {
		return
	}

	go func(forShutdown StateHandler, parentCtx context.Context) {
		if err := forShutdown.Shutdown(parentCtx); err != nil {
			slog.Error("async provider shutdown failed", slog.Any("error", err))
		}
	}(v, ctx)
}

// GetProviderMetadata returns the default FeatureProvider's metadata
func (api *evaluationAPI) GetProviderMetadata() Metadata {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.defaultProvider.Metadata()
}

// GetNamedProviderMetadata returns the default FeatureProvider's metadata
func (api *evaluationAPI) GetNamedProviderMetadata(name string) Metadata {
	api.mu.RLock()
	defer api.mu.RUnlock()

	provider, ok := api.namedProviders[name]
	if !ok {
		return ProviderMetadata()
	}

	return provider.Metadata()
}

// GetNamedProviders returns named providers map.
func (api *evaluationAPI) GetNamedProviders() map[string]FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.namedProviders
}

// GetClient returns a IClient bound to the default provider
func (api *evaluationAPI) GetClient() *Client {
	return newClient("", api, api.eventExecutor)
}

// GetNamedClient returns a IClient bound to the given named provider
func (api *evaluationAPI) GetNamedClient(clientName string) *Client {
	return newClient(clientName, api, api.eventExecutor)
}

func (api *evaluationAPI) SetEvaluationContext(evalCtx EvaluationContext) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.evalCtx = evalCtx
}

func (api *evaluationAPI) AddHooks(hooks ...Hook) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.hks = append(api.hks, hooks...)
}

func (api *evaluationAPI) GetHooks() []Hook {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.hks
}

// AddHandler allows to add API level event handler
func (api *evaluationAPI) AddHandler(eventType EventType, callback EventCallback) {
	api.eventExecutor.AddHandler(eventType, callback)
}

// RemoveHandler allows to remove API level event handler
func (api *evaluationAPI) RemoveHandler(eventType EventType, callback EventCallback) {
	api.eventExecutor.RemoveHandler(eventType, callback)
}

// Shutdown calls shutdown on all registered providers.
// Returns an error if any provider shutdown fails or if context is cancelled during shutdown.
func (api *evaluationAPI) Shutdown(ctx context.Context) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	var errs []error

	// Shutdown default provider
	if api.defaultProvider != nil {
		if stateHandler, ok := api.defaultProvider.(StateHandler); ok {
			if err := stateHandler.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("default provider shutdown failed: %w", err))
			}
		}
	}

	// Shutdown all named providers
	for name, provider := range api.namedProviders {
		if stateHandler, ok := provider.(StateHandler); ok {
			if err := stateHandler.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("named provider %q shutdown failed: %w", name, err))
			}
		}
	}

	return errors.Join(errs...)
}

// ForEvaluation is a helper to retrieve transaction scoped operators.
// Returns the default FeatureProvider if no provider mapping exist for the given client name.
func (api *evaluationAPI) ForEvaluation(clientName string) (FeatureProvider, []Hook, EvaluationContext) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	var provider FeatureProvider

	provider = api.namedProviders[clientName]
	if provider == nil {
		provider = api.defaultProvider
	}

	return provider, api.hks, api.evalCtx
}

// GetProvider returns the default FeatureProvider
func (api *evaluationAPI) GetProvider() FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.defaultProvider
}

// initializerWithContext is a context-aware helper to execute provider initialization and generate appropriate event for the initialization
// If the provider implements ContextAwareStateHandler, InitWithContext is called; otherwise, Init is called for backward compatibility.
// It also returns an error if the initialization resulted in an error or if the context is cancelled.
func initializerWithContext(ctx context.Context, provider FeatureProvider, evalCtx EvaluationContext) (Event, error) {
	event := Event{
		ProviderName: provider.Metadata().Name,
		EventType:    ProviderReady,
		ProviderEventDetails: ProviderEventDetails{
			Message: "Provider initialization successful",
		},
	}
	var err error

	if contextHandler, ok := provider.(StateHandler); ok {
		err = contextHandler.Init(ContextWithEvaluationContext(ctx, evalCtx))
		if err != nil {
			event.EventType = ProviderError

			// Check for specific provider initialization errors first
			var initErr *ProviderInitError
			if errors.As(err, &initErr) {
				event.ErrorCode = initErr.ErrorCode
				event.Message = initErr.Message
			} else if errors.Is(err, context.Canceled) {
				event.Message = "Provider initialization cancelled"
			} else if errors.Is(err, context.DeadlineExceeded) {
				event.Message = "Provider initialization timed out"
			} else {
				event.Message = fmt.Sprintf("Provider initialization failed: %v", err)
			}
		}
	}

	return event, err
}

var statesMap = map[EventType]func(ProviderEventDetails) State{
	ProviderReady:        func(_ ProviderEventDetails) State { return ReadyState },
	ProviderConfigChange: func(_ ProviderEventDetails) State { return ReadyState },
	ProviderStale:        func(_ ProviderEventDetails) State { return StaleState },
	ProviderError: func(e ProviderEventDetails) State {
		if e.ErrorCode == ProviderFatalCode {
			return FatalState
		}
		return ErrorState
	},
}

func stateFromEventOrError(event Event, err error) State {
	if err != nil {
		return stateFromError(err)
	}
	return stateFromEvent(event)
}

func stateFromEvent(event Event) State {
	if stateFn, ok := statesMap[event.EventType]; ok {
		return stateFn(event.ProviderEventDetails)
	}
	return NotReadyState // default
}

func stateFromError(err error) State {
	var e *ProviderInitError
	switch {
	case errors.As(err, &e):
		if e.ErrorCode == ProviderFatalCode {
			return FatalState
		}
	}
	return ErrorState // default
}
