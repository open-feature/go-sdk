package memprovider

import (
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

const (
	// Deprecated: use
	// github.com/open-feature/go-sdk/openfeature/memprovider.Enabled,
	// instead.
	Enabled = memprovider.Enabled
	// Deprecated: use
	// github.com/open-feature/go-sdk/openfeature/memprovider.Disabled,
	// instead.
	Disabled = memprovider.Disabled
)

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature/memprovider.InMemoryProvider,
// instead.
type InMemoryProvider = memprovider.InMemoryProvider

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature/memprovider.NewInMemoryProvider,
// instead.
func NewInMemoryProvider(from map[string]InMemoryFlag) InMemoryProvider {
	return memprovider.NewInMemoryProvider(from)
}

// Type Definitions for InMemoryProvider flag

// State of the feature flag
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature/memprovider.State, instead.
type State = memprovider.State

// ContextEvaluator is a callback to perform openfeature.EvaluationContext backed evaluations.
// This is a callback implemented by the flag definer.
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature/memprovider.ContextEvaluator,
// instead.
type ContextEvaluator = memprovider.ContextEvaluator

// InMemoryFlag is the feature flag representation accepted by InMemoryProvider
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature/memprovider.InMemoryFlag,
// instead.
type InMemoryFlag = memprovider.InMemoryFlag
