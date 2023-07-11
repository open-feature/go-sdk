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
[![Go Reference](https://pkg.go.dev/badge/github.com/open-feature/go-sdk/pkg/openfeature.svg)](https://pkg.go.dev/github.com/open-feature/go-sdk/pkg/openfeature)
[![a](https://img.shields.io/badge/slack-%40cncf%2Fopenfeature-brightgreen?style=flat&logo=slack)](https://cloud-native.slack.com/archives/C0344AANLA1)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-feature/go-sdk)](https://goreportcard.com/report/github.com/open-feature/go-sdk)
[![codecov](https://codecov.io/gh/open-feature/go-sdk/branch/main/graph/badge.svg?token=FZ17BHNSU5)](https://codecov.io/gh/open-feature/go-sdk)
[![v0.6.0](https://img.shields.io/static/v1?label=Specification&message=v0.6.0&color=yellow)](https://github.com/open-feature/spec/tree/v0.6.0) 
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6601/badge)](https://bestpractices.coreinfrastructure.org/projects/6601)
<!-- x-hide-in-docs-start -->

## 👋 Hey there! Thanks for checking out the OpenFeature Go SDK

### What is OpenFeature?

[OpenFeature][openfeature-website] is an open standard that provides a vendor-agnostic, community-driven API for feature flagging that works with your favorite feature flag management tool.

### Why standardize feature flags?

Standardizing feature flags unifies tools and vendors behind a common interface which avoids vendor lock-in at the code level. Additionally, it offers a framework for building extensions and integrations and allows providers to focus on their unique value proposition.

<!-- x-hide-in-docs-end -->
## 🔍 Requirements

- Go 1.18+

## 📦 Installation

```shell
go get github.com/open-feature/go-sdk
```

### Software Bill of Materials (SBOM)

The release workflow generates an SBOM (using [cyclonedx](https://github.com/CycloneDX/cyclonedx-gomod)) and pushes it to the release. It can be found as an asset named `bom.json` within a release.

## 🌟 Features

- support for various backend [providers](https://openfeature.dev/docs/reference/concepts/provider)
- easy integration and extension via [hooks](https://openfeature.dev/docs/reference/concepts/hooks)
- bool, string, numeric, and object flag types
- [context-aware](https://openfeature.dev/docs/reference/concepts/evaluation-context) evaluation
- Supports [OpenFeature Events](https://openfeature.dev/specification/sections/events)

## 🚀 Usage

To configure the SDK you'll need to add a provider to the `openfeature` global singleton. From there, you can generate a `Client` which is usable by your code.
While you'll likely want a provider for your specific backend, we've provided a `NoopProvider`, which simply returns the default passed in.

```go
package main

import (
	"context"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

func main() {
	openfeature.SetProvider(openfeature.NoopProvider{})
	client := openfeature.NewClient("app")
	value, err := client.BooleanValue(
		context.Background(), "v2_enabled", false, openfeature.EvaluationContext{},
	)
}
```

A list of available providers can be found [here](https://openfeature.dev/ecosystem?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Provider&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go).

For complete documentation, visit: https://openfeature.dev/docs/category/concepts

### Context-aware evaluation

Sometimes the value of a flag must take into account some dynamic criteria about the application or user, such as the user location, IP, email address, or the location of the server.
In OpenFeature, we refer to this as [`targeting`](https://openfeature.dev/specification/glossary#targeting).
If the flag system you're using supports targeting, you can provide the input data using the `EvaluationContext`.

```go
// add a value to the global context
openfeature.SetEvaluationContext(openfeature.NewEvaluationContext(
    "foo",
    map[string]interface{}{
        "myGlobalKey":  "myGlobalValue",
    },
))

// add a value to the client context
client := openfeature.NewClient("my-app")
client.SetEvaluationContext(openfeature.NewEvaluationContext(
    "", 
    map[string]interface{}{
        "myGlobalKey":  "myGlobalValue",
    },
))

// add a value to the invocation context
evalCtx := openfeature.NewEvaluationContext(
    "",
    map[string]interface{}{
        "myInvocationKey": "myInvocationValue",
    },
)
boolValue, err := client.BooleanValue("boolFlag", false, evalCtx)
```

### Providers

To develop a provider, you need to create a new project and include the OpenFeature SDK as a dependency. This can be a new repository or included in [the existing contrib repository](https://github.com/open-feature/go-sdk-contrib) available under the OpenFeature organization. Finally, you’ll then need to write the provider itself. This can be accomplished by implementing the `FeatureProvider` interface exported by the OpenFeature SDK.

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

See [here](https://openfeature.dev/ecosystem?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Provider&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go) for a catalog of available providers.

### Hooks

Implement your own hook by conforming to the [Hook interface](./pkg/openfeature/hooks.go).

To satisfy the interface, all methods (`Before`/`After`/`Finally`/`Error`) need to be defined. To avoid defining empty functions
make use of the `UnimplementedHook` struct (which already implements all the empty functions).

```go
type MyHook struct {
  openfeature.UnimplementedHook
}

// overrides UnimplementedHook's Error function
func (h MyHook) Error(hookContext openfeature.HookContext, err error, hookHints openfeature.HookHints) {
	log.Println(err)
}
```

Register the hook at the global, client, or invocation level.

A list of available hooks can be found [here](https://openfeature.dev/ecosystem?instant_search%5BrefinementList%5D%5Btype%5D%5B0%5D=Hook&instant_search%5BrefinementList%5D%5Btechnology%5D%5B0%5D=Go).

### Logging

If not configured, the logger falls back to the standard Go log package at error level only.

In order to avoid coupling to any particular logging implementation the SDK uses the structured logging [logr](https://github.com/go-logr/logr)
API. This allows integration to any package that implements the layer between their logger and this API.
Thankfully there are already [integration implementations](https://github.com/go-logr/logr#implementations-non-exhaustive)
for many of the popular logger packages.

```go
var l logr.Logger
l = integratedlogr.New() // replace with your chosen integrator

openfeature.SetLogger(l) // set the logger at global level

c := openfeature.NewClient("log").WithLogger(l) // set the logger at client level

```

[logr](https://github.com/go-logr/logr) uses incremental verbosity levels (akin to named levels but in integer form).
The SDK logs `info` at level `0` and `debug` at level `1`. Errors are always logged.

### Named clients:

Clients can be given a name. A name is a logical identifier which can be used to associate clients with a particular provider. 
If a name has no associated provider, clients with that name use the global provider.

```go
import "github.com/open-feature/go-sdk/pkg/openfeature"

...

// Registering the default provider
openfeature.SetProvider(NewLocalProvider())
// Registering a named provider
openfeature.SetNamedProvider("clientForCache", NewCachedProvider())

// A Client backed by default provider
clientWithDefault := openfeature.NewClient("")
// A Client backed by NewCachedProvider
clientForCache := openfeature.NewClient("clientForCache")
```

### Events:

Events allow you to react to state changes in the provider or underlying flag management system, such as flag definition changes, provider readiness, or error conditions.
Initialization events (PROVIDER_READY on success, PROVIDER_ERROR on failure) are dispatched for every provider.
Some providers support additional events, such as PROVIDER_CONFIGURATION_CHANGED.

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

### Shutdown: 

The OpenFeature API provides a close function to perform a cleanup of all registered providers.
This should only be called when your application is in the process of shutting down.

```go
import "github.com/open-feature/go-sdk/pkg/openfeature"

...

openfeature.Shutdown()
```

### Named clients:

Clients can be given a name. A name is a logical identifier which can be used to associate clients with a particular provider. 
If a name has no associated provider, clients with that name use the global provider.

```go
import "github.com/open-feature/go-sdk/pkg/openfeature"

...

// Registering the default provider
openfeature.SetProvider(NewLocalProvider())
// Registering a named provider
openfeature.SetNamedProvider("clientForCache", NewCachedProvider())

// A Client backed by default provider
clientWithDefault := openfeature.NewClient("")
// A Client backed by NewCachedProvider
clientForCache := openfeature.NewClient("clientForCache")
```

### Events:

Events allow you to react to state changes in the provider or underlying flag management system, such as flag definition changes, provider readiness, or error conditions.
Initialization events (PROVIDER_READY on success, PROVIDER_ERROR on failure) are dispatched for every provider.
Some providers support additional events, such as PROVIDER_CONFIGURATION_CHANGED.

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

### Shutdown: 

The OpenFeature API provides a close function to perform a cleanup of all registered providers.
This should only be called when your application is in the process of shutting down.

```go
import "github.com/open-feature/go-sdk/pkg/openfeature"

...

openfeature.Shutdown()
```

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

## 📜 License

[Apache License 2.0](LICENSE)

[openfeature-website]: https://openfeature.dev
<!-- x-hide-in-docs-end -->
