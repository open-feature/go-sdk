package openfeature

// EvaluationContext
// https://github.com/open-feature/spec/blob/main/specification/evaluation-context/evaluation-context.md
type EvaluationContext struct {
	TargetingKey string                 `json:"targetingKey"` // uniquely identifying the subject (end-user, or client service) of a flag evaluation
	Attributes   map[string]interface{} `json:"attributes"`
}
