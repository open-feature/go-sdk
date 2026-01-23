// Package memprovider provides an in-memory feature flag provider for OpenFeature.
package memprovider

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/openfeature"
)

const (
	Enabled  State = "ENABLED"
	Disabled State = "DISABLED"
)

type InMemoryProvider struct {
	flags          map[string]InMemoryFlag
	trackingEvents map[string][]InMemoryEvent
}

func NewInMemoryProvider(from map[string]InMemoryFlag) InMemoryProvider {
	return InMemoryProvider{
		flags:          from,
		trackingEvents: map[string][]InMemoryEvent{},
	}
}

func (i InMemoryProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "InMemoryProvider",
	}
}

func (i InMemoryProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	memoryFlag, details, ok := i.find(flag)
	if !ok {
		return openfeature.BoolResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *details,
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(defaultValue, flatCtx)
	result := genericResolve[bool](resolveFlag, defaultValue, &detail)

	return openfeature.BoolResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	memoryFlag, details, ok := i.find(flag)
	if !ok {
		return openfeature.StringResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *details,
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(defaultValue, flatCtx)
	result := genericResolve[string](resolveFlag, defaultValue, &detail)

	return openfeature.StringResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	memoryFlag, details, ok := i.find(flag)
	if !ok {
		return openfeature.FloatResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *details,
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(defaultValue, flatCtx)
	result := genericResolve[float64](resolveFlag, defaultValue, &detail)

	return openfeature.FloatResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	memoryFlag, details, ok := i.find(flag)
	if !ok {
		return openfeature.IntResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *details,
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(defaultValue, flatCtx)
	result := genericResolve[int64](resolveFlag, defaultValue, &detail)

	return openfeature.IntResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	memoryFlag, details, ok := i.find(flag)
	if !ok {
		return openfeature.InterfaceResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: *details,
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(defaultValue, flatCtx)

	var result any
	if resolveFlag != nil {
		result = resolveFlag
	} else {
		result = defaultValue
		detail.Reason = openfeature.ErrorReason
		detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	}

	return openfeature.InterfaceResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) Hooks() []openfeature.Hook {
	return []openfeature.Hook{}
}

func (i InMemoryProvider) Track(ctx context.Context, trackingEventName string, evalCtx openfeature.EvaluationContext, details openfeature.TrackingEventDetails) {
	i.trackingEvents[trackingEventName] = append(i.trackingEvents[trackingEventName], InMemoryEvent{
		Value:             details.Value(),
		Data:              details.Attributes(),
		ContextAttributes: evalCtx.Attributes(),
	})
}

func (i InMemoryProvider) find(flag string) (*InMemoryFlag, *openfeature.ProviderResolutionDetail, bool) {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return nil,
			&openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			}, false
	}

	return &memoryFlag, nil, true
}

// helpers

// genericResolve is a helper to extract type verified evaluation and fill openfeature.ProviderResolutionDetail.
// It coerces smaller numeric types to their canonical forms (int* -> int64, float32 -> float64)
// to provide a more forgiving API for test flag configuration.
//
// Note: Only signed integer types are supported for conversion. Unsigned integer types
// (uint, uint8, uint16, uint32, uint64) will result in type mismatch errors.
func genericResolve[T comparable](value any, defaultValue T, detail *openfeature.ProviderResolutionDetail) T {
	// Try direct type assertion first
	if v, ok := value.(T); ok {
		return v
	}

	// Handle type conversions based on target type
	switch any(defaultValue).(type) {
	case int64:
		// Convert various int types to int64
		switch v := value.(type) {
		case int8:
			return any(int64(v)).(T)
		case int16:
			return any(int64(v)).(T)
		case int32:
			return any(int64(v)).(T)
		case int:
			return any(int64(v)).(T)
		}
	case float64:
		// Convert float32 to float64 and int types to float64
		switch v := value.(type) {
		case float32:
			return any(float64(v)).(T)
		}
	}

	// If no conversion worked, return error
	detail.Reason = openfeature.ErrorReason
	detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	return defaultValue
}

// Type Definitions for InMemoryProvider flag

// State of the feature flag
type State string

// ContextEvaluator is a callback to perform openfeature.EvaluationContext backed evaluations.
// This is a callback implemented by the flag definer.
type ContextEvaluator *func(this InMemoryFlag, flatCtx openfeature.FlattenedContext) (any, openfeature.ProviderResolutionDetail)

// InMemoryFlag is the feature flag representation accepted by InMemoryProvider
type InMemoryFlag struct {
	Key              string
	State            State
	DefaultVariant   string
	Variants         map[string]any
	ContextEvaluator ContextEvaluator
}

func (flag *InMemoryFlag) Resolve(defaultValue any, flatCtx openfeature.FlattenedContext) (
	any, openfeature.ProviderResolutionDetail,
) {
	// check the state
	if flag.State == Disabled {
		return defaultValue, openfeature.ProviderResolutionDetail{
			ResolutionError: openfeature.NewGeneralResolutionError("flag is disabled"),
			Reason:          openfeature.DisabledReason,
		}
	}

	// first resolve from context callback
	if flag.ContextEvaluator != nil {
		return (*flag.ContextEvaluator)(*flag, flatCtx)
	}

	// fallback to evaluation

	return flag.Variants[flag.DefaultVariant], openfeature.ProviderResolutionDetail{
		Reason:  openfeature.StaticReason,
		Variant: flag.DefaultVariant,
	}
}

type InMemoryEvent struct {
	Value             float64
	Data              map[string]any
	ContextAttributes map[string]any
}
