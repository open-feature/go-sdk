package openfeature

import (
	"errors"
	"sync"

	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/openfeature/internal"
	"golang.org/x/exp/maps"
)

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	defaultProvider FeatureProvider
	namedProviders  map[string]FeatureProvider
	hks             []Hook
	evalCtx         EvaluationContext
	logger          logr.Logger
	mu              sync.RWMutex
	eventExecutor   *eventExecutor
}

// newEvaluationAPI is a helper to generate an API. Used internally
func newEvaluationAPI() evaluationAPI {
	logger := logr.New(internal.Logger{})

	return evaluationAPI{
		defaultProvider: NoopProvider{},
		namedProviders:  map[string]FeatureProvider{},
		hks:             []Hook{},
		evalCtx:         EvaluationContext{},
		logger:          logger,
		mu:              sync.RWMutex{},
		eventExecutor:   newEventExecutor(logger),
	}
}

// setProvider sets the default FeatureProvider of the evaluationAPI. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) setProvider(provider FeatureProvider) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}

	// Initialize new default provider and shutdown the old one
	// Provider update must be non-blocking, hence initialization & shutdown happens concurrently
	oldProvider := api.defaultProvider
	api.defaultProvider = provider

	api.initNewAndShutdownOldAsync(provider, oldProvider)
	err := api.eventExecutor.registerDefaultProvider(provider)
	if err != nil {
		return err
	}

	return nil
}

// setProviderAndWait sets the default FeatureProvider of the evaluationAPI. Returns an error if FeatureProvider is nil
// This is a blocking call and will wait for the provider to be ready
func (api *evaluationAPI) setProviderAndWait(provider FeatureProvider) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}

	// Initialize new default provider and shutdown the old one
	// Provider update must be non-blocking, hence initialization & shutdown happens concurrently
	oldProvider := api.defaultProvider
	api.defaultProvider = provider

	api.initNewAndShutdownOldSync(provider, oldProvider)
	err := api.eventExecutor.registerDefaultProvider(provider)
	if err != nil {
		return err
	}

	return nil
}

// getProvider returns the default FeatureProvider
func (api *evaluationAPI) getProvider() FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.defaultProvider
}

// setProvider sets a provider with client name. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) setNamedProvider(clientName string, provider FeatureProvider) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}

	// Initialize new named provider and shutdown the old one
	// Provider update must be non-blocking, hence initialization & shutdown happens concurrently
	oldProvider := api.namedProviders[clientName]
	api.namedProviders[clientName] = provider

	api.initNewAndShutdownOldAsync(provider, oldProvider)
	err := api.eventExecutor.registerNamedEventingProvider(clientName, provider)
	if err != nil {
		return err
	}

	return nil
}

// getNamedProviders returns named providers map.
func (api *evaluationAPI) getNamedProviders() map[string]FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.namedProviders
}

func (api *evaluationAPI) setEvaluationContext(evalCtx EvaluationContext) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.evalCtx = evalCtx
}

func (api *evaluationAPI) setLogger(l logr.Logger) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.logger = l
	api.eventExecutor.updateLogger(l)
}

func (api *evaluationAPI) getLogger() logr.Logger {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.logger
}

func (api *evaluationAPI) addHooks(hooks ...Hook) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.hks = append(api.hks, hooks...)
}

func (api *evaluationAPI) shutdown() {
	api.mu.Lock()
	defer api.mu.Unlock()

	v, ok := api.defaultProvider.(StateHandler)
	if ok {
		v.Shutdown()
	}

	for _, provider := range api.namedProviders {
		v, ok = provider.(StateHandler)
		if ok {
			v.Shutdown()
		}
	}
}

func (api *evaluationAPI) getHooks() []Hook {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.hks
}

// forTransaction is a helper to retrieve transaction(flag evaluation) scoped operators.
// Returns the default FeatureProvider if no provider mapping exist for the given client name.
func (api *evaluationAPI) forTransaction(clientName string) (FeatureProvider, []Hook, EvaluationContext) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	var provider FeatureProvider

	provider = api.namedProviders[clientName]
	if provider == nil {
		provider = api.defaultProvider
	}

	return provider, api.hks, api.evalCtx
}

// initNewAndShutdownOldAsync is a helper to initialise new FeatureProvider and shutdown the old FeatureProvider.
// Operations happen concurrently.
func (api *evaluationAPI) initNewAndShutdownOldAsync(newProvider FeatureProvider, oldProvider FeatureProvider) {
	v, ok := newProvider.(StateHandler)
	if ok && v.Status() == NotReadyState {
		go func(provider FeatureProvider, stateHandler StateHandler, evalCtx EvaluationContext, eventChan chan eventPayload) {
			err := stateHandler.Init(evalCtx)
			// emit ready/error event once initialization is complete
			if err != nil {
				eventChan <- eventPayload{
					Event{
						ProviderName:         provider.Metadata().Name,
						EventType:            ProviderError,
						ProviderEventDetails: ProviderEventDetails{},
					}, provider,
				}
			} else {
				eventChan <- eventPayload{
					Event{
						ProviderName:         provider.Metadata().Name,
						EventType:            ProviderReady,
						ProviderEventDetails: ProviderEventDetails{},
					}, provider,
				}
			}
		}(newProvider, v, api.evalCtx, api.eventExecutor.eventChan)
	}

	v, ok = oldProvider.(StateHandler)

	// oldProvider can be nil or without state handling capability
	if oldProvider == nil || !ok {
		return
	}

	// check for multiple bindings
	if oldProvider == api.defaultProvider || contains(oldProvider, maps.Values(api.namedProviders)) {
		return
	}

	go func(forShutdown StateHandler) {
		forShutdown.Shutdown()
	}(v)
}

// initNewAndShutdownOldSync is a helper to initialise new FeatureProvider and shutdown the old FeatureProvider.
// Operations happen synchronously.
func (api *evaluationAPI) initNewAndShutdownOldSync(newProvider FeatureProvider, oldProvider FeatureProvider) {
	v, ok := newProvider.(StateHandler)
	if ok && v.Status() == NotReadyState {
		err := v.Init(api.evalCtx)
		if err != nil {
			api.eventExecutor.eventChan <- eventPayload{
				Event{
					ProviderName:         newProvider.Metadata().Name,
					EventType:            ProviderError,
					ProviderEventDetails: ProviderEventDetails{},
				}, newProvider,
			}
		} else {
			api.eventExecutor.eventChan <- eventPayload{
				Event{
					ProviderName:         newProvider.Metadata().Name,
					EventType:            ProviderReady,
					ProviderEventDetails: ProviderEventDetails{},
				}, newProvider,
			}
		}
	}

	v, ok = oldProvider.(StateHandler)

	// oldProvider can be nil or without state handling capability
	if oldProvider == nil || !ok {
		return
	}

	// check for multiple bindings
	if oldProvider == api.defaultProvider || contains(oldProvider, maps.Values(api.namedProviders)) {
		return
	}

	v.Shutdown()
}

func contains(provider FeatureProvider, in []FeatureProvider) bool {
	for _, p := range in {
		if provider == p {
			return true
		}
	}

	return false
}
