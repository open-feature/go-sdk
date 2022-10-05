package openfeature

import "sync"

// EvaluationContext implementation(s):
//
// MutableEvaluationContext - construct via NewMutableEvaluationContext
type EvaluationContext interface {
	TargetingKey() string
	SetTargetingKey(key string)
	Attributes() map[string]interface{}
	Attribute(key string) interface{}
	SetAttribute(key string, value interface{})
	SetAttributes(attrs map[string]interface{})
	Merge(evaluationContexts ...EvaluationContext) EvaluationContext
}

// MutableEvaluationContext provides mutable ambient information for the purposes of flag evaluation.
// The use of the constructor, NewMutableEvaluationContext, is enforced to set MutableEvaluationContext's fields in order
// to enforce thread safe mutation.
type MutableEvaluationContext struct {
	mx           sync.Mutex
	targetingKey string // uniquely identifying the subject (end-user, or client service) of a flag evaluation
	attributes   map[string]interface{}
}

// NewMutableEvaluationContext constructs a MutableEvaluationContext
//
// targetingKey - uniquely identifying the subject (end-user, or client service) of a flag evaluation
// attributes - contextual data used in flag evaluation
func NewMutableEvaluationContext(targetingKey string, attributes map[string]interface{}) *MutableEvaluationContext {
	m := &MutableEvaluationContext{targetingKey: targetingKey, attributes: make(map[string]interface{}, len(attributes))}
	m.SetAttributes(attributes)

	return m
}

func (m *MutableEvaluationContext) TargetingKey() string {
	return m.targetingKey
}

func (m *MutableEvaluationContext) SetTargetingKey(key string) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.targetingKey = key
}

// Attributes returns a copy of the underlying map in order to avoid passing the internal map by reference (as this would
// violate the thread safe mutability)
func (m *MutableEvaluationContext) Attributes() map[string]interface{} {
	m.mx.Lock()
	defer m.mx.Unlock()
	if len(m.attributes) == 0 {
		return nil
	}
	// forced to return a copy of the underlying map in order to avoid mutation outside the mutex
	attrs := make(map[string]interface{}, len(m.attributes))
	for key, value := range m.attributes {
		attrs[key] = value
	}

	return attrs
}

func (m *MutableEvaluationContext) Attribute(key string) interface{} {
	m.mx.Lock()
	defer m.mx.Unlock()
	return m.attributes[key]
}

// SetAttribute writes the key value pair to the underlying map
func (m *MutableEvaluationContext) SetAttribute(key string, value interface{}) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.attributes[key] = value
}

// SetAttributes merges the given attributes with higher priority than the existing attributes
func (m *MutableEvaluationContext) SetAttributes(attrs map[string]interface{}) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for key, value := range attrs {
		m.attributes[key] = value
	}
}

// Merge merges attributes from the given EvaluationContexts with the nth EvaluationContext taking precedence in case
// of any conflicts with the (n+1)th EvaluationContext
func (m *MutableEvaluationContext) Merge(evaluationContexts ...EvaluationContext) EvaluationContext {
	mergedCtx := &MutableEvaluationContext{
		targetingKey: m.TargetingKey(),
		attributes:   m.Attributes(),
	}

	if len(evaluationContexts) == 0 {
		return mergedCtx
	}

	for i := 0; i < len(evaluationContexts); i++ {
		if mergedCtx.TargetingKey() == "" && evaluationContexts[i].TargetingKey() != "" {
			mergedCtx.targetingKey = evaluationContexts[i].TargetingKey()
		}

		for k, v := range evaluationContexts[i].Attributes() {
			_, ok := mergedCtx.attributes[k]
			if !ok {
				mergedCtx.attributes[k] = v
			}
		}
	}

	return mergedCtx
}
