package openfeature

import (
	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
	"sync"
)

// api is the global evaluationAPI. This is a singleton and there can only be one instance. Avoid direct access.
var api evaluationAPI

// init initializes the OpenFeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	api = evaluationAPI{
		prvder:  NoopProvider{},
		hks:     []Hook{},
		evalCtx: EvaluationContext{},
		logger:  logr.New(internal.Logger{}),
		mu:      sync.RWMutex{},
	}
}

// SetProvider sets the global provider.
func SetProvider(provider FeatureProvider) {
	api.setProvider(provider)
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

func globalLogger() logr.Logger {
	return api.getLogger()
}

func forTransaction() (FeatureProvider, []Hook, EvaluationContext) {
	return api.forTransaction()
}
