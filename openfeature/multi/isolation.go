package multi

import (
	"context"
	"fmt"
	"sync"

	of "github.com/open-feature/go-sdk/openfeature"
)

type (
	// hookIsolator is used as a wrapper around a provider that prevents context changes from leaking between providers
	// during evaluation
	hookIsolator struct {
		of.UnimplementedHook
		mu sync.Mutex
		of.FeatureProvider
		hooks           []of.Hook
		capturedContext of.HookContext
		capturedHints   of.HookHints
		name            string
	}

	// eventHandlingHookIsolator is equivalent to hookIsolator, but also implements [of.EventHandler]
	eventHandlingHookIsolator struct {
		hookIsolator
	}
)

// Compile-time interface compliance checks
var (
	_ namedProvider      = (*hookIsolator)(nil)
	_ of.FeatureProvider = (*hookIsolator)(nil)
	_ of.Hook            = (*hookIsolator)(nil)
	_ of.EventHandler    = (*eventHandlingHookIsolator)(nil)
)

// isolateProvider wraps a [of.FeatureProvider] to execute its hooks along with any additional ones.
func isolateProvider(provider namedProvider, extraHooks []of.Hook) *hookIsolator {
	return &hookIsolator{
		FeatureProvider: provider,
		hooks:           append(provider.Hooks(), extraHooks...),
		name:            provider.Name(),
	}
}

// isolateProviderWithEvents wraps a [of.FeatureProvider] to execute its hooks along with any additional ones. This is
// identical to [isolateProvider], but also this will also implement [of.EventHandler].
func isolateProviderWithEvents(provider namedProvider, extraHooks []of.Hook) *eventHandlingHookIsolator {
	return &eventHandlingHookIsolator{*isolateProvider(provider, extraHooks)}
}

func (h *eventHandlingHookIsolator) EventChannel() <-chan of.Event {
	return h.FeatureProvider.(of.EventHandler).EventChannel()
}

func (h *hookIsolator) Name() string {
	return h.name
}

func (h *hookIsolator) unwrap() of.FeatureProvider {
	return h.FeatureProvider
}

func (h *hookIsolator) Before(_ context.Context, hookContext of.HookContext, hookHints of.HookHints) (*of.EvaluationContext, error) {
	// Used for capturing the context and hints
	h.mu.Lock()
	defer h.mu.Unlock()
	h.capturedContext = hookContext
	h.capturedHints = hookHints
	// Return copy of original evaluation context so any changes are isolated to each provider's hooks
	evalCtx := h.capturedContext.EvaluationContext()
	return &evalCtx, nil
}

func (h *hookIsolator) Metadata() of.Metadata {
	return h.FeatureProvider.Metadata()
}

func (h *hookIsolator) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx of.FlattenedContext) of.BoolResolutionDetail {
	completeEval := h.evaluate(ctx, flag, of.Boolean, defaultValue, flatCtx)

	return of.BoolResolutionDetail{
		Value:                    completeEval.Value.(bool),
		ProviderResolutionDetail: toProviderResolutionDetail(completeEval),
	}
}

func (h *hookIsolator) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx of.FlattenedContext) of.StringResolutionDetail {
	completeEval := h.evaluate(ctx, flag, of.String, defaultValue, flatCtx)

	return of.StringResolutionDetail{
		Value:                    completeEval.Value.(string),
		ProviderResolutionDetail: toProviderResolutionDetail(completeEval),
	}
}

func (h *hookIsolator) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx of.FlattenedContext) of.FloatResolutionDetail {
	completeEval := h.evaluate(ctx, flag, of.Float, defaultValue, flatCtx)

	return of.FloatResolutionDetail{
		Value:                    completeEval.Value.(float64),
		ProviderResolutionDetail: toProviderResolutionDetail(completeEval),
	}
}

func (h *hookIsolator) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx of.FlattenedContext) of.IntResolutionDetail {
	completeEval := h.evaluate(ctx, flag, of.Int, defaultValue, flatCtx)

	return of.IntResolutionDetail{
		Value:                    completeEval.Value.(int64),
		ProviderResolutionDetail: toProviderResolutionDetail(completeEval),
	}
}

