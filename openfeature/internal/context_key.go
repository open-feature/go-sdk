package internal

// ContextKey is just an empty struct. It exists so TranscationContext can be
// an immutable public variable with a unique type. It's immutable
// because nobody else can create a ContextKey, being unexported.
type ContextKey struct{}

// TranscationContext is the context key to use with golang.org/x/net/context's
// WithValue function to associate an EvaluationContext value with a context.
var TranscationContextKey ContextKey
