package openfeature

import (
	"context"
)

// api is the global evaluationImpl implementation. This is a singleton and there can only be one instance.
var (
	api      evaluationImpl
	eventing eventingImpl
)

// init initializes the OpenFeature evaluation API
func init() {
	initSingleton()
}

func initSingleton() {
	exec := newEventExecutor()
	eventing = exec

	api = newEvaluationAPI(exec)
}

// NewDefaultClient returns a [Client] for the default domain. The default domain [Client] is the [IClient] instance that
// wraps around an unnamed [FeatureProvider]
func NewDefaultClient() *Client {
	return newClient("", api, eventing)
}

// SetProvider sets the default [FeatureProvider] with context-aware initialization.
// If the provider implements StateHandler, Init will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Use this function for non-blocking provider setup with timeout control where you want
// to continue application startup while the provider initializes in background.
func SetProvider(ctx context.Context, provider FeatureProvider) error {
	return api.SetProvider(ctx, provider)
}

// SetProviderAndWait sets the default [FeatureProvider] with initialization and waits for completion.
// If the provider implements StateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this function for synchronous provider setup with guaranteed readiness when you need
// application startup to wait for the provider before continuing.
// Recommended timeout values: 1-5s for local providers, 10-30s for network-based providers.
func SetProviderAndWait(ctx context.Context, provider FeatureProvider) error {
	return api.SetProviderAndWait(ctx, provider)
}

// ProviderMetadata returns the default [FeatureProvider] metadata
func ProviderMetadata() Metadata {
	return api.GetProviderMetadata()
}

// SetNamedProvider sets a [FeatureProvider] mapped to the given [Client] domain with context-aware initialization.
// If the provider implements StateHandler, Init will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Named providers allow different domains to use different feature flag providers,
// enabling multi-tenant applications or microservice architectures.
func SetNamedProvider(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProvider(ctx, domain, provider)
}

// SetNamedProviderAndWait sets a provider mapped to the given [Client] domain with context-aware initialization and waits for completion.
// If the provider implements StateHandler, Init will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this for synchronous named provider setup where you need to ensure
// the provider is ready before proceeding.
func SetNamedProviderAndWait(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProviderAndWait(ctx, domain, provider)
}

// NamedProviderMetadata returns the named provider's Metadata
func NamedProviderMetadata(name string) Metadata {
	return api.GetNamedProviderMetadata(name)
}

// SetEvaluationContext sets the global [EvaluationContext].
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.SetEvaluationContext(evalCtx)
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api.AddHooks(hooks...)
}

// AddHandler allows to add API level event handlers
func AddHandler(eventType EventType, callback EventCallback) {
	api.AddHandler(eventType, callback)
}

// RemoveHandler allows for removal of API level event handlers
func RemoveHandler(eventType EventType, callback EventCallback) {
	api.RemoveHandler(eventType, callback)
}

// Shutdown calls shutdown on all registered providers.
// It resets the state of the API, removing all hooks, event handlers, and providers.
// This is intended to be called when your application is terminating.
// Returns an error if any provider shutdown fails or if context is cancelled during shutdown.
func Shutdown(ctx context.Context) error {
	err := api.Shutdown(ctx)
	initSingleton()
	return err
}
