package openfeature

import (
	"context"
)

// IEvaluation defines the OpenFeature API contract
//
// Deprecated: IEvaluation will be removed in v1.20. Use [EvaluationAPI] instead.
type IEvaluation interface {
	SetProvider(provider FeatureProvider) error
	SetProviderAndWait(provider FeatureProvider) error
	SetProviderWithContext(ctx context.Context, provider FeatureProvider) error
	SetProviderWithContextAndWait(ctx context.Context, provider FeatureProvider) error
	GetProviderMetadata() Metadata
	SetNamedProvider(clientName string, provider FeatureProvider, async bool) error
	SetNamedProviderWithContext(ctx context.Context, clientName string, provider FeatureProvider, async bool) error
	SetNamedProviderWithContextAndWait(ctx context.Context, clientName string, provider FeatureProvider) error
	GetNamedProviderMetadata(name string) Metadata
	GetClient() IClient
	GetNamedClient(clientName string) IClient
	SetEvaluationContext(evalCtx EvaluationContext)
	AddHooks(hooks ...Hook)
	Shutdown()
	ShutdownWithContext(ctx context.Context) error
	IEventing
}

// IClient defines the behaviour required of an OpenFeature client
type IClient interface {
	Metadata() ClientMetadata
	AddHooks(hooks ...Hook)
	SetEvaluationContext(evalCtx EvaluationContext)
	EvaluationContext() EvaluationContext
	BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
	StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
	FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
	IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
	ObjectValue(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (any, error)
	BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
	StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
	FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
	IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
	ObjectValueDetails(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)

	Boolean(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) bool
	String(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) string
	Float(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) float64
	Int(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) int64
	Object(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) any

	State() State

	IEventing
	Tracker
}

// IEventing defines the OpenFeature eventing contract
type IEventing interface {
	AddHandler(eventType EventType, callback EventCallback)
	RemoveHandler(eventType EventType, callback EventCallback)
}

type providerBindingFn func(domain string) (FeatureProvider, []Hook, EvaluationContext)

// clientEvent is an internal reference for OpenFeature Client events
type clientEvent interface {
	AddClientHandler(clientName string, t EventType, c EventCallback)
	RemoveClientHandler(name string, t EventType, c EventCallback)

	State(domain string) State
}
