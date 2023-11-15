package openfeature

import (
	openfeature "github.com/open-feature/go-sdk"
)

const (
	// DefaultReason - the resolved value was configured statically, or otherwise fell back to a pre-configured value.
	//
	// Deprecated: use github.com/open-feature/go-sdk.DefaultReason, instead.
	DefaultReason = openfeature.DefaultReason
	// TargetingMatchReason - the resolved value was the result of a dynamic evaluation, such as a rule or specific user-targeting.
	//
	// Deprecated: use github.com/open-feature/go-sdk.TargetingMatchReason, instead.
	TargetingMatchReason = openfeature.TargetingMatchReason
	// SplitReason - the resolved value was the result of pseudorandom assignment.
	//
	// Deprecated: use github.com/open-feature/go-sdk.SplitReason, instead.
	SplitReason = openfeature.SplitReason
	// DisabledReason - the resolved value was the result of the flag being disabled in the management system.
	//
	// Deprecated: use github.com/open-feature/go-sdk.DisabledReason, instead.
	DisabledReason = openfeature.DisabledReason
	// StaticReason - the resolved value is static (no dynamic evaluation)
	//
	// Deprecated: use github.com/open-feature/go-sdk.StaticReason, instead.
	StaticReason = openfeature.StaticReason
	// CachedReason - the resolved value was retrieved from cache
	//
	// Deprecated: use github.com/open-feature/go-sdk.CachedReason, instead.
	CachedReason = openfeature.CachedReason
	// UnknownReason - the reason for the resolved value could not be determined.
	//
	// Deprecated: use github.com/open-feature/go-sdk.UnknownReason, instead.
	UnknownReason = openfeature.UnknownReason
	// ErrorReason - the resolved value was the result of an error.
	//
	// Deprecated: use github.com/open-feature/go-sdk.ErrorReason, instead.
	ErrorReason = openfeature.ErrorReason

	// Deprecated: use github.com/open-feature/go-sdk.NotReadyState, instead.
	NotReadyState = openfeature.NotReadyState
	// Deprecated: use github.com/open-feature/go-sdk.ReadyState, instead.
	ReadyState = openfeature.ReadyState
	// Deprecated: use github.com/open-feature/go-sdk.ErrorState, instead.
	ErrorState = openfeature.ErrorState
	// Deprecated: use github.com/open-feature/go-sdk.StaleState, instead.
	StaleState = openfeature.StaleState

	// Deprecated: use github.com/open-feature/go-sdk.ProviderReady, instead.
	ProviderReady = openfeature.ProviderReady
	// Deprecated: use github.com/open-feature/go-sdk.ProviderConfigChange, instead.
	ProviderConfigChange = openfeature.ProviderConfigChange
	// Deprecated: use github.com/open-feature/go-sdk.ProviderStale, instead.
	ProviderStale = openfeature.ProviderStale
	// Deprecated: use github.com/open-feature/go-sdk.ProviderError, instead.
	ProviderError = openfeature.ProviderError

	// Deprecated: use github.com/open-feature/go-sdk.TargetingKey, instead.
	TargetingKey = openfeature.TargetingKey // evaluation context map key. The targeting key uniquely identifies the subject (end-user, or client service) of a flag evaluation.
)

// FlattenedContext contains metadata for a given flag evaluation in a
// flattened structure. TargetingKey ("targetingKey") is stored as a string
// value if provided in the evaluation context.
//
// Deprecated: use github.com/open-feature/go-sdk.FlattenedContext,
// instead.
type FlattenedContext = openfeature.FlattenedContext

// Reason indicates the semantic reason for a returned flag value
//
// Deprecated: use github.com/open-feature/go-sdk.Reason, instead.
type Reason = openfeature.Reason

// FeatureProvider interface defines a set of functions that can be called in
// order to evaluate a flag. This should be implemented by flag management
// systems.
//
// Deprecated: use github.com/open-feature/go-sdk.FeatureProvider,
// instead.
type FeatureProvider = openfeature.FeatureProvider

