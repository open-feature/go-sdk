package openfeature

import (
	"sync"

	"github.com/go-logr/logr"
)

type evaluationAPI struct {
	provider          FeatureProvider
	hooks             []Hook
	evaluationContext EvaluationContext
	logger            logr.Logger
	sync.RWMutex
}

// api is the global evaluationAPI.  This is a singleton and there can only be one instance.
var api evaluationAPI

// init initializes the openfeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	api = evaluationAPI{
		provider:          NoopProvider{},
		hooks:             []Hook{},
		evaluationContext: EvaluationContext{},
		logger:            logr.New(logger{}),
	}
}

// EvaluationOptions should contain a list of hooks to be executed for a flag evaluation
type EvaluationOptions struct {
	hooks     []Hook
	hookHints HookHints
}

// NewEvaluationOptions constructs an EvaluationOptions
func NewEvaluationOptions(hooks []Hook, hookHints HookHints) EvaluationOptions {
	return EvaluationOptions{
		hooks:     hooks,
		hookHints: hookHints,
	}
}

// HookHints returns evaluation options' hook hints
func (e EvaluationOptions) HookHints() HookHints {
	return e.hookHints
}

// Hooks returns evaluation options' hooks
func (e EvaluationOptions) Hooks() []Hook {
	return e.hooks
}

func (api *evaluationAPI) setProvider(provider FeatureProvider) {
	api.Lock()
	defer api.Unlock()
	api.provider = provider
	api.logger.V(info).Info("set global provider", "name", provider.Metadata().Name)
}

func (api *evaluationAPI) setEvaluationContext(evalCtx EvaluationContext) {
	api.Lock()
	defer api.Unlock()
	api.evaluationContext = evalCtx
	api.logger.V(info).Info("set global evaluation context", "evaluationContext", evalCtx)
}

func (api *evaluationAPI) setLogger(l logr.Logger) {
	api.Lock()
	defer api.Unlock()
	api.logger = l
	api.logger.V(info).Info("set global logger")
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
	return api.provider.Metadata()
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.Lock()
	defer api.Unlock()
	api.hooks = append(api.hooks, hooks...)
	api.logger.V(info).Info("appended hooks to the global singleton", "hooks", hooks)
}

func globalLogger() logr.Logger {
	return api.logger
}
