package testing

import (
	"context"
	"fmt"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

const (
	Enabled  State = "ENABLED"
	Disabled State = "DISABLED"
)

type InMemoryProvider struct {
	flags map[string]InMemoryFlag
}

func NewInMemoryProvider(from map[string]InMemoryFlag) InMemoryProvider {
	return InMemoryProvider{
		flags: from,
	}
}

func (i InMemoryProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "InMemoryProvider",
	}
}

func (i InMemoryProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(evalCtx)

	var result bool
	res, ok := resolveFlag.(bool)
	if ok {
		result = res
	} else {
		result = defaultValue
		detail.Reason = openfeature.ErrorReason
		detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	}

	return openfeature.BoolResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return openfeature.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(evalCtx)

	var result string
	res, ok := resolveFlag.(string)
	if ok {
		result = res
	} else {
		result = defaultValue
		detail.Reason = openfeature.ErrorReason
		detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	}

	return openfeature.StringResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(evalCtx)

	var result float64
	res, ok := resolveFlag.(float64)
	if ok {
		result = res
	} else {
		result = defaultValue
		detail.Reason = openfeature.ErrorReason
		detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	}

	return openfeature.FloatResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(evalCtx)

	var result int64
	res, ok := resolveFlag.(int)
	if ok {
		result = int64(res)
	} else {
		result = defaultValue
		detail.Reason = openfeature.ErrorReason
		detail.ResolutionError = openfeature.NewTypeMismatchResolutionError("incorrect type association")
	}

	return openfeature.IntResolutionDetail{
		Value:                    result,
		ProviderResolutionDetail: detail,
	}
}

func (i InMemoryProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	memoryFlag, ok := i.flags[flag]
	if !ok {
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("flag for key %s not found", flag)),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	resolveFlag, detail := memoryFlag.Resolve(evalCtx)

	var result interface{}
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
	//TODO implement some hooks
	return []openfeature.Hook{}
}

// Type Definitions for InMemoryProvider flag

// State of the feature flag
type State string

// ContextEvaluator is a callback to perform openfeature.EvaluationContext backed evaluations.
// This is a callback implemented by the flag definer.
type ContextEvaluator *func(this InMemoryFlag, evalCtx openfeature.FlattenedContext) (interface{}, openfeature.ProviderResolutionDetail)

// InMemoryFlag is the feature flag representation accepted by InMemoryProvider
type InMemoryFlag struct {
	Key              string
	State            State
	DefaultVariant   string
	Variants         map[string]interface{}
	ContextEvaluator ContextEvaluator
}

func (flag *InMemoryFlag) Resolve(evalCtx openfeature.FlattenedContext) (
	interface{}, openfeature.ProviderResolutionDetail) {

	// first resolve from context callback
	if flag.ContextEvaluator != nil {
		return (*flag.ContextEvaluator)(*flag, evalCtx)
	}

	// fallback to evaluation

	// check the state
	if flag.State == Disabled {
		return nil, openfeature.ProviderResolutionDetail{
			ResolutionError: openfeature.NewGeneralResolutionError("flag is disabled"),
			Reason:          openfeature.DisabledReason,
		}
	}

	return flag.Variants[flag.DefaultVariant], openfeature.ProviderResolutionDetail{
		Reason:  openfeature.StaticReason,
		Variant: flag.DefaultVariant,
	}
}
