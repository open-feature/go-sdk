# OpenFeature Go SDK - Migration Guide

## v1 to v2

This guide helps you upgrade from OpenFeature Go SDK v1.x to v2.0. The v2 release includes significant improvements to the API design, with several **breaking changes** that require code updates.

## Table of Contents

1. [Summary of Changes](#summary-of-changes)
2. [Breaking Changes](#breaking-changes)
3. [Migration Steps](#migration-steps)
4. [Code Examples](#code-examples)
5. [FAQ](#faq)

---

## Summary of Changes

OpenFeature Go SDK v2 introduces a modern, context-aware API design with the following key improvements:

- **Context-aware provider lifecycle**: All provider methods now require `context.Context` parameter
- **Simplified hook interface**: Hook signatures updated with better context handling
- **Type-safe evaluation**: Improved type system using Go generics and type constraints
- **Consistent naming**: Removed deprecated methods and standardized interface names
- **Better error handling**: Improved error propagation and handling patterns
- **Removed logger dependency**: No longer depends on `go-logr`, use slog or your preferred logging

---

## Breaking Changes

### 1. **IClient Interface Refactored**

**v1:**

```go
type IClient interface {
    Metadata() ClientMetadata
    AddHooks(hooks ...Hook)
    SetEvaluationContext(evalCtx EvaluationContext)
    EvaluationContext() EvaluationContext
    BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
    StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
    // ... other *Value methods
    BooleanValueDetails(...) (BooleanEvaluationDetails, error)
    StringValueDetails(...) (StringEvaluationDetails, error)
    // ... other *ValueDetails methods
    Boolean(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) bool
    String(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) string
    // ... other shorthand methods
    State() State
    IEventing
    Tracker
}
```

**v2:**

- `IClient` is now private (`iClient`) and split into composable interfaces
- Value methods (e.g., `BooleanValue`, `StringValue`) **removed**. Use `BooleanValueDetails`, `StringValueDetails`, etc. instead
- `State()` method removed; use `IEventing` methods directly

**Migration Path:**

```go
// v1: Using *Value methods
value, err := client.StringValue(ctx, "flag-key", "default", evalCtx)

// v2: Use *ValueDetails instead - details contain both value and metadata
details, err := client.StringValueDetails(ctx, "flag-key", "default", evalCtx)
if err != nil {
    return "default"
}
value := details.Value
```

---

### 2. **Context-Aware Provider Lifecycle**

**v1:**

```go
type StateHandler interface {
    Init(evaluationContext EvaluationContext) error
    Shutdown()
}

type ContextAwareStateHandler interface {
    StateHandler
    InitWithContext(ctx context.Context, evaluationContext EvaluationContext) error
    ShutdownWithContext(ctx context.Context) error
}
```

**v2:**

```go
type StateHandler interface {
    Init(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

**Migration Path - Provider Implementation:**

```go
// v1: Old implementation
type MyProvider struct {
}

func (p *MyProvider) Init(evalCtx EvaluationContext) error {
    // Initialize without context awareness
    return nil
}

func (p *MyProvider) Shutdown() {
    // Cleanup without graceful timeout
}

// v2: Updated implementation
type MyProvider struct {
    connection *sql.DB
}

func (p *MyProvider) Init(ctx context.Context) error {
    evalCtx := openfeature.TransactionContext(ctx)
    //...
    return p.c.Ping(ctx)
}

func (p *MyProvider) Shutdown(ctx context.Context) error {
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return p.c.Close(ctx) // or implement graceful close
}
```

---

### 3. **Provider Setup API Changes**

**v1:**

```go
// Async setup (non-blocking)
api.SetProvider(provider)
api.SetProviderWithContext(ctx, provider)
api.SetNamedProvider(clientName, provider, async)
api.SetNamedProviderWithContext(ctx, clientName, provider, async)

// Sync setup (blocking)
api.SetProviderAndWait(provider)
api.SetProviderAndWaitWithContext(ctx, provider)
api.SetNamedProviderWithContextAndWait(ctx, clientName, provider)
```

**v2:**

```go
// All methods now require context
api.SetProvider(ctx, provider)                           // async
api.SetProviderAndWait(ctx, provider)                    // sync
api.SetNamedProvider(ctx, clientName, provider)          // async
api.SetNamedProviderAndWait(ctx, clientName, provider)   // sync
```

**Migration Path:**

```go
// v1
api.SetProvider(provider)

// v2: Explicit context required
api.SetProvider(context.Background(), provider)

// v1: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
api.SetProviderWithContext(ctx, provider)

// v2: Same pattern
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
api.SetProvider(ctx, provider)
```

---

### 4. **Shutdown API Changes**

**v1:**

```go
// Non-context aware
api.Shutdown()

// Context-aware
err := api.ShutdownWithContext(ctx)
```

**v2:**

```go
// All shutdown requires context
err := api.Shutdown(ctx)
```

**Migration Path:**

```go
// v1
api.Shutdown()

// v2: Use context for shutdown timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := api.Shutdown(ctx)
```

---

### 5. **Hook Interface Changes**

**v1:**

```go
type Hook interface {
    Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (*EvaluationContext, error)
    After(ctx context.Context, hookContext HookContext, flagEvaluationDetails InterfaceEvaluationDetails, hookHints HookHints) error
    Error(ctx context.Context, hookContext HookContext, err error, hookHints HookHints)
    Finally(ctx context.Context, hookContext HookContext, flagEvaluationDetails InterfaceEvaluationDetails, hookHints HookHints)
}
```

**v2:**

```go
type Hook interface {
    Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (context.Context, error)
    After(ctx context.Context, hookContext HookContext, flagEvaluationDetails HookEvaluationDetails, hookHints HookHints) error
    Error(ctx context.Context, hookContext HookContext, err error, hookHints HookHints)
    Finally(ctx context.Context, hookContext HookContext, flagEvaluationDetails HookEvaluationDetails, hookHints HookHints)
}
```

**Key Differences:**

- `Before` hook now returns `(context.Context, error)` instead of `(*EvaluationContext, error)`
  - EvaluationContext is now passed through context using `WithTransactionContext()` and retrieved via `TransactionContext()`
  - This allows hooks to modify both context and evaluation context transparently
- `InterfaceEvaluationDetails` renamed to `HookEvaluationDetails` in `After` and `Finally` methods
- Hooks can now influence the evaluation context by modifying the context passed to subsequent hooks

**Migration Path:**

```go
// v1: Hook implementation
func (h *MyHook) Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (*EvaluationContext, error) {
    evalCtx := hookContext.EvaluationContext()
    evalCtx.SetString("user_id", "123")
    return &evalCtx, nil
}


// v2: Updated hook implementation
func (h *MyHook) Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (context.Context, error) {
    evalCtx := hookContext.EvaluationContext()
    evalCtx.SetString("user_id", "123")

    // Attach evaluation context to the context for downstream use
    return WithTransactionContext(ctx, evalCtx), nil
}

```

---

### 6. **Type Naming Changes**

**v1:**

```go
type InterfaceEvaluationDetails = GenericEvaluationDetails[any]
type InterfaceResolutionDetail = GenericResolutionDetail[any]
```

**v2:**

```go
type ObjectEvaluationDetails = EvaluationDetails[any]
type ObjectResolutionDetail = GenericResolutionDetail[any]
```

**Migration Path:**

```go
// v1
var details openfeature.InterfaceEvaluationDetails

// v2
var details openfeature.ObjectEvaluationDetails
```

---

### 7. **Deprecated Methods Removed**

**Removed Methods:**

1. `ClientMetadata.Name()` - Use `ClientMetadata.Domain()` instead
2. `Client.WithLogger()` - Use hooks or external logging instead (e.g., slog)
3. All `*Value()` methods - Use `*ValueDetails()` instead
4. `API.SetLogger()` - Use hooks or external logging instead

**Migration Path:**

```go
// v1: Deprecated
metadata.Name()

// v2: Use Domain()
metadata.Domain()

// v1: Logger on client
client.WithLogger(logger)

// v2: Use LoggingHook from openfeature/hooks package
import "github.com/open-feature/go-sdk/openfeature/hooks"

loggingHook := hooks.NewLoggingHook() // or implement your own
openfeature.GetApi().AddHooks(loggingHook)
```

---

### 8. **Package Restructuring**

**v1:**

```
openfeature/
  memprovider/          # In-memory provider
  multi/                # Multi-provider strategies
```

**v2:**

```
openfeature/
  providers/
    inmemory/           # Renamed from memprovider
    multi/              # Moved under providers
```

**Migration Path:**

```go
// v1
import "github.com/open-feature/go-sdk/openfeature/memprovider"
provider := memprovider.NewInMemoryProvider(flags)

// v2
import "go.openfeature.dev/openfeature/v2/providers/inmemory"
provider := inmemory.NewProvider(flags)
```

---

### 9. **EventCallback Type Change**

**v1:**

```go
type EventCallback *func(details EventDetails)
```

**v2:**

```go
type EventCallback func(details EventDetails)
```

**Migration Path:**

```go
// v1: Pointer to function
callback := &func(details EventDetails) {
    // Handle event
}
api.OnProviderReady(callback)

// v2: Direct function
api.OnProviderReady(func(details EventDetails) {
    // Handle event
})
```

---

### 10. **Interface Visibility Changes**

**v1:**

```go
type IClient interface { ... }      // Public
type IEvaluation interface { ... }  // Public
```

**v2:**

```go
type iClient interface { ... }      // Private (internal)
type Evaluator interface { ... }    // Public, focused interface
type DetailEvaluator interface { ...} // Public, focused interface
```

Internal interfaces are now private. Use the provided public interfaces instead.

---

## Migration Steps

### Step 0: Install

```sh
go get "go.openfeature.dev/openfeature/v2"

```

### Step 1: Update Provider Implementation

If you have custom providers, update the `StateHandler` implementation:

```go
// OLD
func (p *MyProvider) Init(evalCtx EvaluationContext) error {
    // ...
}

func (p *MyProvider) Shutdown() {
    // ...
}

// NEW
func (p *MyProvider) Init(ctx context.Context) error {
    // Use ctx for timeout control
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // ...
}

func (p *MyProvider) Shutdown(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // ...
    return nil
}
```

### Step 2: Update Hook Implementations

If you have custom hooks, update the `Before` method signature:

```go
// OLD
func (h *MyHook) Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (*EvaluationContext, error) {
    evalCtx := hookContext.EvaluationContext()
    return &evalCtx, nil
}

// NEW
func (h *MyHook) Before(ctx context.Context, hookContext HookContext, hookHints HookHints) (context.Context, error) {
    evalCtx := hookContext.EvaluationContext()
    // Attach evaluation context to context for downstream use
    return WithTransactionContext(ctx, evalCtx), nil
}
```

Also rename `InterfaceEvaluationDetails` to `HookEvaluationDetails` in your `After` and `Finally` hook methods, and use `TransactionContext()` to retrieve the evaluation context from the context if needed.

### Step 3: Update Provider Setup Calls

All provider setup now requires explicit `context.Context`:

```go
// OLD
api.SetProvider(myProvider)
api.SetProviderAndWait(myProvider)
api.SetNamedProvider("my-client", myProvider, true)

// NEW
api.SetProvider(context.Background(), myProvider)
api.SetProviderAndWait(context.Background(), myProvider)
api.SetNamedProvider(context.Background(), "my-client", myProvider)
```

### Step 4: Update Shutdown Calls

```go
// OLD
api.Shutdown()

// NEW
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
api.Shutdown(ctx)
```

### Step 5: Update Evaluation Calls

Remove usage of `*Value` methods; use `*ValueDetails` instead:

```go
// OLD
value, err := client.BooleanValue(ctx, "flag", false, evalCtx)
if err != nil {
    // handle error
}
// use value

// NEW
details, err := client.BooleanValueDetails(ctx, "flag", false, evalCtx)
if err != nil {
    // handle error
}
// use details.Value
```

### Step 6: Update Imports

```go
// OLD
import "github.com/open-feature/go-sdk/openfeature/memprovider"

// NEW
import "go.openfeature.dev/openfeature/v2/providers/inmemory"
```

### Step 7: Remove Logger Configuration

```go
// OLD
api.SetLogger(myLogger)
client.WithLogger(myLogger)

// NEW - Use hooks instead
import "go.openfeature.dev/openfeature/v2/hooks"
loggingHook := hooks.NewLoggingHook()
api.AddHooks(loggingHook)
```
