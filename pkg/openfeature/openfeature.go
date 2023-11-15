package openfeature

import (
	"github.com/go-logr/logr"
	openfeature "github.com/open-feature/go-sdk"
)

// SetProvider sets the default provider. Provider initialization is
// asynchronous and status can be checked from provider status
//
// Deprecated: use github.com/open-feature/go-sdk.SetProvider,
// instead.
func SetProvider(provider FeatureProvider) error {
	return openfeature.SetProvider(provider)
}

// SetNamedProvider sets a provider mapped to the given Client name. Provider
// initialization is asynchronous and status can be checked from provider
// status
//
// Deprecated: use github.com/open-feature/go-sdk.SetNamedProvider,
// instead.
func SetNamedProvider(clientName string, provider FeatureProvider) error {
	return openfeature.SetNamedProvider(clientName, provider)
}

// SetEvaluationContext sets the global evaluation context.
//
// Deprecated: use
// github.com/open-feature/go-sdk.SetEvaluationContext, instead.
func SetEvaluationContext(evalCtx EvaluationContext) {
	openfeature.SetEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
//
// Deprecated: use github.com/open-feature/go-sdk.SetLogger,
// instead.
func SetLogger(l logr.Logger) {
	openfeature.SetLogger(l)
}

// ProviderMetadata returns the default provider's metadata
//
// Deprecated: use github.com/open-feature/go-sdk.ProviderMetadata,
// instead.
func ProviderMetadata() Metadata {
	return openfeature.ProviderMetadata()
}

// AddHooks appends to the collection of any previously added hooks
//
// Deprecated: use github.com/open-feature/go-sdk.AddHooks,
// instead.
func AddHooks(hooks ...Hook) {
	openfeature.AddHooks(hooks...)
}

// AddHandler allows to add API level event handler
//
// Deprecated: use github.com/open-feature/go-sdk.AddHandler,
// instead.
func AddHandler(eventType EventType, callback EventCallback) {
	openfeature.AddHandler(eventType, callback)
}

// RemoveHandler allows to remove API level event handler
//
// Deprecated: use github.com/open-feature/go-sdk.RemoveHandler,
// instead.
func RemoveHandler(eventType EventType, callback EventCallback) {
	openfeature.RemoveHandler(eventType, callback)
}

// Shutdown active providers
//
// Deprecated: use github.com/open-feature/go-sdk.Shutdown,
// instead.
func Shutdown() {
	openfeature.Shutdown()
}
