// Package factory is an internal bridge that exposes the openfeature
// package's isolated-instance constructor to the openfeature/isolated
// sub-package without making it part of the public API.
//
// The openfeature package sets NewAPI in its init function; openfeature/isolated
// reads it. External callers cannot import this package (Go's internal/ rule
// restricts imports to paths under openfeature/).
package factory

// NewAPI is set by the openfeature package's init. It returns a new
// *openfeature.EvaluationAPI as any to avoid an import cycle with openfeature.
// Callers in openfeature/isolated must type-assert.
var NewAPI func() any
