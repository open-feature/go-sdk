<!-- markdownlint-disable MD033 -->
<!-- x-hide-in-docs-start -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/white/openfeature-horizontal-white.svg" />
    <img align="center" alt="OpenFeature Logo" src="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/black/openfeature-horizontal-black.svg" />
  </picture>
</p>

<h2 align="center">OpenFeature Go SDK</h2>

<!-- x-hide-in-docs-end -->
<!-- The 'github-badges' class is used in the docs -->
<p align="center" class="github-badges">
  <a href="https://github.com/open-feature/spec/releases/tag/v0.7.0">
    <img alt="Specification" src="https://img.shields.io/static/v1?label=specification&message=v0.7.0&color=yellow&style=for-the-badge" />
  </a>
  <!-- x-release-please-start-version -->
  <a href="https://github.com/open-feature/go-sdk/releases/tag/v1.15.1">
    <img alt="Release" src="https://img.shields.io/static/v1?label=release&message=v1.15.1&color=blue&style=for-the-badge" />
  </a>
  <!-- x-release-please-end -->
  <br/>
  <a href="https://pkg.go.dev/github.com/open-feature/go-sdk/openfeature">
    <img alt="API Reference" src="https://pkg.go.dev/badge/github.com/open-feature/go-sdk/pkg/openfeature.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/open-feature/go-sdk">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/open-feature/go-sdk" />
  </a>
  <a href="https://codecov.io/gh/open-feature/go-sdk">
    <img alt="codecov" src="https://codecov.io/gh/open-feature/go-sdk/branch/main/graph/badge.svg?token=FZ17BHNSU5" />
  </a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/6601">
    <img alt="CII Best Practices" src="https://bestpractices.coreinfrastructure.org/projects/6601/badge" />
  </a>
</p>
<!-- x-hide-in-docs-start -->

