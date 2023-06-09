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

// getProvider returns the default provider of the API. Intended to be used by tests
func getProvider() FeatureProvider {
	return api.defaultProvider
}

// SetNamedProvider sets a provider mapped to the given Client name.
func SetNamedProvider(clientName string, provider FeatureProvider) error {
	return api.setNamedProvider(clientName, provider)
}

// getProvider returns the default provider of the API. Intended to be used by tests
func getNamedProviders() map[string]FeatureProvider {
	return api.namedProviders
}

// SetEvaluationContext sets the global evaluation context.
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.setEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
func SetLogger(l logr.Logger) {
	api.setLogger(l)
}

// ProviderMetadata returns the global provider's metadata
func ProviderMetadata() Metadata {
	return api.provider().Metadata()
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.addHooks(hooks...)
}

// getHooks returns hooks of the API. Intended to be used by tests
func getHooks() []Hook {
	return api.getHooks()
}

func globalLogger() logr.Logger {
	return api.getLogger()
}

func forTransaction(clientName string) (FeatureProvider, []Hook, EvaluationContext) {
	return api.forTransaction(clientName)
}
