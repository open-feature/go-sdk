package openfeature

import (
	"errors"
	"sync"

	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
)

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	defaultProvider FeatureProvider
	namedProviders  map[string]FeatureProvider
	hks             []Hook
	evalCtx         EvaluationContext
	logger          logr.Logger
	mu              sync.RWMutex
	eventExecutor
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
	api.initAndShutdown(provider, oldProvider)
	api.defaultProvider = provider

	err := api.registerDefaultProvider(provider)
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

	// Initialize new default provider and shutdown the old one
	// Provider update must be non-blocking, hence initialization & shutdown happens concurrently
	api.initAndShutdown(provider, api.namedProviders[clientName])
	api.namedProviders[clientName] = provider

	err := api.registerNamedEventingProvider(clientName, provider)
	if err != nil {
		return err
	}

	return nil
}

// getNamedProviders return default providers
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

// initAndShutdown is a helper to initialise new FeatureProvider and shutdown old FeatureProvider.
// Operation happens concurrently.
func (api *evaluationAPI) initAndShutdown(newProvider FeatureProvider, oldProvider FeatureProvider) {
	go func() {
		v, ok := newProvider.(StateHandler)
		if ok {
			v.Init(api.evalCtx)
		}

		// oldProvider can be nil
		if oldProvider != nil {
			v, ok = oldProvider.(StateHandler)
			if ok {
				v.Shutdown()
			}
		}
	}()
}
