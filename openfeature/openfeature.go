package openfeature

import (
	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/openfeature/internal"
)

// api is the global evaluationImpl implementation. This is a singleton and there can only be one instance.
var api evaluationImpl
var eventing eventingImpl
var logger logr.Logger

// init initializes the OpenFeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	logger = logr.New(internal.Logger{})

	var exec = newEventExecutor(logger)
	eventing = exec

	api = newEvaluationAPI(exec, logger)
}

// GetApiInstance returns the current singleton IEvaluation instance.
// This is the preferred interface to interact with OpenFeature functionalities
func GetApiInstance() IEvaluation {
	return api
}

// SetProvider sets the default provider. Provider initialization is asynchronous and status can be checked from
// provider status
func SetProvider(provider FeatureProvider) error {
	return api.SetProvider(provider)
}

// SetProviderAndWait sets the default provider and waits for its initialization.
// Returns an error if initialization cause error
func SetProviderAndWait(provider FeatureProvider) error {
	return api.SetProviderAndWait(provider)
}

// ProviderMetadata returns the default provider's metadata
func ProviderMetadata() Metadata {
	return api.GetProviderMetadata()
}

// SetNamedProvider sets a provider mapped to the given Client domain. Provider initialization is asynchronous and
// status can be checked from provider status
func SetNamedProvider(clientDomain string, provider FeatureProvider) error {
	return api.SetNamedProvider(clientDomain, provider, true)
}

// SetNamedProviderAndWait sets a provider mapped to the given Client domain and waits for its initialization.
// Returns an error if initialization cause error
func SetNamedProviderAndWait(clientDomain string, provider FeatureProvider) error {
	return api.SetNamedProvider(clientDomain, provider, false)
}

// NamedProviderMetadata returns the named provider's Metadata
func NamedProviderMetadata(name string) Metadata {
	return api.GetNamedProviderMetadata(name)
}

// SetEvaluationContext sets the global evaluation context.
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.SetEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
func SetLogger(l logr.Logger) {
	api.SetLogger(l)
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.AddHooks(hooks...)
}

// AddHandler allows to add API level event handler
func AddHandler(eventType EventType, callback EventCallback) {
	api.AddHandler(eventType, callback)
}

// RemoveHandler allows to remove API level event handler
func RemoveHandler(eventType EventType, callback EventCallback) {
	api.RemoveHandler(eventType, callback)
}

// Shutdown active providers
func Shutdown() {
	api.Shutdown()
}
