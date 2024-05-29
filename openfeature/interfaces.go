package openfeature

import "context"

// IEvaluation defines the OpenFeature API contract
type IEvaluation interface {
	SetProvider(provider FeatureProvider) error
	SetProviderAndWait(provider FeatureProvider) error
	GetProviderMetadata() Metadata
	SetNamedProvider(clientName string, provider FeatureProvider, async bool) error
	GetNamedProviderMetadata(name string) Metadata
	GetClient() IClient
	GetNamedClient(clientName string) IClient
	SetEvaluationContext(apiCtx EvaluationContext)
	AddHooks(hooks ...Hook)
	Shutdown()
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
	ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error)
	BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
	StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
	FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
	IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
	ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)

	Boolean(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) bool
	String(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) string
	Float(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) float64
	Int(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) int64
	Object(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) interface{}

	IEventing
}

// IEventing defines the OpenFeature eventing contract
type IEventing interface {
	AddHandler(eventType EventType, callback EventCallback)
	RemoveHandler(eventType EventType, callback EventCallback)
}
