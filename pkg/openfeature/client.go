package openfeature

import (
	openfeature "github.com/open-feature/go-sdk"
)

// IClient defines the behaviour required of an openfeature client
//
// Deprecated: use github.com/open-feature/go-sdk.IClient, instead.
type IClient = openfeature.IClient

// ClientMetadata provides a client's metadata
//
// Deprecated: use github.com/open-feature/go-sdk.ClientMetadata,
// instead.
type ClientMetadata = openfeature.ClientMetadata

// NewClientMetadata constructs ClientMetadata
// Allows for simplified hook test cases while maintaining immutability
//
// Deprecated: use
// github.com/open-feature/go-sdk.NewClientMetadata, instead.
func NewClientMetadata(name string) ClientMetadata {
	return openfeature.NewClientMetadata(name)
}

// Client implements the behaviour required of an openfeature client
//
// Deprecated: use github.com/open-feature/go-sdk.Client, instead.
type Client = openfeature.Client

// NewClient returns a new Client. Name is a unique identifier for this client
//
// Deprecated: use github.com/open-feature/go-sdk.NewClient,
// instead.
func NewClient(name string) *Client {
	return openfeature.NewClient(name)
}

// Type represents the type of a flag
//
// Deprecated: use github.com/open-feature/go-sdk.Type, instead.
type Type = openfeature.Type

const (
	// Deprecated: use github.com/open-feature/go-sdk.Boolean,
	// instead.
	Boolean = openfeature.Boolean
	// Deprecated: use github.com/open-feature/go-sdk.String,
	// instead.
	String = openfeature.String
	// Deprecated: use github.com/open-feature/go-sdk.Float,
	// instead.
	Float = openfeature.Float
	// Deprecated: use github.com/open-feature/go-sdk.Int,
	// instead.
	Int = openfeature.Int
	// Deprecated: use github.com/open-feature/go-sdk.Object,
	// instead.
	Object = openfeature.Object
)

// Deprecated: use
// github.com/open-feature/go-sdk.EvaluationDetails, instead.
type EvaluationDetails = openfeature.EvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk.BooleanEvaluationDetails,
// instead.
type BooleanEvaluationDetails = openfeature.BooleanEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk.StringEvaluationDetails, instead.
type StringEvaluationDetails = openfeature.StringEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk.FloatEvaluationDetails, instead.
type FloatEvaluationDetails = openfeature.FloatEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk.IntEvaluationDetails, instead.
type IntEvaluationDetails = openfeature.IntEvaluationDetails

// Deprecated: use
// github.com/open-feature/go-sdk.InterfaceEvaluationDetails,
// instead.
type InterfaceEvaluationDetails = openfeature.InterfaceEvaluationDetails

// Deprecated: use github.com/open-feature/go-sdk.ResolutionDetail, instead.
type ResolutionDetail = openfeature.ResolutionDetail

// FlagMetadata is a structure which supports definition of arbitrary properties, with keys of type string, and values
// of type boolean, string, int64 or float64. This structure is populated by a provider for use by an Application
// Author (via the Evaluation API) or an Application Integrator (via hooks).
//
// Deprecated: use github.com/open-feature/go-sdk.FlagMetadata,
// instead.
type FlagMetadata = openfeature.FlagMetadata

// Option applies a change to EvaluationOptions
//
// Deprecated: use github.com/open-feature/go-sdk.Option, instead.
type Option = openfeature.Option

// EvaluationOptions should contain a list of hooks to be executed for a flag evaluation
//
// Deprecated: use
// github.com/open-feature/go-sdk.EvaluationOptions, instead.
type EvaluationOptions = openfeature.EvaluationOptions

// WithHooks applies provided hooks.
//
// Deprecated: use github.com/open-feature/go-sdk.WithHooks,
// instead.
func WithHooks(hooks ...Hook) Option {
	return openfeature.WithHooks(hooks...)
}

// WithHookHints applies provided hook hints.
//
// Deprecated: use github.com/open-feature/go-sdk.WithHookHints,
// instead.
func WithHookHints(hookHints HookHints) Option {
	return openfeature.WithHookHints(hookHints)
}
