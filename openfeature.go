package openfeature

import (
	"context"
	"strings"
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

// SetProvider sets the [FeatureProvider] with context-aware initialization.
// If the provider implements StateHandler, Init will be called with the provided context.
// Provider initialization is asynchronous and status can be checked from provider status.
// Returns an error immediately if provider is nil, or if context is cancelled during setup.
//
// Use this function for non-blocking provider setup with timeout control where you want
// to continue application startup while the provider initializes in background.
func SetProvider(ctx context.Context, provider FeatureProvider, opts ...CallOption) error {
	c := newCallOption(opts...)
	if c.domain != "" {
		return api.SetNamedProvider(ctx, c.domain, provider)
	}
	return api.SetProvider(ctx, provider)
}

// SetProviderAndWait sets the [FeatureProvider] with initialization and waits for completion.
// If the provider implements StateHandler, InitWithContext will be called with the provided context.
// Returns an error if initialization causes an error, or if context is cancelled during initialization.
//
// Use this function for synchronous provider setup with guaranteed readiness when you need
// application startup to wait for the provider before continuing.
// Recommended timeout values: 1-5s for local providers, 10-30s for network-based providers.
func SetProviderAndWait(ctx context.Context, provider FeatureProvider, opts ...CallOption) error {
	c := newCallOption(opts...)
	if c.domain != "" {
		return api.SetNamedProviderAndWait(ctx, c.domain, provider)
	}
	return api.SetProviderAndWait(ctx, provider)
}

// ProviderMetadata returns the [FeatureProvider] metadata
func ProviderMetadata(opts ...CallOption) Metadata {
	c := newCallOption(opts...)
	if c.domain != "" {
		return api.GetNamedProviderMetadata(c.domain)
	}
	return api.GetProviderMetadata()
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

type (
	callOption struct {
		domain string
	}
	CallOption func(*callOption)
)

// WithDomain is an option which allows different domains to use different feature flag providers or clients.
// It could be used with [NewClient], [SetProvider], [SetProviderAndWait] and [ProviderMetadata].
func WithDomain(domain string) CallOption {
	return func(co *callOption) {
		co.domain = strings.TrimSpace(domain)
	}
}

func newCallOption(opts ...CallOption) callOption {
	c := callOption{}
	for _, o := range opts {
		o(&c)
	}
	return c
}
