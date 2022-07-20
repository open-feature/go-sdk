package openfeature

// Hook Hooks are a mechanism whereby application developers can add arbitrary behavior to flag evaluation. They operate similarly to middleware in many web frameworks.
// https://github.com/open-feature/spec/blob/main/specification/flag-evaluation/hooks.md
type Hook interface {
	Before(hookContext HookContext, hookHints HookHints) (*EvaluationContext, error)
	After(hookContext HookContext, flagEvaluationDetails EvaluationDetails, hookHints HookHints) error
	Error(hookContext HookContext, err error, hookHints HookHints)
	Finally(hookContext HookContext, hookHints HookHints)
}

// HookHints contains a map of hints for hooks
type HookHints struct {
	mapOfHints map[string]interface{}
}

// NewHookHints constructs HookHints
func NewHookHints(mapOfHints map[string]interface{}) HookHints {
	return HookHints{mapOfHints: mapOfHints}
}

// Value returns the value at the given key in the underlying map.
// Maintains immutability of the map.
func (h HookHints) Value(key string) interface{} {
	return h.mapOfHints[key]
}

// HookContext defines the base level fields of a hook context
type HookContext struct {
	flagKey           string
	flagType          Type
	defaultValue      interface{}
	clientMetadata    ClientMetadata
	providerMetadata  Metadata
	evaluationContext EvaluationContext
}

// FlagKey returns the hook context's flag key
func (h HookContext) FlagKey() string {
	return h.flagKey
}

// FlagType returns the hook context's flag type
func (h HookContext) FlagType() Type {
	return h.flagType
}

// DefaultValue returns the hook context's default value
func (h HookContext) DefaultValue() interface{} {
	return h.defaultValue
}

// ClientMetadata returns the client's metadata
func (h HookContext) ClientMetadata() ClientMetadata {
	return h.clientMetadata
}

// ProviderMetadata returns the provider's metadata
func (h HookContext) ProviderMetadata() Metadata {
	return h.providerMetadata
}

// EvaluationContext returns the hook context's EvaluationContext
func (h HookContext) EvaluationContext() EvaluationContext {
	return h.evaluationContext
}
