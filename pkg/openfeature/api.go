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
}

// newEvaluationAPI is a helper to generate an API. Used internally
func newEvaluationAPI() evaluationAPI {
	return evaluationAPI{
		defaultProvider: NoopProvider{},
		namedProviders:  map[string]FeatureProvider{},
		hks:             []Hook{},
		evalCtx:         EvaluationContext{},
		logger:          logr.New(internal.Logger{}),
		mu:              sync.RWMutex{},
	}
}

// setProvider sets the default FeatureProvider of the evaluationAPI. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) setProvider(provider FeatureProvider) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}

	api.defaultProvider = provider
	api.logger.V(internal.Info).Info("set global provider", "name", provider.Metadata().Name)

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

	api.namedProviders[clientName] = provider
	api.logger.V(internal.Info).Info("set named provider provider", "name", "providerName", clientName, provider.Metadata().Name)

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
	api.logger.V(internal.Info).Info("set global evaluation context", "evaluationContext", evalCtx)
}

func (api *evaluationAPI) setLogger(l logr.Logger) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.logger = l
	api.logger.V(internal.Info).Info("set global logger")
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
	api.logger.V(internal.Info).Info("appended hooks", "hooks", hooks)
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
