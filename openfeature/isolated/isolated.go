// Package isolated provides a factory for creating OpenFeature API instances with state independent of the global singleton.
//
// Use isolated instances when you need DI, multi-tenancy, or test isolation.
//
// Experimental.
package isolated

import (
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/internal/factory"
)

// NewAPI returns a new, independent [openfeature.EvaluationAPI].
//
// Experimental.
func NewAPI() *openfeature.EvaluationAPI {
	return factory.NewAPI().(*openfeature.EvaluationAPI)
}
