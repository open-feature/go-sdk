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

func NewLoggingHook(includeEvaluationContext bool) (*LoggingHook, error) {
	return NewCustomLoggingHook(slog.Default(), includeEvaluationContext)
}

func NewCustomLoggingHook(logger *slog.Logger, includeEvaluationContext bool) (*LoggingHook, error) {
	return &LoggingHook{
		logger:                   logger,
		includeEvaluationContext: includeEvaluationContext,
	}, nil
}

type MarshaledEvaluationContext struct {
	TargetingKey string
	Attributes   map[string]interface{}
}

func (l LoggingHook) buildArgs(hookContext of.HookContext) ([]interface{}, error) {

	args := []interface{}{
		DOMAIN_KEY, hookContext.ClientMetadata().Domain(),
		PROVIDER_NAME_KEY, hookContext.ProviderMetadata().Name,
		FLAG_KEY_KEY, hookContext.FlagKey(),
		DEFAULT_VALUE_KEY, hookContext.DefaultValue(),
	}
	if l.includeEvaluationContext {
		marshaledEvaluationContext := MarshaledEvaluationContext{
			TargetingKey: hookContext.EvaluationContext().TargetingKey(),
			Attributes:   hookContext.EvaluationContext().Attributes(),
		}
		args = append(args, EVALUATION_CONTEXT_KEY, marshaledEvaluationContext)
	}

	return args, nil
}

func (h *LoggingHook) Before(ctx context.Context, hookContext of.HookContext,
	hint of.HookHints) (*of.EvaluationContext, error) {
	var args, err = h.buildArgs(hookContext)
	if err != nil {
		return nil, err
	}
	h.logger.Debug("Before stage", args...)
	return nil, nil
}

func (h *LoggingHook) After(ctx context.Context, hookContext of.HookContext,
	flagEvaluationDetails of.InterfaceEvaluationDetails, hookHints of.HookHints) error {
	var args, err = h.buildArgs(hookContext)
	if err != nil {
		return err
	}
	args = append(args, REASON_KEY, flagEvaluationDetails.Reason)
	args = append(args, VARIANT_KEY, flagEvaluationDetails.Variant)
	args = append(args, VALUE_KEY, flagEvaluationDetails.Value)
	h.logger.Debug("After stage", args...)
	return nil
}

func (h *LoggingHook) Error(ctx context.Context, hookContext of.HookContext, err error, hint of.HookHints) {
	args, buildArgsErr := h.buildArgs(hookContext)
	if buildArgsErr != nil {
		slog.Error("Error building args", "error", buildArgsErr)
	}
	args = append(args, ERROR_CODE_KEY, err)
	h.logger.Error("Error stage", args...)
}

func (h *LoggingHook) Finally(ctx context.Context, hCtx of.HookContext, hint of.HookHints) {

}
