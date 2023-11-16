package openfeature

import (
	"github.com/go-logr/logr"
	"github.com/open-feature/go-sdk/openfeature"
)

// SetProvider sets the default provider. Provider initialization is
// asynchronous and status can be checked from provider status
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.SetProvider,
// instead.
func SetProvider(provider FeatureProvider) error {
	return openfeature.SetProvider(provider)
}

// SetNamedProvider sets a provider mapped to the given Client name. Provider
// initialization is asynchronous and status can be checked from provider
// status
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.SetNamedProvider,
// instead.
func SetNamedProvider(clientName string, provider FeatureProvider) error {
	return openfeature.SetNamedProvider(clientName, provider)
}

// SetEvaluationContext sets the global evaluation context.
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.SetEvaluationContext, instead.
func SetEvaluationContext(evalCtx EvaluationContext) {
	openfeature.SetEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.SetLogger,
// instead.
func SetLogger(l logr.Logger) {
	openfeature.SetLogger(l)
}

// ProviderMetadata returns the default provider's metadata
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.ProviderMetadata,
// instead.
func ProviderMetadata() Metadata {
	return openfeature.ProviderMetadata()
}

// AddHooks appends to the collection of any previously added hooks
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.AddHooks,
// instead.
func AddHooks(hooks ...Hook) {
	openfeature.AddHooks(hooks...)
}

// AddHandler allows to add API level event handler
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.AddHandler,
// instead.
func AddHandler(eventType EventType, callback EventCallback) {
	openfeature.AddHandler(eventType, callback)
}

// RemoveHandler allows to remove API level event handler
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.RemoveHandler,
// instead.
func RemoveHandler(eventType EventType, callback EventCallback) {
	openfeature.RemoveHandler(eventType, callback)
}

// Shutdown active providers
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.Shutdown,
// instead.
func Shutdown() {
	openfeature.Shutdown()
}
