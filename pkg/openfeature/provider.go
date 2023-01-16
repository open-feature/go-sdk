package openfeature

import (
	"context"
	"errors"
)

const (
	// DefaultReason - the resolved value was configured statically, or otherwise fell back to a pre-configured value.
	DefaultReason Reason = "DEFAULT"
	// TargetingMatchReason - the resolved value was the result of a dynamic evaluation, such as a rule or specific user-targeting.
	TargetingMatchReason Reason = "TARGETING_MATCH"
	// SplitReason - the resolved value was the result of pseudorandom assignment.
	SplitReason Reason = "SPLIT"
	// DisabledReason - the resolved value was the result of the flag being disabled in the management system.
	DisabledReason Reason = "DISABLED"
	// StaticReason - the resolved value is static (no dynamic evaluation)
	StaticReason Reason = "STATIC"
	// CachedReason - the resolved value was retrieved from cache
	CachedReason Reason = "CACHED"
	// UnknownReason - the reason for the resolved value could not be determined.	
	UnknownReason Reason = "UNKNOWN"
	// ErrorReason - the resolved value was the result of an error.
	ErrorReason Reason = "ERROR"

	TargetingKey string = "targetingKey" // evaluation context map key. The targeting key uniquely identifies the subject (end-user, or client service) of a flag evaluation.
)

// FlattenedContext contains metadata for a given flag evaluation in a flattened structure.
// TargetingKey ("targetingKey") is stored as a string value if provided in the evaluation context.
type FlattenedContext map[string]interface{}

// Reason indicates the semantic reason for a returned flag value
type Reason string

// FeatureProvider interface defines a set of functions that can be called in order to evaluate a flag.
// vendors should implement
type FeatureProvider interface {
	Metadata() Metadata
	BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx FlattenedContext) BoolResolutionDetail
	StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx FlattenedContext) StringResolutionDetail
	FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx FlattenedContext) FloatResolutionDetail
	IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx FlattenedContext) IntResolutionDetail
	ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx FlattenedContext) InterfaceResolutionDetail
	Hooks() []Hook
}

// ProviderResolutionDetail is a structure which contains a subset of the fields defined in the EvaluationDetail,
// representing the result of the provider's flag resolution process
// see https://github.com/open-feature/spec/blob/main/specification/types.md#resolution-details
// N.B we could use generics but to support older versions of go for now we will have type specific resolution
// detail
type ProviderResolutionDetail struct {
	ResolutionError ResolutionError
	Reason          Reason
	Variant         string
}

func (p ProviderResolutionDetail) ResolutionDetail() ResolutionDetail {
	return ResolutionDetail{
		Variant:      p.Variant,
		Reason:       p.Reason,
		ErrorCode:    p.ResolutionError.code,
		ErrorMessage: p.ResolutionError.message,
	}
}

func (p ProviderResolutionDetail) Error() error {
	if p.ResolutionError.code == "" {
		return nil
	}
	return errors.New(p.ResolutionError.Error())
}

// BoolResolutionDetail provides a resolution detail with boolean type
type BoolResolutionDetail struct {
	Value bool
	ProviderResolutionDetail
}

// StringResolutionDetail provides a resolution detail with string type
type StringResolutionDetail struct {
	Value string
	ProviderResolutionDetail
}

// FloatResolutionDetail provides a resolution detail with float64 type
type FloatResolutionDetail struct {
	Value float64
	ProviderResolutionDetail
}

// IntResolutionDetail provides a resolution detail with int64 type
type IntResolutionDetail struct {
	Value int64
	ProviderResolutionDetail
}

// InterfaceResolutionDetail provides a resolution detail with interface{} type
type InterfaceResolutionDetail struct {
	Value interface{}
	ProviderResolutionDetail
}

// Metadata provides provider name
type Metadata struct {
	Name string
}
