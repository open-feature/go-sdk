package providers

import "github.com/open-feature/golang-sdk/pkg/openfeature"

// NoopProvider implements the FeatureProvider interface and provides functions for evaluating flags
type NoopProvider struct {}

// Name returns the name of the provider
func (e NoopProvider) Name() string {
	return "NoopProvider"
}

// GetBooleanEvaluation returns a boolean flag.
func (e NoopProvider) GetBooleanEvaluation(flag string, defaultValue bool, evalCtx openfeature.EvaluationContext, options ...openfeature.EvaluationOption) openfeature.BoolResolutionDetail {
	return openfeature.BoolResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: openfeature.ResolutionDetail{
			Variant: "default-variant",
			Reason: openfeature.DEFAULT,
		},
	}
}

// GetStringEvaluation returns a string flag.
func (e NoopProvider) GetStringEvaluation(flag string, defaultValue string, evalCtx openfeature.EvaluationContext, options ...openfeature.EvaluationOption) openfeature.StringResolutionDetail {
	return openfeature.StringResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: openfeature.ResolutionDetail{
			Variant: "default-variant",
			Reason: openfeature.DEFAULT,
		},
	}
}

// GetNumberEvaluation returns a number flag.
func (e NoopProvider) GetNumberEvaluation(flag string, defaultValue int64, evalCtx openfeature.EvaluationContext, options ...openfeature.EvaluationOption) openfeature.NumberResolutionDetail {
	return openfeature.NumberResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: openfeature.ResolutionDetail{
			Variant: "default-variant",
			Reason: openfeature.DEFAULT,
		},
	}
}

// GetObjectEvaluation returns an object flag
func (e NoopProvider) GetObjectEvaluation(flag string, defaultValue interface{}, evalCtx openfeature.EvaluationContext, options ...openfeature.EvaluationOption) openfeature.ResolutionDetail {
	return openfeature.ResolutionDetail{
		Value: defaultValue,
		Variant: "default-variant",
		Reason: openfeature.DEFAULT,
	}
}