// State represents the status of the provider
//
// Deprecated: use github.com/open-feature/go-sdk.State, instead.
type State = openfeature.State

// StateHandler is the contract for initialization & shutdown.
// FeatureProvider can opt in for this behavior by implementing the interface
//
// Deprecated: use github.com/open-feature/go-sdk.StateHandler,
// instead.
type StateHandler = openfeature.StateHandler

// NoopStateHandler is a noop StateHandler implementation
// Status always set to ReadyState to comply with specification
//
// Deprecated: use github.com/open-feature/go-sdk.NoopStateHandler,
// instead.
type NoopStateHandler = openfeature.NoopStateHandler

// EventHandler is the eventing contract enforced for FeatureProvider
//
// Deprecated: use github.com/open-feature/go-sdk.EventHandler,
// instead.
type EventHandler = openfeature.EventHandler

// EventType emitted by a provider implementation
//
// Deprecated: use github.com/open-feature/go-sdk.EventType,
// instead.
type EventType = openfeature.EventType

// ProviderEventDetails is the event payload emitted by FeatureProvider
//
// Deprecated: use
// github.com/open-feature/go-sdk.ProviderEventDetails, instead.
type ProviderEventDetails = openfeature.ProviderEventDetails

// Event is an event emitted by a FeatureProvider.
//
// Deprecated: use github.com/open-feature/go-sdk.Event, instead.
type Event = openfeature.Event

// Deprecated: use github.com/open-feature/go-sdk.EventDetails,
// instead.
type EventDetails = openfeature.EventDetails

// Deprecated: use github.com/open-feature/go-sdk.EventCallback,
// instead.
type EventCallback = openfeature.EventCallback

// NoopEventHandler is the out-of-the-box EventHandler which is noop
//
// Deprecated: use github.com/open-feature/go-sdk.NoopEventHandler,
// instead.
type NoopEventHandler = openfeature.NoopEventHandler

// ProviderResolutionDetail is a structure which contains a subset of the
// fields defined in the EvaluationDetail, representing the result of the
// provider's flag resolution process see
// https://github.com/open-feature/spec/blob/main/specification/types.md#resolution-details
// N.B we could use generics but to support older versions of go for now we
// will have type specific resolution detail
//
// Deprecated: use
// github.com/open-feature/go-sdk.ProviderResolutionDetail,
// instead.
type ProviderResolutionDetail = openfeature.ProviderResolutionDetail

// BoolResolutionDetail provides a resolution detail with boolean type
//
// Deprecated: use
// github.com/open-feature/go-sdk.BoolResolutionDetail, instead.
type BoolResolutionDetail = openfeature.BoolResolutionDetail

// StringResolutionDetail provides a resolution detail with string type
//
// Deprecated: use
// github.com/open-feature/go-sdk.StringResolutionDetail, instead.
type StringResolutionDetail = openfeature.StringResolutionDetail

// FloatResolutionDetail provides a resolution detail with float64 type
//
// Deprecated: use
// github.com/open-feature/go-sdk.FloatResolutionDetail, instead.
type FloatResolutionDetail = openfeature.FloatResolutionDetail

// IntResolutionDetail provides a resolution detail with int64 type
//
// Deprecated: use
// github.com/open-feature/go-sdk.IntResolutionDetail, instead.
type IntResolutionDetail = openfeature.IntResolutionDetail

// InterfaceResolutionDetail provides a resolution detail with interface{} type
//
// Deprecated: use
// github.com/open-feature/go-sdk.InterfaceResolutionDetail,
// instead.
type InterfaceResolutionDetail = openfeature.InterfaceResolutionDetail

// Metadata provides provider name
//
// Deprecated: use github.com/open-feature/go-sdk.Metadata,
// instead.
type Metadata = openfeature.Metadata
