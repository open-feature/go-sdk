package openfeature

import (
	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/pkg/openfeature/internal"
	"sync"
)

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	prvder  FeatureProvider
	hks     []Hook
	evalCtx EvaluationContext
	logger  logr.Logger
	mu      sync.RWMutex
}

func (api *evaluationAPI) setProvider(provider FeatureProvider) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	api.prvder = provider
	api.logger.V(internal.Info).Info("set global provider", "name", provider.Metadata().Name)
}

func (api *evaluationAPI) provider() FeatureProvider {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.prvder
}

func (api *evaluationAPI) setEvaluationContext(evalCtx EvaluationContext) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	api.evalCtx = evalCtx
	api.logger.V(internal.Info).Info("set global evaluation context", "evaluationContext", evalCtx)
}

func (api *evaluationAPI) setLogger(l logr.Logger) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	api.logger = l
	api.logger.V(internal.Info).Info("set global logger")
}

func (api *evaluationAPI) getLogger() logr.Logger {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.logger
}

func (api *evaluationAPI) addHooks(hooks ...Hook) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	api.hks = append(api.hks, hooks...)
	api.logger.V(internal.Info).Info("appended hooks", "hooks", hooks)
}

func (api *evaluationAPI) getHooks() []Hook {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.hks
}

func (api *evaluationAPI) forTransaction() (FeatureProvider, []Hook, EvaluationContext) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	return api.prvder, api.hks, api.evalCtx
}
