package openfeature

import openfeature "github.com/open-feature/go-sdk"

// EvaluationContext provides ambient information for the purposes of flag evaluation
// The use of the constructor, NewEvaluationContext, is enforced to set EvaluationContext's fields in order
// to enforce immutability.
// https://openfeature.dev/specification/sections/evaluation-context
//
// Deprecated: use
// github.com/open-feature/go-sdk.EvaluationContext, instead.
type EvaluationContext = openfeature.EvaluationContext

// NewEvaluationContext constructs an EvaluationContext
//
// targetingKey - uniquely identifying the subject (end-user, or client service) of a flag evaluation
// attributes - contextual data used in flag evaluation
//
// Deprecated: use
// github.com/open-feature/go-sdk.NewEvaluationContext, instead.
func NewEvaluationContext(targetingKey string, attributes map[string]interface{}) EvaluationContext {
	return openfeature.NewEvaluationContext(targetingKey, attributes)
}

// NewTargetlessEvaluationContext constructs an EvaluationContext with an empty targeting key
//
// attributes - contextual data used in flag evaluation
//
// Deprecated: use
// github.com/open-feature/go-sdk.NewTargetlessEvaluationContext,
// instead.
func NewTargetlessEvaluationContext(attributes map[string]interface{}) EvaluationContext {
	return openfeature.NewTargetlessEvaluationContext(attributes)
}
