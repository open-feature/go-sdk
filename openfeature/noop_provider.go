package openfeature

import "context"

// NoopProvider implements the FeatureProvider interface and provides functions for evaluating flags
type NoopProvider struct{}

// Metadata returns the metadata of the provider
func (e NoopProvider) Metadata() Metadata {
	return Metadata{Name: "NoopProvider"}
}

// BooleanEvaluation returns a boolean flag.
func (e NoopProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx FlattenedContext) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{
			Variant: "default-variant",
			Reason:  DefaultReason,
		},
	}
}

// StringEvaluation returns a string flag.
func (e NoopProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx FlattenedContext) StringResolutionDetail {
	return StringResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{
			Variant: "default-variant",
			Reason:  DefaultReason,
		},
	}
}

// FloatEvaluation returns a float flag.
func (e NoopProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx FlattenedContext) FloatResolutionDetail {
	return FloatResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{
			Variant: "default-variant",
			Reason:  DefaultReason,
		},
	}
}

// IntEvaluation returns an int flag.
func (e NoopProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx FlattenedContext) IntResolutionDetail {
	return IntResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{
			Variant: "default-variant",
			Reason:  DefaultReason,
		},
	}
}

// ObjectEvaluation returns an object flag
func (e NoopProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx FlattenedContext) InterfaceResolutionDetail {
	return InterfaceResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{
			Variant: "default-variant",
			Reason:  DefaultReason,
		},
	}
}

// Hooks returns hooks
func (e NoopProvider) Hooks() []Hook {
	return []Hook{}
}

func (e NoopProvider) Track(ctx context.Context, eventName string, evalCtx EvaluationContext, details TrackingEventDetails) {
}
