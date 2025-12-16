package openfeature

import (
	"context"
)

// iClient defines the behaviour required of an OpenFeature client
type iClient interface {
	Metadata() ClientMetadata
	Evaluator
	DetailEvaluator
	IEventing
	Tracker
}

// Evaluator defines OpenFeature evaluator contract.
type Evaluator interface {
	Boolean(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) bool
	String(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) string
	Float(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) float64
	Int(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) int64
	Object(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) any
}

// DetailEvaluator defines OpenFeature details evaluator contract.
type DetailEvaluator interface {
	BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
	StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
	FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
	IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
	ObjectValueDetails(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (ObjectEvaluationDetails, error)
}

// IEventing defines the OpenFeature eventing contract
type IEventing interface {
	AddHandler(eventType EventType, callback EventCallback)
	RemoveHandler(eventType EventType, callback EventCallback)
}

// evaluationImpl is an internal reference interface extending IEvaluation
type evaluationImpl interface {
	GetProviderMetadata() Metadata
	GetNamedProviderMetadata(name string) Metadata
	SetProvider(ctx context.Context, provider FeatureProvider) error
	SetProviderAndWait(ctx context.Context, provider FeatureProvider) error
	SetNamedProvider(ctx context.Context, clientName string, provider FeatureProvider) error
	SetNamedProviderAndWait(ctx context.Context, clientName string, provider FeatureProvider) error
	GetClient() *Client
	GetNamedClient(clientName string) *Client
	SetEvaluationContext(evalCtx EvaluationContext)
	AddHooks(hooks ...Hook)
	Shutdown(ctx context.Context) error
	IEventing
	GetProvider() FeatureProvider
	GetNamedProviders() map[string]FeatureProvider
	GetHooks() []Hook

	ForEvaluation(clientName string) (FeatureProvider, []Hook, EvaluationContext)
}

// eventingImpl is an internal reference interface extending IEventing
type eventingImpl interface {
	IEventing
	GetAPIRegistry() map[EventType][]EventCallback
	GetClientRegistry(client string) scopedCallback

	clientEvent
}

// clientEvent is an internal reference for OpenFeature Client events
type clientEvent interface {
	AddClientHandler(clientName string, t EventType, c EventCallback)
	RemoveClientHandler(name string, t EventType, c EventCallback)

	State(domain string) State
}
