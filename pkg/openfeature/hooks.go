package openfeature

import (
	openfeature "github.com/open-feature/go-sdk"
)

// Hook allows application developers to add arbitrary behavior to the flag evaluation lifecycle.
// They operate similarly to middleware in many web frameworks.
// https://github.com/open-feature/spec/blob/main/specification/hooks.md
//
// Deprecated: use github.com/open-feature/go-sdk.Hook, instead.
type Hook = openfeature.Hook

// HookHints contains a map of hints for hooks
//
// Deprecated: use github.com/open-feature/go-sdk.HookHints,
// instead.
type HookHints = openfeature.HookHints

// NewHookHints constructs HookHints
//
// Deprecated: use github.com/open-feature/go-sdk.NewHookHints,
// instead.
func NewHookHints(mapOfHints map[string]interface{}) HookHints {
	return openfeature.NewHookHints(mapOfHints)
}

// HookContext defines the base level fields of a hook context
//
// Deprecated: use github.com/open-feature/go-sdk.HookContext,
// instead.
type HookContext = openfeature.HookContext

// NewHookContext constructs HookContext
// Allows for simplified hook test cases while maintaining immutability
//
// Deprecated: use github.com/open-feature/go-sdk.NewHookContext,
// instead.
func NewHookContext(
	flagKey string,
	flagType Type,
	defaultValue interface{},
	clientMetadata ClientMetadata,
	providerMetadata Metadata,
	evaluationContext EvaluationContext,
) HookContext {
	return openfeature.NewHookContext(flagKey, flagType, defaultValue, clientMetadata, providerMetadata, evaluationContext)
}

// UnimplementedHook implements all hook methods with empty functions
// Include UnimplementedHook in your hook struct to avoid defining empty functions
// e.g.
//
//	type MyHook = openfeature.MyHook
//	  UnimplementedHook
//	}
//
// Deprecated: use
// github.com/open-feature/go-sdk.UnimplementedHook, instead.
type UnimplementedHook = openfeature.UnimplementedHook
