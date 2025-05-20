package hooks

import (
	"context"
	"log/slog"

	of "github.com/open-feature/go-sdk/openfeature"
)

const (
	DOMAIN_KEY             = "domain"
	PROVIDER_NAME_KEY      = "provider_name"
	FLAG_KEY_KEY           = "flag_key"
	DEFAULT_VALUE_KEY      = "default_value"
	EVALUATION_CONTEXT_KEY = "evaluation_context"
	ERROR_MESSAGE_KEY      = "error_message"
	REASON_KEY             = "reason"
	VARIANT_KEY            = "variant"
	VALUE_KEY              = "value"
	STAGE_KEY              = "stage"
)

// LoggingHook is a [of.Hook] that logs the flag evaluation lifecycle.
type LoggingHook struct {
	includeEvaluationContext bool
	logger                   *slog.Logger
}

// NewLoggingHook returns a new [LoggingHook] with the default logger.
// To provide a custom logger, use [NewCustomLoggingHook].
func NewLoggingHook(includeEvaluationContext bool) (*LoggingHook, error) {
	return NewCustomLoggingHook(includeEvaluationContext, slog.Default())
}

// NewCustomLoggingHook returns a new [LoggingHook] with the provided logger.
func NewCustomLoggingHook(includeEvaluationContext bool, logger *slog.Logger) (*LoggingHook, error) {
	return &LoggingHook{
		logger:                   logger,
		includeEvaluationContext: includeEvaluationContext,
	}, nil
}

type MarshaledEvaluationContext struct {
	TargetingKey string
	Attributes   map[string]any
}

func (h *LoggingHook) buildArgs(hookContext of.HookContext) []slog.Attr {
	args := []slog.Attr{
		slog.String(DOMAIN_KEY, hookContext.ClientMetadata().Domain()),
		slog.String(PROVIDER_NAME_KEY, hookContext.ProviderMetadata().Name),
		slog.String(FLAG_KEY_KEY, hookContext.FlagKey()),
		slog.Any(DEFAULT_VALUE_KEY, hookContext.DefaultValue()),
	}
	if h.includeEvaluationContext {
		marshaledEvaluationContext := MarshaledEvaluationContext{
			TargetingKey: hookContext.EvaluationContext().TargetingKey(),
			Attributes:   hookContext.EvaluationContext().Attributes(),
		}
		args = append(args, slog.Any(EVALUATION_CONTEXT_KEY, marshaledEvaluationContext))
	}

	return args
}

func (h *LoggingHook) Before(ctx context.Context, hookContext of.HookContext, hookHints of.HookHints) (*of.EvaluationContext, error) {
	args := h.buildArgs(hookContext)
	args = append(args, slog.String(STAGE_KEY, "before"))
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Before stage", args...)
	return nil, nil
}

func (h *LoggingHook) After(ctx context.Context, hookContext of.HookContext,
	flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints,
) error {
	args := h.buildArgs(hookContext)
	args = append(args,
		slog.String(REASON_KEY, string(flagEvaluationDetails.Reason)),
		slog.String(VARIANT_KEY, flagEvaluationDetails.Variant),
		slog.Any(VALUE_KEY, flagEvaluationDetails.Value),
		slog.String(STAGE_KEY, "after"),
	)
	h.logger.LogAttrs(ctx, slog.LevelDebug, "After stage", args...)
	return nil
}

func (h *LoggingHook) Error(ctx context.Context, hookContext of.HookContext, err error, hookHints of.HookHints) {
	args := h.buildArgs(hookContext)
	args = append(args,
		slog.Any(ERROR_MESSAGE_KEY, err),
		slog.String(STAGE_KEY, "error"),
	)
	h.logger.LogAttrs(ctx, slog.LevelError, "Error stage", args...)
}

func (h *LoggingHook) Finally(ctx context.Context, hookContext of.HookContext, flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints) {
}
