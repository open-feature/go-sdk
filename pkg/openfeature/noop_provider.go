package openfeature

// NoopProvider implements the FeatureProvider interface and provides functions for evaluating flags
type NoopProvider struct{}

// Metadata returns the metadata of the provider
func (e NoopProvider) Metadata() Metadata {
	return Metadata{Name: "NoopProvider"}
}

// BooleanEvaluation returns a boolean flag.
func (e NoopProvider) BooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// StringEvaluation returns a string flag.
func (e NoopProvider) StringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) StringResolutionDetail {
	return StringResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// NumberEvaluation returns a number flag.
func (e NoopProvider) NumberEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) NumberResolutionDetail {
	return NumberResolutionDetail{
		Value: defaultValue,
		ResolutionDetail: ResolutionDetail{
			Variant: "default-variant",
			Reason:  DEFAULT,
		},
	}
}

// ObjectEvaluation returns an object flag
func (e NoopProvider) ObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) ResolutionDetail {
	return ResolutionDetail{
		Value:   defaultValue,
		Variant: "default-variant",
		Reason:  DEFAULT,
	}
}
