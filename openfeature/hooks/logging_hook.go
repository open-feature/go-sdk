package hooks

import (
	"context"
	"log/slog"

	of "github.com/open-feature/go-sdk/openfeature"
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

// NewLoggingHook returns a new [LoggingHook] with the default logger.
// To provide a custom logger, use [NewCustomLoggingHook].
func NewLoggingHook(includeEvaluationContext bool) *LoggingHook {
	return NewCustomLoggingHook(includeEvaluationContext, slog.Default())
}

// NewCustomLoggingHook returns a new [LoggingHook] with the provided logger.
func NewCustomLoggingHook(includeEvaluationContext bool, logger *slog.Logger) *LoggingHook {
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

func (h *LoggingHook) Before(ctx context.Context, hookContext of.HookContext, hookHints of.HookHints) (*of.EvaluationContext, error) {
	args := h.buildArgs(hookContext)
	args = append(args, slog.String(stageKey, "before"))
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Before stage", args...)
	return nil, nil
}

func (h *LoggingHook) After(ctx context.Context, hookContext of.HookContext,
	flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints,
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

func (h *LoggingHook) Finally(ctx context.Context, hookContext of.HookContext, flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints) {
}
