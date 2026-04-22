package openfeature

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/go-logr/logr"
)

// providerBindingEntry holds a strong reference to the provider and the API instance it is bound to.
// The strong reference prevents the provider from being garbage-collected while bound, ensuring
// that the uintptr key in providerBindings cannot be reused by a new allocation.
type providerBindingEntry struct {
	provider FeatureProvider
	api      *evaluationAPI
}

// providerBindings is a global registry mapping provider pointer identity (uintptr) to the API
// instance it is currently bound to. This enforces spec requirement 1.8.4: a provider instance
// SHOULD NOT be bound to more than one API instance simultaneously.
//
// uintptr keys are used (instead of FeatureProvider interface keys) because FeatureProvider
// implementations may contain unhashable fields (e.g. maps or slices), which would panic if
// used as map keys.
//
// Lock ordering: always acquire evaluationAPI.mu before providerBindingsMu to avoid deadlocks.
var (
	providerBindings   = make(map[uintptr]*providerBindingEntry)
	providerBindingsMu sync.Mutex
)

// providerBindingKey returns a stable, hashable identity for provider suitable for use as a map key,
// and reports whether the provider should be tracked at all.
//
// Only non-nil pointer-type providers whose pointed-to type has non-zero size are tracked:
//   - Value-type providers are skipped — they lack stable pointer identity.
//   - Pointers to zero-size types (e.g. *NoopProvider) are skipped — Go's allocator reuses the
//     same "zerobase" address for all zero-size allocations, making such pointers indistinguishable.
func providerBindingKey(provider FeatureProvider) (uintptr, bool) {
	rv := reflect.ValueOf(provider)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return 0, false
	}
	if rv.Type().Elem().Size() == 0 {
		return 0, false
	}
	return rv.Pointer(), true
}

// bindProvider registers provider as bound to apiInst. Returns an error if the provider is already
// bound to a different API instance (spec 1.8.4). Must be called with apiInst.mu write-locked.
func bindProvider(provider FeatureProvider, apiInst *evaluationAPI) error {
	key, ok := providerBindingKey(provider)
	if !ok {
		return nil
	}

	providerBindingsMu.Lock()
	defer providerBindingsMu.Unlock()

	if entry, exists := providerBindings[key]; exists && entry.api != apiInst {
		return fmt.Errorf("provider %q is already bound to a different API instance", provider.Metadata().Name)
	}
	providerBindings[key] = &providerBindingEntry{provider: provider, api: apiInst}
	return nil
}

// unbindProvider removes provider from the global binding registry for apiInst.
// Safe to call with apiInst.mu held (lock order: evaluationAPI.mu → providerBindingsMu).
func unbindProvider(provider FeatureProvider, apiInst *evaluationAPI) {
	key, ok := providerBindingKey(provider)
	if !ok {
		return
	}

	providerBindingsMu.Lock()
	defer providerBindingsMu.Unlock()

	if entry, exists := providerBindings[key]; exists && entry.api == apiInst {
		delete(providerBindings, key)
	}
}

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
	_, err := api.setProviderWithContext(ctx, provider)
	return err
}

// SetProviderWithContextAndWait sets the default FeatureProvider with context-aware initialization and waits for completion.
func (api *evaluationAPI) SetProviderWithContextAndWait(ctx context.Context, provider FeatureProvider) error {
	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}
	initCh, err := api.setProviderWithContext(ctx, provider)
	if err != nil {
		return err
	}
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

	initCh, err := api.setNamedProviderWithContext(ctx, clientName, provider)
	if err != nil {
		return err
	}

	if async {
		return nil
	}

	return <-initCh
}

// SetNamedProviderWithContextAndWait sets a provider with client name using context-aware initialization and waits for completion.
func (api *evaluationAPI) SetNamedProviderWithContextAndWait(ctx context.Context, clientName string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContext(ctx, clientName, provider, false)
}

