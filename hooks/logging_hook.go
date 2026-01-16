package hooks

import (
	"context"
	"log/slog"

	of "go.openfeature.dev/openfeature/v2"
)

const (
	domainKey            = "domain"
	providerNameKey      = "provider_name"
	flagKeyKey           = "flag_key"
	defaultValueKey      = "default_value"
	evaluationContextKey = "evaluation_context"
	targetingKeyKey      = "targeting_key"
	attributesKey        = "attributes"
	errorMessageKey      = "error_message"
	reasonKey            = "reason"
	variantKey           = "variant"
	valueKey             = "value"
	stageKey             = "stage"
)

// LoggingHook is a [of.Hook] that logs the flag evaluation lifecycle.
type LoggingHook struct {
	includeEvaluationContext bool
	logger                   *slog.Logger
}

var _ of.Hook = (*LoggingHook)(nil)

// NewLoggingHook returns a new [LoggingHook] with the provided logger.
func NewLoggingHook(includeEvaluationContext bool, logger *slog.Logger) *LoggingHook {
	return &LoggingHook{
		logger:                   logger,
		includeEvaluationContext: includeEvaluationContext,
	}
}

func (h *LoggingHook) buildArgs(hookContext of.HookContext) []slog.Attr {
	args := []slog.Attr{
		slog.String(domainKey, hookContext.ClientMetadata().Domain()),
		slog.String(providerNameKey, hookContext.ProviderMetadata().Name),
		slog.String(flagKeyKey, hookContext.FlagKey()),
		slog.Any(defaultValueKey, hookContext.DefaultValue()),
	}
	if h.includeEvaluationContext {
		args = append(args,
			slog.Group(evaluationContextKey,
				slog.String(targetingKeyKey, hookContext.EvaluationContext().TargetingKey()),
				slog.Any(attributesKey, hookContext.EvaluationContext().Attributes()),
			))
	}

	return args
}

func (h *LoggingHook) Before(ctx context.Context, hookContext of.HookContext, hookHints of.HookHints) (context.Context, error) {
	args := h.buildArgs(hookContext)
	args = append(args, slog.String(stageKey, "before"))
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Before stage", args...)
	return ctx, nil
}

func (h *LoggingHook) After(ctx context.Context, hookContext of.HookContext,
	flagEvaluationDetails of.EvaluationDetails[of.FlagTypes], hookHints of.HookHints,
) error {
	args := h.buildArgs(hookContext)
	args = append(args,
		slog.String(reasonKey, string(flagEvaluationDetails.Reason)),
		slog.String(variantKey, flagEvaluationDetails.Variant),
		slog.Any(valueKey, flagEvaluationDetails.Value),
		slog.String(stageKey, "after"),
	)
	h.logger.LogAttrs(ctx, slog.LevelDebug, "After stage", args...)
	return nil
}

func (h *LoggingHook) Error(ctx context.Context, hookContext of.HookContext, err error, hookHints of.HookHints) {
	args := h.buildArgs(hookContext)
	args = append(args,
		slog.Any(errorMessageKey, err),
		slog.String(stageKey, "error"),
	)
	h.logger.LogAttrs(ctx, slog.LevelError, "Error stage", args...)
}

func (h *LoggingHook) Finally(ctx context.Context, hookContext of.HookContext, flagEvaluationDetails of.EvaluationDetails[of.FlagTypes], hookHints of.HookHints) {
}
