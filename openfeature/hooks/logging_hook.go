package hooks

import (
	"context"

	"github.com/open-feature/go-sdk/openfeature"
	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/exp/slog"
)

const (
	DOMAIN_KEY             = "domain"
	PROVIDER_NAME_KEY      = "provider_name"
	FLAG_KEY_KEY           = "flag_key"
	DEFAULT_VALUE_KEY      = "default_value"
	EVALUATION_CONTEXT_KEY = "evaluation_context"
	ERROR_CODE_KEY         = "error_code"
	ERROR_MESSAGE_KEY      = "error_message"
	REASON_KEY             = "reason"
	VARIANT_KEY            = "variant"
	VALUE_KEY              = "value"
)

type LoggingHook struct {
	includeEvaluationContext bool
	logger                   *slog.Logger
}

func NewLoggingHook(logger *slog.Logger, includeEvaluationContext bool) (*LoggingHook, error) {
	return &LoggingHook{
		logger:                   logger,
		includeEvaluationContext: includeEvaluationContext,
	}, nil
}

func (l LoggingHook) buildArgs(hookContext openfeature.HookContext) []interface{} {

	args := []interface{}{
		DOMAIN_KEY, hookContext.ClientMetadata().Domain(),
		PROVIDER_NAME_KEY, hookContext.ProviderMetadata().Name,
		FLAG_KEY_KEY, hookContext.FlagKey(),
		DEFAULT_VALUE_KEY, hookContext.DefaultValue(),
	}
	if l.includeEvaluationContext {
		args = append(args, EVALUATION_CONTEXT_KEY, hookContext.EvaluationContext())
	}

	return args
}

func (h *LoggingHook) Before(ctx context.Context, hookContext openfeature.HookContext,
	hint openfeature.HookHints) (*openfeature.EvaluationContext, error) {
	var args = h.buildArgs(hookContext)
	h.logger.Debug("Before stage", args...)
	return nil, nil
}

func (h *LoggingHook) After(ctx context.Context, hookContext of.HookContext,
	flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints) error {
	args := h.buildArgs(hookContext)
	args = append(args, REASON_KEY, flagEvaluationDetails.Reason)
	args = append(args, VARIANT_KEY, flagEvaluationDetails.Variant)
	args = append(args, VALUE_KEY, flagEvaluationDetails.Value)
	h.logger.Debug("After stage", args...)
	return nil
}

func (h *LoggingHook) Error(ctx context.Context, hookContext openfeature.HookContext, err error, hint openfeature.HookHints) {
	args := h.buildArgs(hookContext)
	args = append(args, ERROR_CODE_KEY, err) // TODO ??
	h.logger.Error("Error stage", args...)
}

func (h *LoggingHook) Finally(ctx context.Context, hCtx openfeature.HookContext, hint openfeature.HookHints) {

}
