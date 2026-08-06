package openfeature

import (
	"reflect"
)

// newProviderRef creates a new providerReference instance that wraps around a FeatureProvider implementation
func newProviderRef(provider FeatureProvider) providerReference {
	return providerReference{
		featureProvider:   provider,
		typeOf:            reflect.TypeOf(provider),
		shutdownSemaphore: make(chan any, 1), // Buffered to allow both send and close operations
	}
}

// providerReference is a helper struct to store FeatureProvider along with their
// shutdown semaphore
type providerReference struct {
	featureProvider   FeatureProvider
	typeOf            reflect.Type
	shutdownSemaphore chan any
}

// equals reports whether pr and other refer to the same provider.
//
// For providers whose dynamic type is comparable, equality is determined using
// Go's == operator. For non-comparable provider types (such as those containing
// maps or slices), equality falls back to reflect.DeepEqual.
//
// The providers must have the same dynamic type to be considered equal.
func (pr providerReference) equals(other providerReference) bool {
	if pr.typeOf == nil {
		return false
	}
	if pr.typeOf != reflect.TypeOf(other.featureProvider) {
		return false
	}
	if pr.typeOf.Comparable() {
		return pr.featureProvider == other.featureProvider
	}
	return reflect.DeepEqual(pr.featureProvider, other.featureProvider)
}
