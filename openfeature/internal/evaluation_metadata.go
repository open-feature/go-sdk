package internal

import "context"

type evaluationMetadataKey struct{}

// EvaluationMetadata carries immutable hook metadata for one client evaluation.
// HookHints holds openfeature.HookHints as any to avoid an import cycle.
type EvaluationMetadata struct {
	ClientDomain string
	ProviderName string
	HookHints    any
}

// WithEvaluationMetadata creates a fresh scope, including for nested evaluations.
func WithEvaluationMetadata(ctx context.Context, metadata EvaluationMetadata) context.Context {
	return context.WithValue(ctx, evaluationMetadataKey{}, metadata)
}

// EvaluationMetadataFromContext returns metadata when evaluation started through a client.
func EvaluationMetadataFromContext(ctx context.Context) (EvaluationMetadata, bool) {
	metadata, ok := ctx.Value(evaluationMetadataKey{}).(EvaluationMetadata)
	return metadata, ok
}