// maybeUnbindOldProvider removes oldProvider from the global binding registry if it is no longer
// referenced by this API instance. Must be called with api.mu held.
// All comparisons use pointer identity (via providerBindingKey) to avoid panics from unhashable
// FeatureProvider implementations that contain maps or slices.
func (api *evaluationAPI) maybeUnbindOldProvider(oldProvider FeatureProvider) {
	oldKey, tracked := providerBindingKey(oldProvider)
	if !tracked {
		return
	}
	// Is oldProvider still the default?
	if k, ok := providerBindingKey(api.defaultProvider); ok && k == oldKey {
		return
	}
	// Is oldProvider still registered as a named provider?
	for _, p := range api.namedProviders {
		if k, ok := providerBindingKey(p); ok && k == oldKey {
			return
		}
	}
	unbindProvider(oldProvider, api)
}

// setProviderWithContext sets the default FeatureProvider of the evaluationAPI with context-aware initialization.
// Returns an error immediately if the provider is already bound to a different API instance (spec 1.8.4).
func (api *evaluationAPI) setProviderWithContext(ctx context.Context, provider FeatureProvider) (<-chan error, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if err := bindProvider(provider, api); err != nil {
		return nil, err
	}

	oldProvider := api.defaultProvider
	api.defaultProvider = provider

	api.eventExecutor.registerDefaultProvider(provider)

	api.shutdownOld(ctx, oldProvider)

	// Unbind the old provider if it is no longer referenced by this API instance.
	if oldProvider != nil {
		api.maybeUnbindOldProvider(oldProvider)
	}

	return api.initNew(ctx, "", provider), nil
}

func (api *evaluationAPI) setNamedProviderWithContext(ctx context.Context, clientName string, provider FeatureProvider) (<-chan error, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if err := bindProvider(provider, api); err != nil {
		return nil, err
	}

	// Initialize new named provider and Shutdown the old one
	oldProvider := api.namedProviders[clientName]
	api.namedProviders[clientName] = provider

	api.eventExecutor.registerNamedEventingProvider(clientName, provider)

	api.shutdownOld(ctx, oldProvider)

	// Unbind the old provider if it is no longer referenced by this API instance.
	if oldProvider != nil {
		api.maybeUnbindOldProvider(oldProvider)
	}

	return api.initNew(ctx, clientName, provider), nil
}

func (api *evaluationAPI) initNew(ctx context.Context, clientName string, newProvider FeatureProvider) <-chan error {
	errCh := make(chan error, 1)

	// Initialize new provider async. The caller may wait on the channel.
	go func(executor *eventExecutor, evalCtx EvaluationContext, ctx context.Context, provider FeatureProvider, clientName string) {
		event, err := initializerWithContext(ctx, provider, evalCtx)
		executor.triggerEvent(event, provider)

		if err != nil {
			if clientName == "" {
				err = fmt.Errorf("failed to initialize default provider %q: %w", provider.Metadata().Name, err)
			} else {
				err = fmt.Errorf("failed to initialize named provider %q for domain %q: %w", provider.Metadata().Name, clientName, err)
			}
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

	// Release all provider bindings so providers can be re-registered elsewhere.
	api.unbindAllProvidersLocked()

	return errors.Join(errs...)
}

// unbindAllProviders releases all provider bindings for this API instance from the global registry.
// Must be called without api.mu held.
func (api *evaluationAPI) unbindAllProviders() {
	api.mu.RLock()
	defer api.mu.RUnlock()
	api.unbindAllProvidersLocked()
}

// unbindAllProvidersLocked releases all provider bindings. Must be called with api.mu held (any level).
func (api *evaluationAPI) unbindAllProvidersLocked() {
	seen := make(map[FeatureProvider]bool)
	if api.defaultProvider != nil {
		seen[api.defaultProvider] = true
	}
	for _, p := range api.namedProviders {
		seen[p] = true
	}
	for p := range seen {
		unbindProvider(p, api)
	}
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

func stateFromEvent(event Event) State {
	if stateFn, ok := statesMap[event.EventType]; ok {
		return stateFn(event.ProviderEventDetails)
	}
	return NotReadyState // default
}
