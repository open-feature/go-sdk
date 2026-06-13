package openfeature

import (
	"context"
	"sync/atomic"

	"github.com/open-feature/go-sdk/openfeature/internal/factory"
)

// api is the global evaluationImpl implementation. This is a singleton and there can only be one instance.
var (
	apiInstance atomic.Pointer[EvaluationAPI]
)

// init initializes the OpenFeature evaluation API
func init() {
	// register the isolated-instance constructor; see openfeature/internal/factory.
	factory.NewAPI = func() any {
		return newAPI()
	}
	factory.CurrentAPI = func() any {
		return api()
	}
	resetSingleton()
}

// resetSingleton stops (if running) the event executor and starts a new one.
func resetSingleton() {
	_ = resetSingletonWithContext(context.Background())
}

// resetSingletonWithContext stops (if running) the event executor and starts a new one.
func resetSingletonWithContext(ctx context.Context) error {
	nextAPI := newAPI()
	oldAPI := apiInstance.Swap(nextAPI)
	if oldAPI != nil {
		return oldAPI.Shutdown(ctx)
	}
	return nil
}

// api returns the current singleton [EvaluationAPI].
func api() *EvaluationAPI {
	return apiInstance.Load()
}

// newAPI creates a fresh *EvaluationAPI.
func newAPI() *EvaluationAPI {
	exec := newEventExecutor()
	return newEvaluationAPI(exec)
}

// NewDefaultClient returns a [Client] for the default domain. The default domain [Client] is the [IClient] instance that
// wraps around an unnamed [FeatureProvider]
func NewDefaultClient() *Client {
	return api().NewClient()
}

// SetProvider sets the default [FeatureProvider]. Provider initialization is asynchronous and status can be checked from
// provider status
func SetProvider(provider FeatureProvider) error {
	return api().SetProvider(context.Background(), provider)
}

// SetProviderAndWait sets the default [FeatureProvider] and waits for its initialization.
// Returns an error if initialization causes an error
func SetProviderAndWait(provider FeatureProvider) error {
	return api().SetProviderAndWait(context.Background(), provider)
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
	return api().SetProvider(ctx, provider)
}

// SetProviderWithContextAndWait sets the default [FeatureProvider] with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this function for synchronous provider setup with guaranteed readiness when you need
// application startup to wait for the provider before continuing.
// Recommended timeout values: 1-5s for local providers, 10-30s for network-based providers.
func SetProviderWithContextAndWait(ctx context.Context, provider FeatureProvider) error {
	return api().SetProviderAndWait(ctx, provider)
}

// ProviderMetadata returns the default [FeatureProvider] metadata
func ProviderMetadata() Metadata {
	return api().getProviderMetadata()
}

// SetNamedProvider sets a [FeatureProvider] mapped to the given [Client] domain. Provider initialization is asynchronous
// and status can be checked from provider status
func SetNamedProvider(domain string, provider FeatureProvider) error {
	return api().SetProvider(context.Background(), provider, WithDomain(domain))
}

// SetNamedProviderAndWait sets a provider mapped to the given [Client] domain and waits for its initialization.
// Returns an error if initialization cause error
func SetNamedProviderAndWait(domain string, provider FeatureProvider) error {
	return api().SetProviderAndWait(context.Background(), provider, WithDomain(domain))
}

// SetNamedProviderWithContext sets a [FeatureProvider] mapped to the given [Client] domain with context-aware initialization.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Named providers allow different domains to use different feature flag providers,
// enabling multi-tenant applications or microservice architectures.
func SetNamedProviderWithContext(ctx context.Context, domain string, provider FeatureProvider) error {
	return api().SetProvider(ctx, provider, WithDomain(domain))
}

// SetNamedProviderWithContextAndWait sets a provider mapped to the given [Client] domain with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this for synchronous named provider setup where you need to ensure
// the provider is ready before proceeding.
func SetNamedProviderWithContextAndWait(ctx context.Context, domain string, provider FeatureProvider) error {
	return api().SetProviderAndWait(ctx, provider, WithDomain(domain))
}

// NamedProviderMetadata returns the named provider's Metadata
func NamedProviderMetadata(name string) Metadata {
	return api().getDomainProviderMetadata(name)
}

// SetEvaluationContext sets the global [EvaluationContext].
func SetEvaluationContext(evalCtx EvaluationContext) {
	api().SetEvaluationContext(evalCtx)
}

// AddHooks appends to the collection of any previously added hooks
func AddHooks(hooks ...Hook) {
	api().AddHooks(hooks...)
}

// AddHandler allows to add API level event handlers
func AddHandler(eventType EventType, callback EventCallback) {
	api().AddHandler(eventType, callback)
}

// RemoveHandler allows for removal of API level event handlers
func RemoveHandler(eventType EventType, callback EventCallback) {
	api().RemoveHandler(eventType, callback)
}

// Shutdown unconditionally calls shutdown on all registered providers,
// regardless of their state. It resets the state of the API, removing all
// hooks, event handlers, and providers.
func Shutdown() {
	_ = ShutdownWithContext(context.Background())
}

// ShutdownWithContext calls context-aware shutdown on all registered providers.
// If providers implement ContextAwareStateHandler, ShutdownWithContext will be called with the provided context.
// It resets the state of the API, removing all hooks, event handlers, and providers.
// This is intended to be called when your application is terminating.
// Returns an error if any provider shutdown fails or if context is cancelled during shutdown.
func ShutdownWithContext(ctx context.Context) error {
	return resetSingletonWithContext(ctx)
}
