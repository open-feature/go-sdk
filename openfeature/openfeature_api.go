package openfeature

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/go-logr/logr"
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

func (api *evaluationAPI) SetProvider(provider FeatureProvider) error {
	return api.SetProviderWithContext(context.Background(), provider)
}

func (api *evaluationAPI) SetProviderAndWait(provider FeatureProvider) error {
	return api.SetProviderWithContextAndWait(context.Background(), provider)
}

// SetProviderWithContext sets the default FeatureProvider with context-aware initialization.
func (api *evaluationAPI) SetProviderWithContext(ctx context.Context, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}
	api.setProviderWithContext(ctx, provider)
	return nil
}

// SetProviderWithContextAndWait sets the default FeatureProvider with context-aware initialization and waits for completion.
func (api *evaluationAPI) SetProviderWithContextAndWait(ctx context.Context, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}
	initCh := api.setProviderWithContext(ctx, provider)
	return <-initCh
}

// SetNamedProvider sets a provider with client name. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) SetNamedProvider(clientName string, provider FeatureProvider, async bool) error {
	return api.SetNamedProviderWithContext(context.Background(), clientName, provider, async)
}

// SetNamedProviderWithContext sets a provider with client name using context-aware initialization.
func (api *evaluationAPI) SetNamedProviderWithContext(ctx context.Context, clientName string, provider FeatureProvider, async bool) error {
	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}

	initCh := api.setNamedProviderWithContext(ctx, clientName, provider)

	if async {
		return nil
	}

	return <-initCh
}

// SetNamedProviderWithContextAndWait sets a provider with client name using context-aware initialization and waits for completion.
func (api *evaluationAPI) SetNamedProviderWithContextAndWait(ctx context.Context, clientName string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContext(ctx, clientName, provider, false)
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

	// Initialize new named provider and Shutdown the old one
	oldProvider := api.namedProviders[clientName]
	api.namedProviders[clientName] = provider

	api.eventExecutor.registerNamedEventingProvider(clientName, provider)

	api.shutdownOld(ctx, oldProvider)

	return api.initNew(ctx, clientName, provider)
}

func (api *evaluationAPI) initNew(ctx context.Context, clientName string, newProvider FeatureProvider) <-chan error {
	errCh := make(chan error, 1)

	// Initialize new provider async. The caller may wait on the channel.
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
		// Check if the provider supports context-aware shutdown
		if contextHandler, ok := forShutdown.(ContextAwareStateHandler); ok {
			// Use the provided context directly - user controls timeout
			_ = contextHandler.ShutdownWithContext(parentCtx)
		} else {
			// Fall back to regular shutdown for backward compatibility
			forShutdown.Shutdown()
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
func (api *evaluationAPI) GetClient() IClient {
	return newClient("", api, api.eventExecutor)
}

// GetNamedClient returns a IClient bound to the given named provider
func (api *evaluationAPI) GetNamedClient(clientName string) IClient {
	return newClient(clientName, api, api.eventExecutor)
}

func (api *evaluationAPI) SetEvaluationContext(evalCtx EvaluationContext) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.evalCtx = evalCtx
}

// Deprecated: use [github.com/open-feature/go-sdk/openfeature/hooks.LoggingHook] instead.
func (api *evaluationAPI) SetLogger(l logr.Logger) {
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

func (api *evaluationAPI) Shutdown() {
	// Use the context-aware shutdown with background context and ignore errors
	// to maintain backward compatibility (Shutdown doesn't return an error)
	_ = api.ShutdownWithContext(context.Background())
}

// ShutdownWithContext calls context-aware shutdown on all registered providers.
// If providers implement ContextAwareStateHandler, ShutdownWithContext will be called with the provided context.
// Returns an error if any provider shutdown fails or if context is cancelled during shutdown.
func (api *evaluationAPI) ShutdownWithContext(ctx context.Context) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	var errs []error

	// Shutdown default provider
	if api.defaultProvider != nil {
		if contextHandler, ok := api.defaultProvider.(ContextAwareStateHandler); ok {
			if err := contextHandler.ShutdownWithContext(ctx); err != nil {
				errs = append(errs, fmt.Errorf("default provider shutdown failed: %w", err))
			}
		} else if stateHandler, ok := api.defaultProvider.(StateHandler); ok {
			stateHandler.Shutdown()
		}
	}

	// Shutdown all named providers
	for name, provider := range api.namedProviders {
		if contextHandler, ok := provider.(ContextAwareStateHandler); ok {
			if err := contextHandler.ShutdownWithContext(ctx); err != nil {
				errs = append(errs, fmt.Errorf("named provider %q shutdown failed: %w", name, err))
			}
		} else if stateHandler, ok := provider.(StateHandler); ok {
			stateHandler.Shutdown()
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

	// Check for context-aware handler first
	if contextHandler, ok := provider.(ContextAwareStateHandler); ok {
		err := contextHandler.InitWithContext(ctx, evalCtx)
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
		return event, err
	}

	// Fall back to regular StateHandler for backward compatibility
	handler, ok := provider.(StateHandler)
	if !ok {
		// Note - a provider without state handling capability can be assumed to be ready immediately.
		return event, nil
	}

	err := handler.Init(evalCtx)
	if err != nil {
		event.EventType = ProviderError
		event.Message = fmt.Sprintf("Provider initialization failed: %v", err)
		var initErr *ProviderInitError
		if errors.As(err, &initErr) {
			event.EventType = ProviderError
			event.ErrorCode = initErr.ErrorCode
			event.Message = initErr.Message
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
