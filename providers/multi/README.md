## OpenFeature Multi-Provider

> [!WARNING]
> The multi package for the go-sdk is experimental.

The multi-provider allows you to use multiple underlying providers as sources of flag data for the OpenFeature server SDK.
The multi-provider acts as a wrapper providing a unified interface to interact with all of those providers at once.
When a flag is being evaluated, the Multi-Provider will consult each underlying provider it is managing in order to
determine the final result. Different evaluation strategies can be defined to control which providers get evaluated and
which result is used.

The multi-provider is defined within [Appendix A: Included Utilities](https://openfeature.dev/specification/appendix-a#multi-provider)
of the openfeature spec.

The multi-provider is a powerful tool for performing migrations between flag providers, or combining multiple providers
into a single feature flagging interface. For example:

- **Migration**: When migrating between two providers, you can run both in parallel under a unified flagging interface.
  As flags are added to the new provider, the multi-provider will automatically find and return them, falling back to the old provider
  if the new provider does not have
- **Multiple Data Sources**: The multi-provider allows you to seamlessly combine many sources of flagging data, such as
  environment variables, local files, database values and SaaS hosted feature management systems.

# Usage

```go
import (
 "go.openfeature.dev/openfeature/v2"
 "go.openfeature.dev/openfeature/v2/providers/inmemory"
 "go.openfeature.dev/openfeature/v2/providers/multi"
)

mprovider, err := multi.NewProvider(
  multi.StrategyFirstMatch,
  multi.WithProvider("providerA", inmemory.NewProvider(/*...*/)),
  multi.WithProvider("providerB", myCustomProvider),
)
if err != nil {
  return err
}

openfeature.SetNamedProviderAndWait(context.TODO(), "multiprovider", mprovider)
```

# Strategies

There are three strategies that are defined by the spec and are available within this multi-provider implementation. In
addition to the three provided strategies a custom strategy can be defined as well.

The three provided strategies are:

- _First Match_
- _First Success_
- _Comparison_

## First Match Strategy

The first match strategy works by **sequentially** calling each provider until a valid result is returned.
The first provider that returns a result will be used. It will try calling the next provider whenever it encounters a `FLAG_NOT_FOUND`
error. However, if a provider returns an error other than `FLAG_NOT_FOUND` the provider will stop and return the default
value along with setting the error details if a detailed request is issued.

## First Success Strategy

The first success strategy also works by calling each provider **sequentially**. The first provider that returns a response
with no errors is used. This differs from the first match strategy in that any provider raising an error will not halt
calling the next provider if a successful result has not yet been encountered. If no provider provides a successful result
the default value will be returned to the caller.

## Comparison Strategy

The comparison strategy works by calling each provider in **parallel**. All results are collected from each provider and
then the resolved results are compared to each other. If they all agree then that value is returned. If not a fallback
provider can be specified to be executed instead or the default value will be returned. If a provider returns
`FLAG_NOT_FOUND` that result will not be included in the comparison. If all providers return not found then the default
value is returned. Finally, if any provider returns an error other than `FLAG_NOT_FOUND` the evaluation immediately stops
and that error result is returned with the default value.

The fallback provider can be set using the `WithFallbackProvider` [`Option`](#options).

Special care must be taken when this strategy is used with `ObjectEvaluation`. If the resulting value is not a
[`comparable`](https://go.dev/blog/comparable) type then the default result or fallback provider will always be used. In
order to evaluate non `comparable` types a `Comparator` function must be provided as an `Option` to the constructor.

## Custom Strategies

A custom strategy can be defined using the `WithCustomStrategy` `Option` along with the `StrategyCustom` constant.
A custom strategy is defined by the following generic function signature:

```go
StrategyFn[T FlagTypes] func(resolutions ResolutionIterator[T], defaultValue T, fallbackEvaluator FallbackEvaluator[T]) openfeature.GenericResolutionDetail[T]
```

Where:

```go
ResolutionIterator[T FlagTypes] = iter.Seq2[string, openfeature.GenericResolutionDetail[T]]
FallbackEvaluator[T FlagTypes] = func(fallbackProvider openfeature.FeatureProvider) openfeature.GenericResolutionDetail[T]
```

The strategy function receives:

- `resolutions`: An iterator of provider names and their resolution results
- `defaultValue`: The default value to return if strategy fails
- `fallbackEvaluator`: A function to evaluate the fallback provider if needed

The `StrategyConstructor` type is used to create your custom strategy:

```go
type StrategyConstructor func() StrategyFn[FlagTypes]
```

Build your custom strategy like this:

```go
option := multi.WithCustomStrategy(func() StrategyFn[openfeature.FlagTypes] {
 return func(resolutions ResolutionIterator[openfeature.FlagTypes], defaultValue openfeature.FlagTypes, fallbackEvaluator FallbackEvaluator[openfeature.FlagTypes]) *openfeature.GenericResolutionDetail[openfeature.FlagTypes] {
  // Iterate through provider resolutions
  for name, resolution := range resolutions {
   // Your custom logic here
   // ...
  }
  // Return selected resolution or use fallbackEvaluator if needed
  return resolution
    }
})
```

It is highly recommended to use the provided exposed function `BuildDefaultResult` when building your custom strategy.

The `BuildDefaultResult` method should be called when an error is encountered or the strategy "fails" and needs to return
the default result passed to one of the Evaluation methods of `openfeature.FeatureProvider`.

# Options

The `multi.NewProvider` constructor implements the optional pattern for setting additional configuration.

## General Options

### `WithLogger`

Allows for providing a `slog.Logger` instance for internal logging of the multi-provider's evaluation processing for debugging
purposes. By default, are logs are discarded unless this option is used.

### `WithCustomStrategy`

Allows for setting a custom strategy function for the evaluation of providers. This must be used in conjunction with the
`StrategyCustom` `EvaluationStrategy` parameter. The option itself takes a `StrategyConstructor` function, which is
essentially a factory that allows the `StrategyFn` to wrap around a slice of `NamedProvider` instances.

### `WithGlobalHooks`

Allows for setting global hooks for the multi-provider. These are `openfeature.Hook` implementations that affect
**all** internal `FeatureProvider` instances.

### `WithProvider`

Allows for registering a specific `FeatureProvider` instance under a unique provider name. Optional `openfeature.Hook`
implementations may also be provided, which will execute only for this specific provider. This option can be used multiple
times with unique provider names to register multiple providers.
The order in which `WithProvider` options are provided determines the order in which the providers are registered and evaluated.

## `StrategyComparision` specific options

There are two options specifically for usage with the `StrategyComparision` `EvaluationStrategy`. If these options are
used with a different `EvaluationStrategy` they are ignored.

### `WithFallbackProvider`

When the results are not in agreement with each other the fallback provider will be called. The result of this provider
is what will be returned to the caller. If no fallback provider is set then the default value will be returned instead.

### `WithCustomComparator`

When using `ObjectEvaluation` there are cases where the results are not able to be compared to each other by default.
This happens if the returned type is not `comparable`. In that situation all the results are passed to the custom `Comparator`
to evaluate if they are in agreement or not. If not provided and the return type is not `comparable` then either the fallback
provider is used or the default value.