[OpenFeature](https://openfeature.dev) is an open specification that provides a vendor-agnostic, community-driven API for feature flagging that works with your favorite feature flag management tool.

<!-- x-hide-in-docs-end -->
## 🚀 Quick start

### Requirements

Go language version: [1.24](https://go.dev/doc/devel/release#go1.24.0)

> [!NOTE]
> The OpenFeature Go SDK only supports currently maintained Go language versions.

### Install

```shell
go get github.com/open-feature/go-sdk
```

### Usage

```go
package main

import (
    "fmt"
    "context"
    "github.com/open-feature/go-sdk/openfeature"
)

func main() {
    // Register your feature flag provider
    openfeature.SetProviderAndWait(openfeature.NoopProvider{})
    // Create a new client
    client := openfeature.NewClient("app")
    // Evaluate your feature flag
    v2Enabled := client.Boolean(
        context.TODO(), "v2_enabled", true, openfeature.EvaluationContext{},
    )
    // Use the returned flag value
    if v2Enabled {
        fmt.Println("v2 is enabled")
    }
}
```

Try this example in the [Go Playground](https://go.dev/play/p/fSSK8s42hA2).

### API Reference

See [here](https://pkg.go.dev/github.com/open-feature/go-sdk/openfeature) for the complete API documentation.

## 🌟 Features

| Status | Features                                                            | Description                                                                                                                                                  |
| ------ |---------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ✅      | [Providers](#providers)                                             | Integrate with a commercial, open source, or in-house feature management tool.                                                                               |
| ✅      | [Targeting](#targeting)                                             | Contextually-aware flag evaluation using [evaluation context](https://openfeature.dev/docs/reference/concepts/evaluation-context).                           |
| ✅      | [Hooks](#hooks)                                                     | Add functionality to various stages of the flag evaluation life-cycle.                                                                                       |
| ✅      | [Tracking](#tracking)                                               | Associate user actions with feature flag evaluations.                                                                                                        |
| ✅      | [Logging](#logging)                                                 | Integrate with popular logging packages.                                                                                                                     |
| ✅      | [Domains](#domains)                                                 | Logically bind clients with providers.                                                                                                                       |
| ✅      | [Eventing](#eventing)                                               | React to state changes in the provider or flag management system.                                                                                            |
| ✅      | [Shutdown](#shutdown)                                               | Gracefully clean up a provider during application shutdown.                                                                                                  |
| ✅      | [Transaction Context Propagation](#transaction-context-propagation) | Set a specific [evaluation context](https://openfeature.dev/docs/reference/concepts/evaluation-context) for a transaction (e.g. an HTTP request or a thread) |
| ✅      | [Extending](#extending)                                             | Extend OpenFeature with custom providers and hooks.                                                                                                          |

<sub>Implemented: ✅ | In-progress: ⚠️ | Not implemented yet: ❌</sub>

### Providers

[Providers](https://openfeature.dev/docs/reference/concepts/provider) are an abstraction between a flag management system and the OpenFeature SDK.
Look [here](https://openfeature.dev/ecosystem?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Provider&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go) for a complete list of available providers.
If the provider you're looking for hasn't been created yet, see the [develop a provider](#develop-a-provider) section to learn how to build it yourself.

Once you've added a provider as a dependency, it can be registered with OpenFeature like this:

```go
openfeature.SetProviderAndWait(MyProvider{})
```

In some situations, it may be beneficial to register multiple providers in the same application.
This is possible using [domains](#domains), which is covered in more details below.

### Targeting

Sometimes, the value of a flag must consider some dynamic criteria about the application or user, such as the user's location, IP, email address, or the server's location.
In OpenFeature, we refer to this as [targeting](https://openfeature.dev/specification/glossary#targeting).
If the flag management system you're using supports targeting, you can provide the input data using the [evaluation context](https://openfeature.dev/docs/reference/concepts/evaluation-context).

```go
// set a value to the global context
openfeature.SetEvaluationContext(openfeature.NewTargetlessEvaluationContext(
    map[string]any{
        "region":  "us-east-1-iah-1a",
    },
))

// set a value to the client context
client := openfeature.NewClient("my-app")
client.SetEvaluationContext(openfeature.NewTargetlessEvaluationContext(
    map[string]any{
        "version":  "1.4.6",
    },
))

// set a value to the invocation context
evalCtx := openfeature.NewEvaluationContext(
    "user-123",
    map[string]any{
        "company": "Initech",
    },
)
boolValue, err := client.BooleanValue("boolFlag", false, evalCtx)
```


### Hooks

[Hooks](https://openfeature.dev/docs/reference/concepts/hooks) allow for custom logic to be added at well-defined points of the flag evaluation life-cycle
Look [here](https://openfeature.dev/ecosystem/?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Hook&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go) for a complete list of available hooks.
If the hook you're looking for hasn't been created yet, see the [develop a hook](#develop-a-hook) section to learn how to build it yourself.

Once you've added a hook as a dependency, it can be registered at the global, client, or flag invocation level.

```go
// add a hook globally, to run on all evaluations
openfeature.AddHooks(ExampleGlobalHook{})

// add a hook on this client, to run on all evaluations made by this client
client := openfeature.NewClient("my-app")
client.AddHooks(ExampleClientHook{})

// add a hook for this evaluation only
value, err := client.BooleanValue(
    context.TODO(), "boolFlag", false, openfeature.EvaluationContext{}, WithHooks(ExampleInvocationHook{}),
)
```

### Tracking

The [tracking API](https://openfeature.dev/specification/sections/tracking/) allows you to use OpenFeature abstractions and objects to associate user actions with feature flag evaluations.
This is essential for robust experimentation powered by feature flags.
For example, a flag enhancing the appearance of a UI component might drive user engagement to a new feature; to test this hypothesis, telemetry collected by a [hook](#hooks) or [provider](#providers) can be associated with telemetry reported in the client's `track` function.

```go
// initilize a client
client := openfeature.NewClient('my-app')

// trigger tracking event action
client.Track(
    context.TODO(),
    'visited-promo-page',
    openfeature.EvaluationContext{},
    openfeature.NewTrackingEventDetails(99.77).Add("currencyCode", "USD"),
    )
```

Note that some providers may not support tracking; check the documentation for your provider for more information.

### Logging

Note that in accordance with the OpenFeature specification, the SDK doesn't generally log messages during flag evaluation.

#### Logging Hook

The GO SDK includes a `LoggingHook`, which logs detailed information at key points during flag evaluation, using [slog](https://pkg.go.dev/log/slog) structured logging API.
This hook can be particularly helpful for troubleshooting and debugging; simply attach it at the global, client or invocation level and ensure your log level is set to "debug".

##### Usage example

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/open-feature/go-sdk/openfeature"
    "github.com/open-feature/go-sdk/openfeature/hooks"
    "github.com/open-feature/go-sdk/openfeature/memprovider"
)

func main() {
    // Register an in-memory provider with no flags
    openfeature.SetNamedProviderAndWait("example", memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{}))

    // Configure slog
    handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
    logger := slog.New(handler)

    // Register a logging hook globally to run on all evaluations
    loggingHook := hooks.NewLoggingHook(false, logger)
    openfeature.AddHooks(loggingHook)

    // Create a new client
    client := openfeature.NewClient("example")

    // Attempt to evaluate a flag that doesn't exist
    _ = client.Boolean(context.TODO(), "not-exist", true, openfeature.EvaluationContext{})
}
```

###### Output

```sh
{"time":"2025-06-03T10:49:23.100783-04:00","level":"DEBUG","msg":"Before stage","domain":"example","provider_name":"InMemoryProvider","flag_key":"not-exist","default_value":true,"stage":"before"}
{"time":"2025-06-03T10:49:23.101037-04:00","level":"ERROR","msg":"Error stage","domain":"example","provider_name":"InMemoryProvider","flag_key":"not-exist","default_value":true,"error_message":"error code: FLAG_NOT_FOUND: flag for key not-exist not found","stage":"error"}
```

See [hooks](#hooks) for more information on configuring hooks.

### Domains

Clients can be assigned to a domain. A domain is a logical identifier that can be used to associate clients with a particular provider. If a domain has no associated provider, the default provider is used.

```go
import "github.com/open-feature/go-sdk/openfeature"

// Registering the default provider
openfeature.SetProviderAndWait(NewLocalProvider())
// Registering a named provider
openfeature.SetNamedProvider("clientForCache", NewCachedProvider())

// A Client backed by default provider
clientWithDefault := openfeature.NewClient("")
// A Client backed by NewCachedProvider
clientForCache := openfeature.NewClient("clientForCache")
```

### Eventing

Events allow you to react to state changes in the provider or underlying flag management system, such as flag definition changes, provider readiness, or error conditions.
Initialization events (`PROVIDER_READY` on success, `PROVIDER_ERROR` on failure) are dispatched for every provider.
Some providers support additional events, such as `PROVIDER_CONFIGURATION_CHANGED`.

Please refer to the documentation of the provider you're using to see what events are supported.

```go
import "github.com/open-feature/go-sdk/openfeature"

...
var readyHandlerCallback = func(details openfeature.EventDetails) {
    // callback implementation
}

// Global event handler
openfeature.AddHandler(openfeature.ProviderReady, &readyHandlerCallback)

...

var providerErrorCallback = func(details openfeature.EventDetails) {
    // callback implementation
}

client := openfeature.NewClient("clientName")

// Client event handler
client.AddHandler(openfeature.ProviderError, &providerErrorCallback)
```

### Shutdown

The OpenFeature API provides a close function to perform a cleanup of all registered providers.
This should only be called when your application is in the process of shutting down.

```go
import "github.com/open-feature/go-sdk/openfeature"

openfeature.Shutdown()
```


### Transaction Context Propagation

Transaction context is a container for transaction-specific evaluation context (e.g. user id, user agent, IP).
Transaction context can be set where specific data is available (e.g. an auth service or request handler), and by using the transaction context propagator, it will automatically be applied to all flag evaluations within a transaction (e.g. a request or thread).

```go
import "github.com/open-feature/go-sdk/openfeature"

// set the TransactionContext
ctx := openfeature.WithTransactionContext(context.TODO(), openfeature.EvaluationContext{})

// get the TransactionContext from a context
ec := openfeature.TransactionContext(ctx)

// merge an EvaluationContext with the existing TransactionContext, preferring
// the context that is passed to MergeTransactionContext
tCtx := openfeature.MergeTransactionContext(ctx, openfeature.EvaluationContext{})

// use TransactionContext in a flag evaluation
client.BooleanValue(tCtx, ....)
```

## Extending

### Develop a provider

To develop a provider, you need to create a new project and include the OpenFeature SDK as a dependency.
This can be a new repository or included in [the existing contrib repository](https://github.com/open-feature/go-sdk-contrib) available under the OpenFeature organization.
You’ll then need to write the provider by implementing the `FeatureProvider` interface exported by the OpenFeature SDK.

```go
package myfeatureprovider

import (
  "context"
  "github.com/open-feature/go-sdk/openfeature"
)

// MyFeatureProvider implements the FeatureProvider interface and provides functions for evaluating flags
type MyFeatureProvider struct{}

// Required: Methods below implements openfeature.FeatureProvider interface
// This is the core interface implementation required from a provider
// Metadata returns the metadata of the provider
func (i MyFeatureProvider) Metadata() openfeature.Metadata {
  return openfeature.Metadata{
    Name: "MyFeatureProvider",
  }
}

// Hooks returns a collection of openfeature.Hook defined by this provider
func (i MyFeatureProvider) Hooks() []openfeature.Hook {
  // Hooks that should be included with the provider
  return []openfeature.Hook{}
}
// BooleanEvaluation returns a boolean flag
func (i MyFeatureProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
  // code to evaluate boolean
}

// StringEvaluation returns a string flag
func (i MyFeatureProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
  // code to evaluate string
}

// FloatEvaluation returns a float flag
func (i MyFeatureProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
  // code to evaluate float
}

// IntEvaluation returns an int flag
func (i MyFeatureProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
  // code to evaluate int
}

// ObjectEvaluation returns an object flag
func (i MyFeatureProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
  // code to evaluate object
}

// Optional: openfeature.StateHandler implementation
// Providers can opt-in for initialization & shutdown behavior by implementing this interface

// Init holds initialization logic of the provider
func (i MyFeatureProvider) Init(evaluationContext openfeature.EvaluationContext) error {
  // code to initialize your provider
}

// Shutdown define the shutdown operation of the provider
func (i MyFeatureProvider) Shutdown() {
  // code to shutdown your provider
}

// Optional: openfeature.EventHandler implementation.
// Providers can opt-in for eventing support by implementing this interface

// EventChannel returns the event channel of this provider
func (i MyFeatureProvider) EventChannel() <-chan openfeature.Event {
  // expose event channel from this provider. SDK listen to this channel and invoke event handlers
}
```

> Built a new provider? [Let us know](https://github.com/open-feature/openfeature.dev/issues/new?assignees=&labels=provider&projects=&template=document-provider.yaml&title=%5BProvider%5D%3A+) so we can add it to the docs!

### Develop a hook

To develop a hook, you need to create a new project and include the OpenFeature SDK as a dependency.
This can be a new repository or included in [the existing contrib repository](https://github.com/open-feature/go-sdk-contrib) available under the OpenFeature organization.
Implement your own hook by conforming to the [Hook interface](./pkg/openfeature/hooks.go).
To satisfy the interface, all methods (`Before`/`After`/`Finally`/`Error`) need to be defined.
To avoid defining empty functions make use of the `UnimplementedHook` struct (which already implements all the empty functions).

```go
import (
  "context"
  "github.com/open-feature/go-sdk/openfeature"
)

type MyHook struct {
  openfeature.UnimplementedHook
}

// overrides UnimplementedHook's Error function
func (h MyHook) Error(context context.Context, hookContext openfeature.HookContext, err error, hookHints openfeature.HookHints) {
  // code that runs when there's an error during a flag evaluation
}
```

> Built a new hook? [Let us know](https://github.com/open-feature/openfeature.dev/issues/new?assignees=&labels=hook&projects=&template=document-hook.yaml&title=%5BHook%5D%3A+) so we can add it to the docs!

## Testing

The SDK provides a `NewTestProvider` which allows you to set flags for the scope of a test.
The `TestProvider` is thread-safe and can be used in tests that run in parallel.

Call `testProvider.UsingFlags(t, tt.flags)` to set flags for a test, and clean them up with `testProvider.Cleanup()`

```go
import (
  "github.com/open-feature/go-sdk/openfeature"
  "github.com/open-feature/go-sdk/openfeature/testing"
)

testProvider := NewTestProvider()
err := openfeature.GetApiInstance().SetProviderAndWait(testProvider)
if err != nil {
  t.Errorf("unable to set provider")
}

// configure flags for this test suite
tests := map[string]struct {
  flags map[string]memprovider.InMemoryFlag
  want  bool
}{
  "test when flag is true": {
    flags: map[string]memprovider.InMemoryFlag{
      "my_flag": {
        State:          memprovider.Enabled,
        DefaultVariant: "on",
        Variants: map[string]any{
          "on": true,
        },
      },
    },
    want: true,
  },
  "test when flag is false": {
    flags: map[string]memprovider.InMemoryFlag{
      "my_flag": {
        State:          memprovider.Enabled,
        DefaultVariant: "off",
        Variants: map[string]any{
          "off": false,
        },
      },
    },
    want: false,
  },
}

for name, tt := range tests {
  tt := tt
  name := name
  t.Run(name, func(t *testing.T) {

    // be sure to clean up your flags
    defer testProvider.Cleanup()
    testProvider.UsingFlags(t, tt.flags)

    // your code under test
    got := functionUnderTest()

    if got != tt.want {
      t.Fatalf("uh oh, value is not as expected: got %v, want %v", got, tt.want)
    }
  })
}
```

### Mocks

Mocks are also available for testing purposes for all interfaces within the OpenFeature SDK. These are primarily
intended for internal use for testing the SDK, but have been exported to ease the testing burden for any extensions
or custom components (e.g. hooks & providers). These mocks are not include in builds by default. The build tag 
`testtools` must be used to have the mocks included in builds.

<!-- x-hide-in-docs-start -->
## ⭐️ Support the project

- Give this repo a ⭐️!
- Follow us on social media:
  - Twitter: [@openfeature](https://twitter.com/openfeature)
  - LinkedIn: [OpenFeature](https://www.linkedin.com/company/openfeature/)
- Join us on [Slack](https://cloud-native.slack.com/archives/C0344AANLA1)
- For more, check out our [community page](https://openfeature.dev/community/)

## 🤝 Contributing

Interested in contributing? Great, we'd love your help! To get started, take a look at the [CONTRIBUTING](CONTRIBUTING.md) guide.

### Thanks to everyone that has already contributed

<a href="https://github.com/open-feature/go-sdk/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=open-feature/go-sdk" alt="Pictures of the folks who have contributed to the project" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
<!-- x-hide-in-docs-end -->
