// Package factory bridges the openfeature package's isolated-instance
// constructor to the openfeature/isolated sub-package without exposing it
// publicly. Internal-only.
package factory

// NewAPI is set by openfeature.init and read by openfeature/isolated.NewAPI.
// Returns the openfeature evaluation API as any to avoid an import cycle;
// callers must type-assert to [openfeature.IEvaluation].
var NewAPI func() any
