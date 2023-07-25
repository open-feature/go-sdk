<!-- markdownlint-disable MD033 -->
<!-- x-hide-in-docs-start -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/white/openfeature-horizontal-white.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/black/openfeature-horizontal-black.svg">
    <img align="center" alt="OpenFeature Logo">
  </picture>
</p>

<h2 align="center">OpenFeature Go SDK</h2>

<!-- x-hide-in-docs-end -->
<!-- Add repo status only if it's WIP or Concept -->
[![Go package](https://pkg.go.dev/badge/github.com/open-feature/go-sdk/pkg/openfeature.svg)](https://pkg.go.dev/github.com/open-feature/go-sdk/pkg/openfeature)
[![Specification](https://img.shields.io/static/v1?label=specification&message=v0.6.0&color=yellow)](https://github.com/open-feature/spec/tree/v0.6.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-feature/go-sdk)](https://goreportcard.com/report/github.com/open-feature/go-sdk)
[![codecov](https://codecov.io/gh/open-feature/go-sdk/branch/main/graph/badge.svg?token=FZ17BHNSU5)](https://codecov.io/gh/open-feature/go-sdk)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6601/badge)](https://bestpractices.coreinfrastructure.org/projects/6601)
<!-- x-hide-in-docs-start -->

[OpenFeature](https://openfeature.dev) is an open standard that provides a vendor-agnostic, community-driven API for feature flagging that works with your favorite feature flag management tool.

<!-- x-hide-in-docs-end -->

## 🚀 Quick start

### Requirements

- Go 1.18+


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
    "github.com/open-feature/go-sdk/pkg/openfeature"
)

func main() {
    // Register your feature flag provider
    openfeature.SetProvider(openfeature.NoopProvider{})
    // Create a new client
    client := openfeature.NewClient("app")
    // Evaluate your feature flag
    v2Enabled, _ := client.BooleanValue(
        context.Background(), "v2_enabled", true, openfeature.EvaluationContext{},
    )
    // Use the returned flag value
    if v2Enabled {
        fmt.Println("v2 is enabled")
    }
}
```

Try this example in the [Go Playground](https://go.dev/play/p/3v6jbaGGldA).

### API Reference

See [here](https://pkg.go.dev/github.com/open-feature/go-sdk/pkg/openfeature) for the complete API documentation.

## 🌟 Features

| Status | Features                                         | Description                                                                                                |
| ------ | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| ✅      | [Providers](#providers)                            | Common interface across tools and vendors, avoiding vendor lock-in at the code level.                      |
| ✅      | [Context-aware evaluation](#context-aware-evaluation) | Context-aware evaluation using `EvaluationContext` for dynamic criteria based on application or user data. |
| ✅      | [Hooks](#hooks)                          | Extend the OpenFeature SDK's functionality by accessing the flag evaluation life-cycle.                                 |
| ✅      | [Logging](#logging)                   | Integrate with popular logging packages.                                                                   |
| ✅      | [Named clients](#named-clients)                 | Utilize multiple providers in a single application.                                 |
| ✅      | [Eventing](#eventing)                              | Listen for state changes in the provider or flag management system and react on the events.                |
| ✅      | [Shutdown](#shutdown)                | Gracefully clean up a provider during application shutdown.             |
| ✅      | [Extending](#extending)       | Extend OpenFeature with custom providers and hooks to support your exact use case                        |

<sub>Implemented: ✅ | Partially implemented: ⚠️ | Not implemented yet: ❌</sub>

### Providers

[Providers](https://openfeature.dev/docs/reference/concepts/provider) are an abstraction between a flag management system and the OpenFeature SDK.
Look [here](https://openfeature.dev/ecosystem?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Provider&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go) for a complete list of available providers.
If the provider you're looking for hasn't been created yet, see the [develop a provider](#develop-a-provider) section to learn how to build it yourself.

Once you've added a provider as a dependency, it can be registered with OpenFeature like this:

```go
openfeature.SetProvider(MyProvider{})
```

In some situations, it may be beneficial to register multiple providers in the same application.
This is possible using [named clients](#named-clients), which is covered in more details below.

### Context-aware evaluation

Sometimes, the value of a flag must consider some dynamic criteria about the application or user, such as the user's location, IP, email address, or the server's location.
In OpenFeature, we refer to this as [`targeting`](https://openfeature.dev/specification/glossary#targeting).
If the flag management system you're using supports targeting, you can provide the input data using the `EvaluationContext`.

```go
// set a value to the global context
openfeature.SetEvaluationContext(openfeature.NewEvaluationContext(
    "",
    map[string]interface{}{
        "region":  "us-east-1-iah-1a",
    },
))

// set a value to the client context
client := openfeature.NewClient("my-app")
client.SetEvaluationContext(openfeature.NewEvaluationContext(
    "", 
    map[string]interface{}{
        "version":  "1.4.6",
    },
))

// set a value to the invocation context
evalCtx := openfeature.NewEvaluationContext(
    "user-123",
    map[string]interface{}{
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
    context.Background(), "boolFlag", false, openfeature.EvaluationContext{}, WithHooks(ExampleInvocationHook{}),
)
```

### Logging

The standard Go log package is used by default to show error logs.
This can be overridden using the structured logging, [logr](https://github.com/go-logr/logr) API, allowing integration to any package.
There are already [integration implementations](https://github.com/go-logr/logr#implementations-non-exhaustive) for many of the popular logger packages.

```go
var l logr.Logger
l = integratedlogr.New() // replace with your chosen integrator

openfeature.SetLogger(l) // set the logger at global level

c := openfeature.NewClient("log").WithLogger(l) // set the logger at client level
```

[logr](https://github.com/go-logr/logr) uses incremental verbosity levels (akin to named levels but in integer form).
The SDK logs `info` at level `0` and `debug` at level `1`. Errors are always logged.

### Named clients

Clients can be given a name.
A name is a logical identifier which can be used to associate clients with a particular provider.
If a name has no associated provider, the global provider is used.

```go
import "github.com/open-feature/go-sdk/pkg/openfeature"

// Registering the default provider
openfeature.SetProvider(NewLocalProvider())
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
import "github.com/open-feature/go-sdk/pkg/openfeature"

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
import "github.com/open-feature/go-sdk/pkg/openfeature"

openfeature.Shutdown()
```

## Extend

### Develop a provider

To develop a provider, you need to create a new project and include the OpenFeature SDK as a dependency.
This can be a new repository or included in [the existing contrib repository](https://github.com/open-feature/go-sdk-contrib) available under the OpenFeature organization.
You’ll then need to write the provider by implementing the `FeatureProvider` interface exported by the OpenFeature SDK.

```go
package provider

// MyFeatureProvider implements the FeatureProvider interface and provides functions for evaluating flags
type MyFeatureProvider struct{}

// Metadata returns the metadata of the provider
func (e MyFeatureProvider) Metadata() Metadata {
    return Metadata{Name: "MyFeatureProvider"}
}

// BooleanEvaluation returns a boolean flag
func (e MyFeatureProvider) BooleanEvaluation(flag string, defaultValue bool, evalCtx EvaluationContext) BoolResolutionDetail {
    // code to evaluate boolean
}

// StringEvaluation returns a string flag
func (e MyFeatureProvider) StringEvaluation(flag string, defaultValue string, evalCtx EvaluationContext) StringResolutionDetail {
    // code to evaluate string
}

// FloatEvaluation returns a float flag
func (e MyFeatureProvider) FloatEvaluation(flag string, defaultValue float64, evalCtx EvaluationContext) FloatResolutionDetail {
    // code to evaluate float
}

// IntEvaluation returns an int flag
func (e MyFeatureProvider) IntEvaluation(flag string, defaultValue int64, evalCtx EvaluationContext) IntResolutionDetail {
    // code to evaluate int
}

// ObjectEvaluation returns an object flag
func (e MyFeatureProvider) ObjectEvaluation(flag string, defaultValue interface{}, evalCtx EvaluationContext) ResolutionDetail {
    // code to evaluate object
}

// Hooks returns hooks
func (e MyFeatureProvider) Hooks() []Hook {
    // code to retrieve hooks
}
```

TODO: Extend the example to include initialize, ready, events, and shutdown.

> Built a new provider? [Let us know](https://github.com/open-feature/openfeature.dev/issues/new?assignees=&labels=provider&projects=&template=document-provider.yaml&title=%5BProvider%5D%3A+) so we can add it to the docs!

### Develop a hook

To develop a hook, you need to create a new project and include the OpenFeature SDK as a dependency.
This can be a new repository or included in [the existing contrib repository](https://github.com/open-feature/go-sdk-contrib) available under the OpenFeature organization.
Implement your own hook by conforming to the [Hook interface](./pkg/openfeature/hooks.go).
To satisfy the interface, all methods (`Before`/`After`/`Finally`/`Error`) need to be defined.
To avoid defining empty functions make use of the `UnimplementedHook` struct (which already implements all the empty functions).

```go
type MyHook struct {
  openfeature.UnimplementedHook
}

// overrides UnimplementedHook's Error function
func (h MyHook) Error(hookContext openfeature.HookContext, err error, hookHints openfeature.HookHints) {
	log.Println(err)
}
```

> Built a new hook? [Let us know](https://github.com/open-feature/openfeature.dev/issues/new?assignees=&labels=hook&projects=&template=document-hook.yaml&title=%5BHook%5D%3A+) so we can add it to the docs!

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