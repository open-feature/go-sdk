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

// SetProvider sets the default provider.
func SetProvider(provider FeatureProvider) error {
	return api.setProvider(provider)
}

// SetNamedProvider sets a provider mapped to the given Client name.
func SetNamedProvider(clientName string, provider FeatureProvider) error {
	return api.setNamedProvider(clientName, provider)
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

// getProvider returns the default provider of the API. Intended to be used by tests
func getProvider() FeatureProvider {
	return api.getProvider()
}

// getNamedProviders returns the default provider of the API. Intended to be used by tests
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
