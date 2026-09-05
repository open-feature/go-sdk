# In-memory provider

`InMemoryProvider` is an OpenFeature compliant provider implementation with an in-memory flag storage. 

While the main usage of this provider is SDK testing, you may use it for minimal OpenFeature use cases where appropriate.

## Updating flags

`UpdateFlags` replaces the flag set and emits a `PROVIDER_CONFIGURATION_CHANGED`
event listing the union of the previous and the new flag keys:

```go
provider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
    "my-flag": {Key: "my-flag", State: memprovider.Enabled, DefaultVariant: "on", Variants: map[string]any{"on": true, "off": false}},
})
if err := openfeature.SetProviderAndWait(provider); err != nil {
    // handle err
}

// handlers registered for openfeature.ProviderConfigChange are invoked
provider.UpdateFlags(map[string]memprovider.InMemoryFlag{
    "my-flag": {Key: "my-flag", State: memprovider.Enabled, DefaultVariant: "off", Variants: map[string]any{"on": true, "off": false}},
})
```

The provider must be held as `*InMemoryProvider`, which is what
`NewInMemoryProvider` returns. The event is dropped rather than blocking when the
provider is not registered with an API, since nothing is listening.

### Flag ownership

Both `NewInMemoryProvider` and `UpdateFlags` copy the map they are given, so
mutating it afterwards does not affect the provider. The copy is shallow: the
provider takes ownership of each flag's `Variants` and `ContextEvaluator`, which
must not be mutated once the flag has been handed over. `UpdateFlags` is the
only supported way to change the flag set — it is what emits the
`PROVIDER_CONFIGURATION_CHANGED` event.
