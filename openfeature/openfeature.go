package openfeature

import (
	"github.com/go-logr/logr"
)

// api is the global evaluationAPI. This is a singleton and there can only be one instance.
// Avoid direct access.
var api evaluationAPI

// init initializes the OpenFeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	api = newEvaluationAPI()
}

// SetProvider sets the default provider. Provider initialization is asynchronous (non-blocking) and status can be checked from
// provider status
func SetProvider(provider FeatureProvider) error {
	return api.setProvider(provider)
}

// SetProviderAndWait sets the default provider. Provider initialization is synchronous (blocking) and status can be checked from
// provider status
func SetProviderAndWait(provider FeatureProvider) error {
	return api.setProviderAndWait(provider)
}

// SetNamedProvider sets a provider mapped to the given Client name. Provider initialization is asynchronous and
// status can be checked from provider status
func SetNamedProvider(clientName string, provider FeatureProvider) error {
	return api.setNamedProvider(clientName, provider)
}

// SetNamedProviderAndWait sets a provider mapped to the given Client name. Provider initialization is synchronous and
// status can be checked from provider status
func SetNamedProviderAndWait(clientName string, provider FeatureProvider) error {
	return api.setNamedProviderAndWait(clientName, provider)
}

// SetEvaluationContext sets the global evaluation context.
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.setEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
func SetLogger(l logr.Logger) {
	api.setLogger(l)
}

// ProviderMetadata returns the default provider's metadata
func ProviderMetadata() Metadata {
	return api.getProvider().Metadata()
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.addHooks(hooks...)
}

// AddHandler allows to add API level event handler
func AddHandler(eventType EventType, callback EventCallback) {
	api.eventExecutor.registerApiHandler(eventType, callback)
}

// addClientHandler is a helper for Client to add an event handler
func addClientHandler(name string, t EventType, c EventCallback) {
	api.eventExecutor.registerClientHandler(name, t, c)
}

// RemoveHandler allows to remove API level event handler
func RemoveHandler(eventType EventType, callback EventCallback) {
	api.eventExecutor.removeApiHandler(eventType, callback)
}

// removeClientHandler is a helper for Client to add an event handler
func removeClientHandler(name string, t EventType, c EventCallback) {
	api.eventExecutor.removeClientHandler(name, t, c)
}

// getAPIEventRegistry is a helper for testing
func getAPIEventRegistry() map[EventType][]EventCallback {
	return api.eventExecutor.apiRegistry
}

// getClientRegistry is a helper for testing
func getClientRegistry(client string) *scopedCallback {
	if v, ok := api.eventExecutor.scopedRegistry[client]; ok {
		return &v
	}

	return nil
}

// Shutdown active providers
func Shutdown() {
	api.shutdown()
}

// getProvider returns the default provider of the API. Intended to be used by tests
func getProvider() FeatureProvider {
	return api.getProvider()
}

// getNamedProviders returns the named provider map of the API. Intended to be used by tests
func getNamedProviders() map[string]FeatureProvider {
	return api.getNamedProviders()
}

// getHooks returns hooks of the API. Intended to be used by tests
func getHooks() []Hook {
	return api.getHooks()
}

// globalLogger return the global logger set at the API
func globalLogger() logr.Logger {
	return api.getLogger()
}

// forTransaction is a helper to retrieve transaction scoped operators by Client.
// Here, transaction means a flag evaluation.
func forTransaction(clientName string) (FeatureProvider, []Hook, EvaluationContext) {
	return api.forTransaction(clientName)
}
