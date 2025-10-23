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
// Use this function when you want:
// - Non-blocking provider setup with timeout control
// - To continue application startup while provider initializes in background
// - Fine-grained control over initialization timeouts
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	err := SetProviderWithContext(ctx, myProvider)
//	if err != nil {
//		log.Printf("Failed to start provider setup: %v", err)
//		return err
//	}
//	// Provider continues initializing in background
//
// For providers that don't implement ContextAwareStateHandler, this behaves
// identically to SetProvider() but with timeout protection.
func SetProviderWithContext(ctx context.Context, provider FeatureProvider) error {
	return api.SetProviderWithContext(ctx, provider)
}

// SetProviderWithContextAndWait sets the default [FeatureProvider] with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this function when you need:
// - Synchronous provider setup with guaranteed readiness
// - Application startup to wait for provider before continuing
// - Error handling for initialization failures
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	err := SetProviderWithContextAndWait(ctx, myProvider)
//	if err != nil {
//		if errors.Is(err, context.DeadlineExceeded) {
//			log.Println("Provider initialization timed out")
//		} else {
//			log.Printf("Provider initialization failed: %v", err)
//		}
//		return err
//	}
//	// Provider is now ready to use
//
// Recommended timeout values:
// - Local/in-memory providers: 1-5 seconds
// - Network-based providers: 10-30 seconds
// - Database-dependent providers: 15-60 seconds
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
//
// Example:
//
//	// Set up different providers for different services
//	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancel()
//
//	err := SetNamedProviderWithContext(ctx, "user-service", userProvider)
//	if err != nil {
//		return fmt.Errorf("failed to setup user service provider: %w", err)
//	}
//
//	err = SetNamedProviderWithContext(ctx, "billing-service", billingProvider)
//	if err != nil {
//		return fmt.Errorf("failed to setup billing service provider: %w", err)
//	}
//
//	// Create clients for different domains
//	userClient := NewClient("user-service")
//	billingClient := NewClient("billing-service")
func SetNamedProviderWithContext(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContext(ctx, domain, provider, true)
}

// SetNamedProviderWithContextAndWait sets a provider mapped to the given [Client] domain with context-aware initialization and waits for completion.
// If the provider implements ContextAwareStateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this for synchronous named provider setup where you need to ensure
// the provider is ready before proceeding.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// Wait for critical providers to be ready
//	if err := SetNamedProviderWithContextAndWait(ctx, "critical-service", criticalProvider); err != nil {
//		return fmt.Errorf("critical provider failed to initialize: %w", err)
//	}
//
//	// Now safe to use the client
//	client := NewClient("critical-service")
//	enabled, _ := client.BooleanValue(context.Background(), "feature-x", false, EvaluationContext{})
func SetNamedProviderWithContextAndWait(ctx context.Context, domain string, provider FeatureProvider) error {
	return api.SetNamedProviderWithContext(ctx, domain, provider, false)
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
