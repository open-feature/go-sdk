package openfeature

import (
	"context"
	"errors"
)

const (
	DISABLED        string = "disabled"     // variant returned because feature is disabled
	TARGETING_MATCH string = "target match" // variant returned because matched target rule
	SPLIT           string = "split"        // variant returned because of pseudorandom assignment
	DEFAULT         string = "default"      // variant returned the default
	UNKNOWN         string = "unknown"      // variant returned for unknown reason
	ERROR           string = "error"        // variant returned due to error

	TargetingKey string = "targetingKey" // evaluation context map key. The targeting key uniquely identifies the subject (end-user, or client service) of a flag evaluation.
)

// FeatureProvider interface defines a set of functions that can be called in order to evaluate a flag.
// vendors should implement
type FeatureProvider interface {
	Metadata() Metadata
	BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx map[string]interface{}) BoolResolutionDetail
	StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx map[string]interface{}) StringResolutionDetail
	FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx map[string]interface{}) FloatResolutionDetail
	IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx map[string]interface{}) IntResolutionDetail
	ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx map[string]interface{}) InterfaceResolutionDetail
	Hooks() []Hook
}

// ResolutionDetail is a structure which contains a subset of the fields defined in the EvaluationDetail,
// representing the result of the provider's flag resolution process
// see https://github.com/open-feature/spec/blob/main/specification/types.md#resolution-details
// N.B we could use generics but to support older versions of go for now we will have type specific resolution
// detail
type ResolutionDetail struct {
	ErrorCode string
	Reason    string
	Variant   string
}

func (resolution ResolutionDetail) Error() error {
	if resolution.ErrorCode == "" {
		return nil
	}
	return errors.New(resolution.ErrorCode)
}

// BoolResolutionDetail provides a resolution detail with boolean type
type BoolResolutionDetail struct {
	Value bool
	ResolutionDetail
}

// StringResolutionDetail provides a resolution detail with string type
type StringResolutionDetail struct {
	Value string
	ResolutionDetail
}

// FloatResolutionDetail provides a resolution detail with float64 type
type FloatResolutionDetail struct {
	Value float64
	ResolutionDetail
}

// IntResolutionDetail provides a resolution detail with int64 type
type IntResolutionDetail struct {
	Value int64
	ResolutionDetail
}

// InterfaceResolutionDetail provides a resolution detail with interface{} type
type InterfaceResolutionDetail struct {
	Value interface{}
	ResolutionDetail
}

// Metadata provides provider name
type Metadata struct {
	Name string
}
