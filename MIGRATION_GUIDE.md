# OpenFeature Go SDK - Migration Guide

## v1 to v2

This guide helps you upgrade from OpenFeature Go SDK v1.x to v2.0. The v2 release includes significant improvements to the API design, with several **breaking changes** that require code updates.

## Table of Contents

1. [Summary of Changes](#summary-of-changes)
2. [Breaking Changes](#breaking-changes)
3. [Migration Steps](#migration-steps)

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
- Value methods that return `(value, error)` (e.g., `BooleanValue`, `StringValue`) **removed**. Use `Boolean`, `String`, etc. (non-error) or `BooleanValueDetails`, `StringValueDetails`, etc. (with metadata) instead
- `State()` method removed; use `Eventing` methods directly

**Migration Path:**

```go
// v1: Using *Value methods (returns value with error)
value, err := client.StringValue(ctx, "flag-key", "default", evalCtx)

// v2: Use non-error variants - String, Boolean, Int, Float, Object
// These return only the value (with default on error)
value := client.String(ctx, "flag-key", "default", evalCtx)

// Or use *ValueDetails for evaluation metadata
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
}

func (p *MyProvider) Init(ctx context.Context) error {
    evalCtx := openfeature.EvaluationContextFromContext(ctx)
    // Use passed context directly, or create a new timeout context if needed
    return nil// or err
}

func (p *MyProvider) Shutdown(ctx context.Context) error {
    // Implement graceful close
    return nil// or err
}
```

---

### 3. **Provider Setup API Changes**

**v1:**

```go
// Async setup (non-blocking)
api.SetProvider(provider)
api.SetProviderWithContext(ctx, provider)
api.SetNamedProvider(domain, provider)
api.SetNamedProviderWithContext(ctx, domain, provider)

// Sync setup (blocking)
api.SetProviderAndWait(provider)
api.SetProviderAndWaitWithContext(ctx, provider)
api.SetNamedProviderAndWait(domain, provider)
api.SetNamedProviderWithContextAndWait(ctx, domain, provider)
```

**v2:**

```go
// All methods now require context and use options
api.SetProvider(ctx, provider)                                            // async
api.SetProviderAndWait(ctx, provider)                                     // sync
api.SetProvider(ctx, provider, openfeature.WithDomain(domain))            // async with domain
api.SetProviderAndWait(ctx, provider, openfeature.WithDomain(domain))     // sync with domain
```

**Migration Path:**

```go
// v1: Async without context
api.SetProvider(provider)

// v2: Explicit context required (use context.TODO() if no timeout needed)
api.SetProvider(context.TODO(), provider)

// v1: Async with context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
api.SetProviderWithContext(ctx, provider)

// v2: Same pattern - context is now required
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
api.SetProvider(ctx, provider)

// v1: Named provider (domain-specific)
api.SetNamedProvider("user-service", userProvider)
api.SetNamedProviderWithContext(ctx, "user-service", userProvider)

// v2: Use WithDomain option
api.SetProvider(context.TODO(), userProvider, openfeature.WithDomain("user-service"))
api.SetProvider(ctx, userProvider, openfeature.WithDomain("user-service"))

// v1: Sync operations
api.SetProviderAndWait(provider)
api.SetProviderAndWaitWithContext(ctx, provider)
api.SetNamedProviderAndWait("user-service", userProvider)
api.SetNamedProviderWithContextAndWait(ctx, "user-service", userProvider)

// v2: Sync with context and domain option
api.SetProviderAndWait(context.TODO(), provider)
api.SetProviderAndWait(ctx, provider)
api.SetProviderAndWait(context.TODO(), userProvider, openfeature.WithDomain("user-service"))
api.SetProviderAndWait(ctx, userProvider, openfeature.WithDomain("user-service"))
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
  - EvaluationContext is now passed through context using `ContextWithEvaluationContext()` and retrieved via `EvaluationContextFromContext()`
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
    return ContextWithEvaluationContext(ctx, evalCtx), nil
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
import "go.openfeature.dev/openfeature/v2/hooks"

loggingHook := hooks.NewLoggingHook() // or implement your own
openfeature.AddHooks(loggingHook)
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

### 11. **Client Creation and Metadata API**

**v1:**

```go
// Client creation with positional domain argument
client := openfeature.NewClient("domain-name")
client := openfeature.NewDefaultClient()  // Default domain client

// Named provider metadata
metadata := api.NamedProviderMetadata("domain-name")
```

**v2:**

```go
// Client creation using WithDomain option
client := openfeature.NewClient(openfeature.WithDomain("domain-name"))
client := openfeature.NewClient()  // Default domain client (no argument needed)

// Unified provider metadata
metadata := api.ProviderMetadata(openfeature.WithDomain("domain-name"))
```

**Key Differences:**

- `NewClient(domainName)` → `NewClient(openfeature.WithDomain(domainName))` - domain now passed as option
- `NewDefaultClient()` → `NewClient()` - simplified to single method
- `NamedProviderMetadata(domain)` → `ProviderMetadata(openfeature.WithDomain(domain))` - unified with default metadata API
- Domain is now specified as a `CallOption` rather than a positional string argument

**Migration Path:**

```go
// v1: Domain as positional argument
client := openfeature.NewClient("user-service")
defaultClient := openfeature.NewDefaultClient()
metadata := api.NamedProviderMetadata("user-service")

// v2: Domain as option
client := openfeature.NewClient(openfeature.WithDomain("user-service"))
defaultClient := openfeature.NewClient()
metadata := api.ProviderMetadata(openfeature.WithDomain("user-service"))
```

---

## Migration Steps

### Automated Migration with gopatch (Optional)

To help automate the migration, you can use the [uber-go/gopatch](https://github.com/uber-go/gopatch) tool to apply common transformations across your codebase.

#### Install gopatch

```sh
go install github.com/uber-go/gopatch@latest
```

#### Apply the patch file

The `openfeature_v1_to_v2.patch` file in this repository contains transformations for the most common breaking changes. Apply it to your codebase:

```sh
gopatch -p ./openfeature_v1_to_v2.patch ./...
```

#### What the patch covers

The patch file automatically handles:

- **Evaluation methods**: Converts `StringValue()`, `BooleanValue()`, `IntValue()`, `FloatValue()`, and `ObjectValue()` to their non-error counterparts `String()`, `Boolean()`, `Int()`, `Float()`, and `Object()`
- **Provider setup**: Updates all provider setup calls to require explicit context:
  - `SetProvider(provider)` → `SetProvider(ctx, provider)`
  - `SetProviderAndWait(provider)` → `SetProviderAndWait(ctx, provider)`
  - `SetNamedProvider()` and related methods
- **Shutdown calls**: Migrates `Shutdown()` and `ShutdownWithContext()` to context-aware `Shutdown(ctx)`
- **Package imports**: Updates provider imports from `memprovider` to `providers/inmemory`
- **Type names**: Renames `InterfaceEvaluationDetails` to `ObjectEvaluationDetails`

#### After running gopatch

After running the patch, review the changes and manually update:

- **Hook implementations** - The patch cannot fully automate hook migrations. See Step 2 below for manual updates needed to `Before` method signatures.
- **Custom provider implementations** - Update `StateHandler` implementations as shown in Step 1 below.
- **Error handling** - Review error handling logic for `*Value()` methods which may need adjustment.
- Do `go mod tidy` and install missing dependencies

---

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
    // ...
    return nil
}

func (p *MyProvider) Shutdown(ctx context.Context) error {
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
    return ContextWithEvaluationContext(ctx, evalCtx), nil
}
```

Also rename `InterfaceEvaluationDetails` to `HookEvaluationDetails` in your `After` and `Finally` hook methods, and use `EvaluationContextFromContext()` to retrieve the evaluation context from the context if needed.

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
api.SetProvider(context.Background(), myProvider, openfeature.WithDomain("my-client"))
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

Replace usage of `*Value` methods with non-error variants (`Boolean`, `String`, `Int`, `Float`, `Object`) for simple cases, or use `*ValueDetails` for detailed evaluation metadata:

```go
// OLD
value, err := client.BooleanValue(ctx, "flag", false, evalCtx)
if err != nil {
    // handle error
}
// use value

// NEW: Simple case - use non-error variant
value := client.Boolean(ctx, "flag", false, evalCtx)
// use value

// NEW: Need evaluation metadata - use *ValueDetails
details, err := client.BooleanValueDetails(ctx, "flag", false, evalCtx)
if err != nil {
    // handle error
}
// use details.Value, details.Reason, details.Variant, etc.
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

### Step 8: Update new client and set provider calls

```go
// OLD
api.SetNamedProvider(ctx, "user-service", userProvider)
api.SetNamedProviderAndWait(ctx, "billing-service", billingProvider)
userClient := openfeature.NewClient("user-service")
billingClient := openfeature.NewClient("billing-service")
defaultClient := openfeature.NewDefaultClient()
metadata := api.NamedProviderMetadata("user-service")

// NEW
api.SetProvider(ctx, userProvider, openfeature.WithDomain("user-service"))
api.SetProviderAndWait(ctx, billingProvider, openfeature.WithDomain("billing-service"))
userClient := openfeature.NewClient(openfeature.WithDomain("user-service"))
billingClient := openfeature.NewClient(openfeature.WithDomain("billing-service"))
defaultClient := openfeature.NewClient()
metadata := api.ProviderMetadata(openfeature.WithDomain("user-service"))
```