func (h *hookIsolator) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	completeEval := h.evaluate(ctx, flag, of.Object, defaultValue, flatCtx)

	return of.InterfaceResolutionDetail{
		Value:                    completeEval.Value,
		ProviderResolutionDetail: toProviderResolutionDetail(completeEval),
	}
}

// toProviderResolutionDetail Converts a [of.InterfaceEvaluationDetails] to a [of.ProviderResolutionDetail].
func toProviderResolutionDetail(evalDetails of.InterfaceEvaluationDetails) of.ProviderResolutionDetail {
	var resolutionErr of.ResolutionError
	var reason of.Reason
	switch evalDetails.ErrorCode {
	case of.GeneralCode:
		resolutionErr = of.NewGeneralResolutionError(evalDetails.ErrorMessage)
		reason = of.ErrorReason
	case of.FlagNotFoundCode:
		resolutionErr = of.NewFlagNotFoundResolutionError(evalDetails.ErrorMessage)
		reason = of.DefaultReason
	case of.TargetingKeyMissingCode:
		resolutionErr = of.NewTargetingKeyMissingResolutionError(evalDetails.ErrorMessage)
		reason = of.TargetingMatchReason
	case of.TypeMismatchCode:
		resolutionErr = of.NewTypeMismatchResolutionError(evalDetails.ErrorMessage)
		reason = of.ErrorReason
	case of.ParseErrorCode:
		resolutionErr = of.NewParseErrorResolutionError(evalDetails.ErrorMessage)
		reason = of.ErrorReason
	case of.InvalidContextCode:
		resolutionErr = of.NewInvalidContextResolutionError(evalDetails.ErrorMessage)
		reason = of.ErrorReason
	}
	return of.ProviderResolutionDetail{
		ResolutionError: resolutionErr,
		Reason:          reason,
		Variant:         evalDetails.Variant,
		FlagMetadata:    evalDetails.FlagMetadata,
	}
}

func (h *hookIsolator) Hooks() []of.Hook {
	// return self as hook to capture contexts
	return []of.Hook{h}
}

// evaluate Executes evaluation of the flag wrapped by executing hooks.
func (h *hookIsolator) evaluate(ctx context.Context, flag string, flagType of.Type, defaultValue any, flatCtx of.FlattenedContext) of.InterfaceEvaluationDetails {
	evalDetails := of.InterfaceEvaluationDetails{
		Value: defaultValue,
		EvaluationDetails: of.EvaluationDetails{
			FlagKey:  flag,
			FlagType: flagType,
		},
	}

	defer func() {
		h.finallyHooks(ctx, evalDetails)
	}()

	evalCtx, err := h.beforeHooks(ctx)
	// Update hook context unconditionally
	h.updateEvalContext(evalCtx)
	if err != nil {
		err = fmt.Errorf("before hook: %w", err)
		h.errorHooks(ctx, err)
		evalDetails.ResolutionDetail = of.ResolutionDetail{
			Reason:       of.ErrorReason,
			ErrorCode:    of.GeneralCode,
			ErrorMessage: err.Error(),
			FlagMetadata: nil,
		}
		return evalDetails
	}

	// Merge together the passed in flat context and the captured evaluation context and transform back into a flattened
	// context for evaluation
	flatCtx = flattenContext(mergeContexts(h.capturedContext.EvaluationContext(), deepenContext(flatCtx)))

	var resolution of.InterfaceResolutionDetail
	switch flagType {
	case of.Object:
		resolution = h.FeatureProvider.ObjectEvaluation(ctx, flag, defaultValue, flatCtx)
	case of.Boolean:
		defValue := defaultValue.(bool)
		res := h.FeatureProvider.BooleanEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.String:
		defValue := defaultValue.(string)
		res := h.FeatureProvider.StringEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.Float:
		defValue := defaultValue.(float64)
		res := h.FeatureProvider.FloatEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case of.Int:
		defValue := defaultValue.(int64)
		res := h.FeatureProvider.IntEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	}

	err = resolution.Error()
	if err != nil {
		err = fmt.Errorf("error code: %w", err)
		h.errorHooks(ctx, err)
		evalDetails.ResolutionDetail = resolution.ResolutionDetail()
		evalDetails.Reason = of.ErrorReason
		return evalDetails
	}
	evalDetails.Value = resolution.Value
	evalDetails.ResolutionDetail = resolution.ResolutionDetail()

	if err := h.afterHooks(ctx, evalDetails); err != nil {
		err = fmt.Errorf("after hook: %w", err)
		h.errorHooks(ctx, err)
		return evalDetails
	}

	return evalDetails
}

