package openfeature

// EvaluationContext provides ambient information for the purposes of flag evaluation
// The use of the constructor, NewEvaluationContext, is enforced to set EvaluationContext's fields in order
// to enforce immutability.
// https://github.com/open-feature/spec/blob/main/specification/evaluation-context/evaluation-context.md
type EvaluationContext struct {
	targetingKey string // uniquely identifying the subject (end-user, or client service) of a flag evaluation
	attributes   map[string]interface{}
}

// Attribute retrieves the attribute with the given key
func (e EvaluationContext) Attribute(key string) interface{} {
	return e.attributes[key]
}

// TargetingKey returns the key uniquely identifying the subject (end-user, or client service) of a flag evaluation
func (e EvaluationContext) TargetingKey() string {
	return e.targetingKey
}

// Attributes returns a copy of the EvaluationContext's attributes
func (e EvaluationContext) Attributes() map[string]interface{} {
	// copy attributes to new map to prevent mutation (maps are passed by reference)
	attrs := make(map[string]interface{}, len(e.attributes))
	for key, value := range e.attributes {
		attrs[key] = value
	}

	return attrs
}

// NewEvaluationContext constructs an EvaluationContext
//
// targetingKey - uniquely identifying the subject (end-user, or client service) of a flag evaluation
// attributes - contextual data used in flag evaluation
func NewEvaluationContext(targetingKey string, attributes map[string]interface{}) EvaluationContext {
	// copy attributes to new map to avoid reference being externally available, thereby enforcing immutability
	attrs := make(map[string]interface{}, len(attributes))
	for key, value := range attributes {
		attrs[key] = value
	}

	return EvaluationContext{
		targetingKey: targetingKey,
		attributes:   attrs,
	}
}
