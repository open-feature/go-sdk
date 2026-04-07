# Testing provider

`TestProvider` is an OpenFeature compliant provider implementation designed for testing applications with feature flags.

The testing provider allows you to define feature flag values scoped to individual tests, ensuring test isolation and preventing flag state from leaking between tests. It uses the `InMemoryProvider` internally with per-test flag storage.

# Usage

```go
import (
 "testing"

 "go.openfeature.dev/openfeature/v2"
 "go.openfeature.dev/openfeature/v2/providers/inmemory"
 "go.openfeature.dev/openfeature/v2/providers/testing"
)

testProvider := testing.NewProvider()
err := openfeature.SetProviderAndWait(t.Context(), testProvider)
if err != nil {
 t.Fatal(err)
}

ctx := testProvider.UsingFlags(t, map[string]memprovider.InMemoryFlag{
 "my_feature": {
  State:          memprovider.Enabled,
  DefaultVariant: "on",
  Variants:       map[string]any{"on": true},
 },
})

client := openfeature.NewClient()
result := client.Boolean(ctx, "my_feature", false, openfeature.EvaluationContext{})
```

If your test code runs in a different goroutine, `TestProvider.UsingFlags` returns a context that should be used for evaluations.

You can pass `*testing.T` directly.

```go
// In your test, 't' is a *testing.T
ctx := testProvider.UsingFlags(t, tt.flags)

go func() {
    // Make sure to use the context returned by UsingFlags in the new goroutine.
    // The context carries the necessary information for the TestProvider.
    _ = openfeature.NewDefaultClient().Boolean(ctx, "my_flag", false, openfeature.EvaluationContext{})
}()
```