// beforeHooks Executes all before hooks together, merging the changes to the [of.EvaluationContext] as it goes. The
// return of this function is a merged version of the evaluation context
func (h *hookIsolator) beforeHooks(ctx context.Context) (of.EvaluationContext, error) {
	contexts := []of.EvaluationContext{h.capturedContext.EvaluationContext()}
	for _, hook := range h.hooks {
		resultEvalCtx, err := hook.Before(ctx, h.capturedContext, h.capturedHints)
		if resultEvalCtx != nil {
			contexts = append(contexts, *resultEvalCtx)
		}
		if err != nil {
			return mergeContexts(contexts...), err
		}
	}

	return mergeContexts(contexts...), nil
}

// afterHooks executes all after [of.Hook] instances together.
func (h *hookIsolator) afterHooks(ctx context.Context, evalDetails of.InterfaceEvaluationDetails) error {
	for _, hook := range h.hooks {
		if err := hook.After(ctx, h.capturedContext, evalDetails, h.capturedHints); err != nil {
			return err
		}
	}

	return nil
}

// errorHooks executes all error [of.Hook] instances together.
func (h *hookIsolator) errorHooks(ctx context.Context, err error) {
	for _, hook := range h.hooks {
		hook.Error(ctx, h.capturedContext, err, h.capturedHints)
	}
}

// finallyHooks execute all finally [of.Hook] instances together.
func (h *hookIsolator) finallyHooks(ctx context.Context, details of.InterfaceEvaluationDetails) {
	for _, hook := range h.hooks {
		hook.Finally(ctx, h.capturedContext, details, h.capturedHints)
	}
}

// updateEvalContext returns a new [of.HookContext] with an updated [of.EvaluationContext] value. [of.HookContext] is
// immutable, and this returns a new [of.HookContext] with all other values copied
func (h *hookIsolator) updateEvalContext(evalCtx of.EvaluationContext) {
	hookCtx := of.NewHookContext(
		h.capturedContext.FlagKey(),
		h.capturedContext.FlagType(),
		h.capturedContext.DefaultValue(),
		h.capturedContext.ClientMetadata(),
		h.capturedContext.ProviderMetadata(),
		evalCtx,
	)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.capturedContext = hookCtx
}

// deepenContext converts a [of.FlattenedContext] to a [of.EvaluationContext].
func deepenContext(flatCtx of.FlattenedContext) of.EvaluationContext {
	noTargetingKey := make(map[string]any)
	for k, v := range flatCtx {
		if k != of.TargetingKey {
			noTargetingKey[k] = v
		}
	}
	var targetingKey string
	if tk, ok := flatCtx[of.TargetingKey]; ok {
		targetingKey, _ = tk.(string)
	}
	return of.NewEvaluationContext(targetingKey, noTargetingKey)
}

// flattenContext converts a [of.EvaluationContext] to a [of.FlattenedContext]
func flattenContext(evalCtx of.EvaluationContext) of.FlattenedContext {
	flatCtx := evalCtx.Attributes()
	flatCtx[of.TargetingKey] = evalCtx.TargetingKey()
	return flatCtx
}

// mergeContexts merges attributes from the given EvaluationContexts with the nth [of.EvaluationContext] taking precedence
// in case of any conflicts with the (n+1)th [of.EvaluationContext].
func mergeContexts(evaluationContexts ...of.EvaluationContext) of.EvaluationContext {
	if len(evaluationContexts) == 0 {
		return of.EvaluationContext{}
	}
	// create initial values
	attributes := evaluationContexts[0].Attributes()
	targetingKey := evaluationContexts[0].TargetingKey()

	for i := 1; i < len(evaluationContexts); i++ {
		if targetingKey == "" && evaluationContexts[i].TargetingKey() != "" {
			targetingKey = evaluationContexts[i].TargetingKey()
		}

		for k, v := range evaluationContexts[i].Attributes() {
			_, ok := attributes[k]
			if !ok {
				attributes[k] = v
			}
		}
	}

	return of.NewEvaluationContext(targetingKey, attributes)
}
