package openfeature

import (
	"context"
	"maps"
)

// Hook allows application developers to add arbitrary behavior to the flag evaluation lifecycle.
// They operate similarly to middleware in many web frameworks.
// https://github.com/open-feature/spec/blob/main/specification/sections/04-hooks.md
type Hook interface {
	Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (*EvaluationContext, error)
	After(ctx context.Context, hookContext HookContext, flagEvaluationDetails InterfaceEvaluationDetails, hookHints HookHints) error
	Error(ctx context.Context, hookContext HookContext, err error, hookHints HookHints)
	Finally(ctx context.Context, hookContext HookContext, flagEvaluationDetails InterfaceEvaluationDetails, hookHints HookHints)
}

// HookHints contains a map of hints for hooks
type HookHints struct {
	mapOfHints map[string]any
}

// NewHookHints constructs HookHints
//
// mapOfHints - arbitrary data made available to hooks. The map is copied, so
// changes the caller makes to it afterwards are not reflected in the hints.
// The copy is shallow: values that are themselves mutable remain shared.
func NewHookHints(mapOfHints map[string]any) HookHints {
	// copy hints to new map to avoid reference being externally available, thereby enforcing immutability
	hints := make(map[string]any, len(mapOfHints))
	maps.Copy(hints, mapOfHints)

	return HookHints{mapOfHints: hints}
}

// Value returns the value at the given key in the underlying map.
func (h HookHints) Value(key string) any {
	return h.mapOfHints[key]
}

// HookContext defines the base level fields of a hook context
type HookContext struct {
	flagKey           string
	flagType          Type
	defaultValue      any
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
func (h HookContext) DefaultValue() any {
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

// NewHookContext constructs HookContext
// Allows for simplified hook test cases while maintaining immutability
func NewHookContext(
	flagKey string,
	flagType Type,
	defaultValue any,
	clientMetadata ClientMetadata,
	providerMetadata Metadata,
	evaluationContext EvaluationContext,
) HookContext {
	return HookContext{
		flagKey:           flagKey,
		flagType:          flagType,
		defaultValue:      defaultValue,
		clientMetadata:    clientMetadata,
		providerMetadata:  providerMetadata,
		evaluationContext: evaluationContext,
	}
}

// check at compile time that UnimplementedHook implements the Hook interface
var _ Hook = UnimplementedHook{}

// UnimplementedHook implements all hook methods with empty functions
// Include UnimplementedHook in your hook struct to avoid defining empty functions
// e.g.
//
//	type MyHook struct {
//	  UnimplementedHook
//	}
type UnimplementedHook struct{}

func (UnimplementedHook) Before(context.Context, HookContext, HookHints) (*EvaluationContext, error) {
	return nil, nil
}

func (UnimplementedHook) After(context.Context, HookContext, InterfaceEvaluationDetails, HookHints) error {
	return nil
}

func (UnimplementedHook) Error(context.Context, HookContext, error, HookHints) {}

func (UnimplementedHook) Finally(context.Context, HookContext, InterfaceEvaluationDetails, HookHints) {
}
