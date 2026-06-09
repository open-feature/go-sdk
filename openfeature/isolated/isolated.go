// Package isolated provides the factory for creating isolated OpenFeature
// API instances.
//
// Each call to [NewAPI] returns a new, independent [openfeature.EvaluationAPI]
// with its own state (providers, evaluation context, hooks, event handlers).
// Instances do not share state with the global [openfeature] singleton or
// with each other.
//
// This factory lives in a distinct package from the global singleton to be
// intentionally less discoverable, per spec requirement 1.8.3. Use isolated
// instances for dependency-injection frameworks, multi-tenant scenarios,
// or testing in isolation from the global singleton.
//
// Experimental: this API is part of spec section 1.8 which is experimental.
package isolated

import (
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/internal/factory"
)

// NewAPI returns a new, independent [openfeature.EvaluationAPI] instance.
//
// Each instance conforms to the same contract as the global singleton
// (spec 1.8.2). Per spec 1.8.4, a provider instance SHOULD NOT be bound to
// more than one API instance at a time; attempting to do so will return an
// error from [openfeature.EvaluationAPI.SetProvider] or
// [openfeature.EvaluationAPI.SetNamedProvider].
//
// Callers MUST invoke [openfeature.EvaluationAPI.Shutdown] when the instance
// is no longer needed to release provider resources.
//
// Experimental: this API is part of spec section 1.8 which is experimental.
func NewAPI() *openfeature.EvaluationAPI {
	return factory.NewAPI().(*openfeature.EvaluationAPI)
}
