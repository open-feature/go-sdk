package openfeature

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
)

// evaluationImpl is an internal reference interface extending IEvaluation
type evaluationImpl interface {
	IEvaluation
	GetProvider() FeatureProvider
	GetNamedProviders() map[string]FeatureProvider
	GetHooks() []Hook
	SetLogger(l logr.Logger)
	ForEvaluation(clientName string) (FeatureProvider, []Hook, EvaluationContext)
}

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	defaultProvider FeatureProvider
	namedProviders  map[string]FeatureProvider
	hks             []Hook
	apiCtx          EvaluationContext
	eventExecutor   *eventExecutor
	logger          logr.Logger
	mu              sync.RWMutex
}

// newEvaluationAPI is a helper to generate an API. Used internally
func newEvaluationAPI(eventExecutor *eventExecutor, log logr.Logger) *evaluationAPI {
	return &evaluationAPI{
		defaultProvider: NoopProvider{},
		namedProviders:  map[string]FeatureProvider{},
		hks:             []Hook{},
		apiCtx:          EvaluationContext{},
		logger:          log,
		mu:              sync.RWMutex{},
		eventExecutor:   eventExecutor,
	}
}

func (api *evaluationAPI) SetProvider(provider FeatureProvider) error {
	return api.setProvider(provider, true)
}

func (api *evaluationAPI) SetProviderAndWait(provider FeatureProvider) error {
	return api.setProvider(provider, false)
}

// GetProviderMetadata returns the default FeatureProvider's metadata
func (api *evaluationAPI) GetProviderMetadata() Metadata {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.defaultProvider.Metadata()
}

// SetNamedProvider sets a provider with client name. Returns an error if FeatureProvider is nil
func (api *evaluationAPI) SetNamedProvider(clientName string, provider FeatureProvider, async bool) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if provider == nil {
		return errors.New("provider cannot be set to nil")
	}

	// Initialize new named provider and Shutdown the old one
	// Provider update must be non-blocking, hence initialization & Shutdown happens concurrently
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
	return newClient("", api, api.eventExecutor, api.logger)
}

// GetNamedClient returns a IClient bound to the given named provider
func (api *evaluationAPI) GetNamedClient(clientName string) IClient {
	return newClient(clientName, api, api.eventExecutor, api.logger)
}

func (api *evaluationAPI) SetEvaluationContext(apiCtx EvaluationContext) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.apiCtx = apiCtx
}

func (api *evaluationAPI) SetLogger(l logr.Logger) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.logger = l
	api.eventExecutor.updateLogger(l)
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

	return provider, api.hks, api.apiCtx
}

// GetProvider returns the default FeatureProvider
func (api *evaluationAPI) GetProvider() FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.defaultProvider
}

// SetProvider sets the default FeatureProvider of the evaluationAPI.
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

// initNewAndShutdownOld is a helper to initialise new FeatureProvider and Shutdown the old FeatureProvider.
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
