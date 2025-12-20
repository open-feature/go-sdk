package openfeature

import (
	"context"
	"maps"
)

// EvaluationContext provides ambient information for the purposes of flag evaluation
// The use of the constructor, NewEvaluationContext, is enforced to set EvaluationContext's fields in order
// to enforce immutability.
// https://openfeature.dev/specification/sections/evaluation-context
type EvaluationContext struct {
	targetingKey string // uniquely identifying the subject (end-user, or client service) of a flag evaluation
	attributes   map[string]any
}

// Attribute retrieves the attribute with the given key
func (e EvaluationContext) Attribute(key string) any {
	return e.attributes[key]
}

// TargetingKey returns the key uniquely identifying the subject (end-user, or client service) of a flag evaluation
func (e EvaluationContext) TargetingKey() string {
	return e.targetingKey
}

// Attributes returns a copy of the EvaluationContext's attributes
func (e EvaluationContext) Attributes() map[string]any {
	// copy attributes to new map to prevent mutation (maps are passed by reference)
	attrs := make(map[string]any, len(e.attributes))
	maps.Copy(attrs, e.attributes)

	return attrs
}

// Flattened converts EvaluationContext to a [FlattenedContext].
func (e EvaluationContext) Flattened() FlattenedContext {
	flatCtx := FlattenedContext{}
	if e.attributes != nil {
		flatCtx = e.Attributes()
	}
	if e.targetingKey != "" {
		flatCtx[TargetingKey] = e.targetingKey
	}
	return flatCtx
}

// NewEvaluationContext constructs an EvaluationContext
//
// targetingKey - uniquely identifying the subject (end-user, or client service) of a flag evaluation
// attributes - contextual data used in flag evaluation
func NewEvaluationContext(targetingKey string, attributes map[string]any) EvaluationContext {
	// copy attributes to new map to avoid reference being externally available, thereby enforcing immutability
	attrs := make(map[string]any, len(attributes))
	maps.Copy(attrs, attributes)

	return EvaluationContext{
		targetingKey: targetingKey,
		attributes:   attrs,
	}
}

// NewTargetlessEvaluationContext constructs an EvaluationContext with an empty targeting key
//
// attributes - contextual data used in flag evaluation
func NewTargetlessEvaluationContext(attributes map[string]any) EvaluationContext {
	return NewEvaluationContext("", attributes)
}

// ContextWithEvaluationContext constructs a TransactionContext.
//
// ctx - the context to embed the EvaluationContext in
// ec - the EvaluationContext to embed into the context
func ContextWithEvaluationContext(ctx context.Context, ec EvaluationContext) context.Context {
	return context.WithValue(ctx, transactionContext, ec)
}

// MergeTransactionContext merges the provided EvaluationContext with the current TransactionContext (if it exists)
//
// ctx - the context to pull existing TransactionContext from
// ec - the EvaluationContext to merge with the existing TransactionContext
func MergeTransactionContext(ctx context.Context, ec EvaluationContext) context.Context {
	oldTc := EvaluationContextFromContext(ctx)
	mergedTc := mergeContexts(ec, oldTc)
	return ContextWithEvaluationContext(ctx, mergedTc)
}

// EvaluationContextFromContext extracts a EvaluationContext from the current
// context. if no EvaluationContext exist, it will construct
// an empty EvaluationContext.
//
// ctx - the context to pull EvaluationContext from TransactionContext
func EvaluationContextFromContext(ctx context.Context) EvaluationContext {
	ec, ok := extractEvaluationContextFromContext(ctx)
	if !ok {
		return EvaluationContext{}
	}

	return ec
}

// extractEvaluationContextFromContext extracts an EvaluationContext from the context.
// It returns the EvaluationContext and a boolean indicating whether one was found.
//
// ctx - the context to extract the EvaluationContext from
func extractEvaluationContextFromContext(ctx context.Context) (EvaluationContext, bool) {
	ec, ok := ctx.Value(transactionContext).(EvaluationContext)
	return ec, ok
}

// contextKey is just an empty struct. It exists so transactionContext can be
// an immutable variable with a unique type. It's immutable
// because nobody else can create a contextKey, being unexported.
type contextKey struct{}

// transactionContext is the context key to use with golang.org/x/net/context's
// WithValue function to associate an EvaluationContext value with a context.
var transactionContext contextKey
