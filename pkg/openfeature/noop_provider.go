package openfeature

// NoopProvider implements the FeatureProvider interface and provides functions for evaluating flags
type NoopProvider struct{}

// Metadata returns the metadata of the provider
func (e NoopProvider) Metadata() Metadata {
	return Metadata{Name: "NoopProvider"}
}

// GetBooleanEvaluation returns a boolean flag.
func (e NoopProvider) GetBooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// GetStringEvaluation returns a string flag.
func (e NoopProvider) GetStringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) StringResolutionDetail {
	return StringResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// GetNumberEvaluation returns a number flag.
func (e NoopProvider) GetNumberEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) NumberResolutionDetail {
	return NumberResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// GetObjectEvaluation returns an object flag
func (e NoopProvider) GetObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) ResolutionDetail {
	return ResolutionDetail{
		Value:   defaultValue,
		Variant: "default-variant",
		Reason:  DEFAULT,
	}
}
