package openfeature

import (
	"github.com/open-feature/go-sdk/openfeature"
)

// IClient defines the behaviour required of an openfeature client
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.IClient, instead.
type IClient = openfeature.IClient

// ClientMetadata provides a client's metadata
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.ClientMetadata,
// instead.
type ClientMetadata = openfeature.ClientMetadata

// NewClientMetadata constructs ClientMetadata
// Allows for simplified hook test cases while maintaining immutability
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.NewClientMetadata, instead.
func NewClientMetadata(name string) ClientMetadata {
	return openfeature.NewClientMetadata(name)
}

// Client implements the behaviour required of an openfeature client
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.Client, instead.
type Client = openfeature.Client

// NewClient returns a new Client. Name is a unique identifier for this client
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewClient,
// instead.
func NewClient(name string) *Client {
	return openfeature.NewClient(name)
}

// Type represents the type of a flag
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.Type, instead.
type Type = openfeature.Type

const (
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.Boolean,
	// instead.
	Boolean = openfeature.Boolean
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.String,
	// instead.
	String = openfeature.String
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.Float,
	// instead.
	Float = openfeature.Float
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.Int,
	// instead.
	Int = openfeature.Int
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.Object,
	// instead.
	Object = openfeature.Object
)

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.EvaluationDetails, instead.
type EvaluationDetails = openfeature.EvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.BooleanEvaluationDetails,
// instead.
type BooleanEvaluationDetails = openfeature.BooleanEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.StringEvaluationDetails, instead.
type StringEvaluationDetails = openfeature.StringEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.FloatEvaluationDetails, instead.
type FloatEvaluationDetails = openfeature.FloatEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.IntEvaluationDetails, instead.
type IntEvaluationDetails = openfeature.IntEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.InterfaceEvaluationDetails,
// instead.
type InterfaceEvaluationDetails = openfeature.InterfaceEvaluationDetails

// Deprecated: use github.com/open-feature/go-sdk/openfeature.ResolutionDetail, instead.
type ResolutionDetail = openfeature.ResolutionDetail

// FlagMetadata is a structure which supports definition of arbitrary properties, with keys of type string, and values
// of type boolean, string, int64 or float64. This structure is populated by a provider for use by an Application
// Author (via the Evaluation API) or an Application Integrator (via hooks).
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.FlagMetadata,
// instead.
type FlagMetadata = openfeature.FlagMetadata

// Option applies a change to EvaluationOptions
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.Option, instead.
type Option = openfeature.Option

// EvaluationOptions should contain a list of hooks to be executed for a flag evaluation
//
// Deprecated: use
// github.com/open-feature/go-sdk/openfeature.EvaluationOptions, instead.
type EvaluationOptions = openfeature.EvaluationOptions

// WithHooks applies provided hooks.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.WithHooks,
// instead.
func WithHooks(hooks ...Hook) Option {
	return openfeature.WithHooks(hooks...)
}

// WithHookHints applies provided hook hints.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.WithHookHints,
// instead.
func WithHookHints(hookHints HookHints) Option {
	return openfeature.WithHookHints(hookHints)
}
