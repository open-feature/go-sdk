package openfeature

import (
	"errors"
	"fmt"
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
	apiCtx          EvaluationContext
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
		apiCtx:          EvaluationContext{},
		logger:          logger,
		mu:              sync.RWMutex{},
		eventExecutor:   newEventExecutor(logger),
	}
}

// setProvider sets the default FeatureProvider of the evaluationAPI.
// Returns an error if provider registration cause an error
func (api *evaluationAPI) setProvider(provider FeatureProvider, async bool) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("default provider cannot be set to nil")
	}

	oldProvider := api.defaultProvider
	api.defaultProvider = provider

	err := api.initNewAndShutdownOld(provider, oldProvider, async)
	if err != nil {
		return err
	}

	err = api.eventExecutor.registerDefaultProvider(provider)
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

// setProvider sets a provider with client domain. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) setNamedProvider(clientName string, provider FeatureProvider, async bool) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}

	// Initialize new named provider and shutdown the old one
	// Provider update must be non-blocking, hence initialization & shutdown happens concurrently
	oldProvider := api.namedProviders[clientName]
	api.namedProviders[clientName] = provider

	err := api.initNewAndShutdownOld(provider, oldProvider, async)
	if err != nil {
		return err
	}

	err = api.eventExecutor.registerNamedEventingProvider(clientName, provider)
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

func (api *evaluationAPI) setEvaluationContext(apiCtx EvaluationContext) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.apiCtx = apiCtx
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
// Returns the default FeatureProvider if no provider mapping exist for the given client domain.
func (api *evaluationAPI) forTransaction(clientDomain string) (FeatureProvider, []Hook, EvaluationContext) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	var provider FeatureProvider

	provider = api.namedProviders[clientDomain]
	if provider == nil {
		provider = api.defaultProvider
	}

	return provider, api.hks, api.apiCtx
}

// initNewAndShutdownOld is a helper to initialise new FeatureProvider and shutdown the old FeatureProvider.
func (api *evaluationAPI) initNewAndShutdownOld(newProvider FeatureProvider, oldProvider FeatureProvider, async bool) error {
	if async {
		go func(executor *eventExecutor, ctx EvaluationContext) {
			// for async initialization, error is conveyed as an event
			event, _ := initializer(newProvider, ctx)
			executor.triggerEvent(event, newProvider)
		}(api.eventExecutor, api.apiCtx)
	} else {
		event, err := initializer(newProvider, api.apiCtx)
		api.eventExecutor.triggerEvent(event, newProvider)
		if err != nil {
			return err
		}
	}

	v, ok := oldProvider.(StateHandler)

	// oldProvider can be nil or without state handling capability
	if oldProvider == nil || !ok {
		return nil
	}

	// check for multiple bindings
	if oldProvider == api.defaultProvider || contains(oldProvider, maps.Values(api.namedProviders)) {
		return nil
	}

	go func(forShutdown StateHandler) {
		forShutdown.Shutdown()
	}(v)

	return nil
}

// initializer is a helper to execute provider initialization and generate appropriate event for the initialization
// It also returns an error if the initialization resulted in an error
func initializer(provider FeatureProvider, apiCtx EvaluationContext) (Event, error) {
	var event = Event{
		ProviderName: provider.Metadata().Name,
		EventType:    ProviderReady,
		ProviderEventDetails: ProviderEventDetails{
			Message: "Provider initialization successful",
		},
	}

	handler, ok := provider.(StateHandler)
	if !ok {
		// Note - a provider without state handling capability can be assumed to be ready immediately.
		return event, nil
	}

	err := handler.Init(apiCtx)
	if err != nil {
		event.EventType = ProviderError
		event.Message = fmt.Sprintf("Provider initialization error, %v", err)
	}

	return event, err
}

func contains(provider FeatureProvider, in []FeatureProvider) bool {
	for _, p := range in {
		if provider == p {
			return true
		}
	}

	return false
}
