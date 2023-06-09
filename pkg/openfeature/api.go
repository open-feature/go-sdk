package openfeature

import (
	"github.com/go-logr/logr"
	"sync"
)

// evaluationAPI wraps OpenFeature evaluation API functionalities
type evaluationAPI struct {
	prvder  FeatureProvider
	hks     []Hook
	evalCtx EvaluationContext
	logger  logr.Logger
	mutex
}

func NewEvaluationAPI() evaluationAPI {
	return evaluationAPI{
		prvder:  NoopProvider{},
		hks:     []Hook{},
		evalCtx: EvaluationContext{},
		logger:  logr.New(logger{}),
		mutex:   &sync.RWMutex{},
	}
}

func (api *evaluationAPI) provider() FeatureProvider {
	api.RLock()
	defer api.RUnlock()
	return api.prvder
}

func (api *evaluationAPI) setProvider(provider FeatureProvider) {
	api.Lock()
	defer api.Unlock()
	api.prvder = provider
	api.logger.V(info).Info("set global provider", "name", provider.Metadata().Name)
}

func (api *evaluationAPI) setEvaluationContext(evalCtx EvaluationContext) {
	api.Lock()
	defer api.Unlock()
	api.evalCtx = evalCtx
	api.logger.V(info).Info("set global evaluation context", "evaluationContext", evalCtx)
}

func (api *evaluationAPI) setLogger(l logr.Logger) {
	api.Lock()
	defer api.Unlock()
	api.logger = l
	api.logger.V(info).Info("set global logger")
}

func (api *evaluationAPI) hooks() []Hook {
	api.RLock()
	defer api.RUnlock()
	return api.hks
}
