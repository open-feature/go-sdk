package openfeature

import (
	"sync"
)

type evaluationAPI struct {
	provider FeatureProvider
	sync.RWMutex
}

// api is the global evaluationAPI.  This is a singleton and there can only be one instance.
var api evaluationAPI

// init initializes the openfeature evaluation API
func init() {
	api = evaluationAPI{
		provider: NoopProvider{},
	}
}

// EvaluationOption should contain a list of hooks to be executed for a flag evaluation
type EvaluationOption interface {}

func (api *evaluationAPI) setProvider(provider FeatureProvider) {
	api.Lock()
	defer api.Unlock()
	api.provider = provider
}

// SetProvider sets the global provider.
func SetProvider(provider FeatureProvider) {
	api.setProvider(provider)
}

// ProviderMetadata returns the global provider's metadata
func ProviderMetadata() Metadata {
	return api.provider.Metadata()
}
