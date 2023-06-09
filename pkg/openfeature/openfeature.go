package openfeature

import (
	"github.com/go-logr/logr"
)

// api is the global evaluationAPI.  This is a singleton and there can only be one instance.
var api evaluationAPI

// init initializes the OpenFeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	api = NewEvaluationAPI()
}

// SetProvider sets the global provider.
func SetProvider(provider FeatureProvider) {
	api.setProvider(provider)
}

// SetEvaluationContext sets the global evaluation context.
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.setEvaluationContext(evalCtx)
}

// SetLogger sets the global logger.
func SetLogger(l logr.Logger) {
	api.setLogger(l)
}

// ProviderMetadata returns the global provider's metadata
func ProviderMetadata() Metadata {
	return api.provider().Metadata()
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.Lock()
	defer api.Unlock()
	api.hks = append(api.hks, hooks...)
	api.logger.V(info).Info("appended hooks to the global singleton", "hooks", hooks)
}

func globalLogger() logr.Logger {
	api.RLock()
	defer api.RUnlock()
	return api.logger
}
