package openfeature

import (
	"context"

	"github.com/go-logr/logr"
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

// GetApiInstance returns the current singleton IEvaluation instance.
//
// Deprecated: use [NewDefaultClient] or [NewClient] directly instead
//
//nolint:staticcheck // Renaming this now would be a breaking change.
func GetApiInstance() IEvaluation {
	return api
}

// NewDefaultClient returns a [Client] for the default domain. The default domain [Client] is the [IClient] instance that
// wraps around an unnamed [FeatureProvider]
func NewDefaultClient() *Client {
	return newClient("", api, eventing)
}

// SetProvider sets the default [FeatureProvider]. Provider initialization is asynchronous and status can be checked from
// provider status
func SetProvider(provider FeatureProvider) error {
	return api.SetProvider(provider)
}

// SetProviderAndWait sets the default [FeatureProvider] and waits for its initialization.
// Returns an error if initialization causes an error
func SetProviderAndWait(provider FeatureProvider) error {
	return api.SetProviderAndWait(provider)
}

// SetProviderWithContext sets the default [FeatureProvider] with context-aware initialization.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Use this function for non-blocking provider setup with timeout control where you want
// to continue application startup while the provider initializes in background.
// For providers that don't implement ContextAwareStateHandler, this behaves
// identically to SetProvider() but with timeout protection.
func SetProviderWithContext(ctx context.Context, provider FeatureProvider) error {
	return api.SetProviderWithContext(ctx, provider)
}

// SetProviderWithContextAndWait sets the default [FeatureProvider] with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this function for synchronous provider setup with guaranteed readiness when you need
// application startup to wait for the provider before continuing.
// Recommended timeout values: 1-5s for local providers, 10-30s for network-based providers.
func SetProviderWithContextAndWait(ctx context.Context, provider FeatureProvider) error {
	return api.SetProviderWithContextAndWait(ctx, provider)
}

// ProviderMetadata returns the default [FeatureProvider] metadata
func ProviderMetadata() Metadata {
	return api.GetProviderMetadata()
}

// SetNamedProvider sets a [FeatureProvider] mapped to the given [Client] domain. Provider initialization is asynchronous
// and status can be checked from provider status
func SetNamedProvider(domain string, provider FeatureProvider) error {
	return api.SetNamedProvider(domain, provider, true)
}

// SetNamedProviderAndWait sets a provider mapped to the given [Client] domain and waits for its initialization.
// Returns an error if initialization cause error
func SetNamedProviderAndWait(domain string, provider FeatureProvider) error {
	return api.SetNamedProvider(domain, provider, false)
}

// SetNamedProviderWithContext sets a [FeatureProvider] mapped to the given [Client] domain with context-aware initialization.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Named providers allow different domains to use different feature flag providers,
// enabling multi-tenant applications or microservice architectures.
func SetNamedProviderWithContext(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContext(ctx, domain, provider, true)
}

// SetNamedProviderWithContextAndWait sets a provider mapped to the given [Client] domain with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this for synchronous named provider setup where you need to ensure
// the provider is ready before proceeding.
func SetNamedProviderWithContextAndWait(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContextAndWait(ctx, domain, provider)
}

// NamedProviderMetadata returns the named provider's Metadata
func NamedProviderMetadata(name string) Metadata {
	return api.GetNamedProviderMetadata(name)
}

// SetEvaluationContext sets the global [EvaluationContext].
func SetEvaluationContext(evalCtx EvaluationContext) {
	api.SetEvaluationContext(evalCtx)
}

// SetLogger sets the global Logger.
//
// Deprecated: use [github.com/open-feature/go-sdk/openfeature/hooks.LoggingHook] instead.
func SetLogger(l logr.Logger) {
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

// Shutdown unconditionally calls shutdown on all registered providers,
// regardless of their state. It resets the state of the API, removing all
// hooks, event handlers, and providers.
func Shutdown() {
	api.Shutdown()
	initSingleton()
}

// ShutdownWithContext calls context-aware shutdown on all registered providers.
// If providers implement ContextAwareStateHandler, ShutdownWithContext will be called with the provided context.
// It resets the state of the API, removing all hooks, event handlers, and providers.
// This is intended to be called when your application is terminating.
// Returns an error if any provider shutdown fails or if context is cancelled during shutdown.
func ShutdownWithContext(ctx context.Context) error {
	err := api.ShutdownWithContext(ctx)
	initSingleton()
	return err
}
