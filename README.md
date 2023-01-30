# OpenFeature SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/open-feature/go-sdk/pkg/openfeature.svg)](https://pkg.go.dev/github.com/open-feature/go-sdk/pkg/openfeature)
[![a](https://img.shields.io/badge/slack-%40cncf%2Fopenfeature-brightgreen?style=flat&logo=slack)](https://cloud-native.slack.com/archives/C0344AANLA1)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-feature/go-sdk)](https://goreportcard.com/report/github.com/open-feature/go-sdk)
[![codecov](https://codecov.io/gh/open-feature/go-sdk/branch/main/graph/badge.svg?token=FZ17BHNSU5)](https://codecov.io/gh/open-feature/go-sdk)
[![v0.5.1](https://img.shields.io/static/v1?label=Specification&message=v0.5.1&color=yellow)](https://github.com/open-feature/spec/tree/v0.5.1)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6601/badge)](https://bestpractices.coreinfrastructure.org/projects/6601)

This is the Go implementation of [OpenFeature](https://openfeature.dev), a vendor-agnostic abstraction library for evaluating feature flags.

We support multiple data types for flags (floats, integers, strings, booleans, objects) as well as hooks, which can alter the lifecycle of a flag evaluation.

## Installation

```shell
go get github.com/open-feature/go-sdk
```

## Usage

To configure the sdk you'll need to add a provider to the `openfeature` global singleton. From there, you can generate a `Client` which is usable by your code.
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

A list of available providers can be found [here](https://docs.openfeature.dev/docs/reference/technologies/server/go).

For complete documentation, visit: https://docs.openfeature.dev/docs/category/concepts

### Hooks

Implement your own hook by conforming to the [Hook interface](./pkg/openfeature/hooks.go).

To satisfy the interface all methods (`Before`/`After`/`Finally`/`Error`) need to be defined. To avoid defining empty functions
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

Register the hook at global, client or invocation level.

A list of available hooks can be found [here](https://docs.openfeature.dev/docs/reference/technologies/server/go).

## Configuration

### Logging

If not configured, the logger falls back to the standard Go log package at error level only.

In order to avoid coupling to any particular logging implementation the sdk uses the structured logging [logr](https://github.com/go-logr/logr)
API. This allows integration to any package that implements the layer between their logger and this API.
Thankfully there is already [integration implementations](https://github.com/go-logr/logr#implementations-non-exhaustive)
for many of the popular logger packages.

```go
var l logr.Logger
l = integratedlogr.New() // replace with your chosen integrator

openfeature.SetLogger(l) // set the logger at global level

c := openfeature.NewClient("log").WithLogger(l) // set the logger at client level

```

[logr](https://github.com/go-logr/logr) uses incremental verbosity levels (akin to named levels but in integer form).
The sdk logs `info` at level `0` and `debug` at level `1`. Errors are always logged.

## Development

### Installation and Dependencies

Install dependencies with `go get ./...`

We value having as few runtime dependencies as possible. The addition of any dependencies requires careful consideration and review.

### Testing

#### Unit tests

Run unit tests with `make test`.

#### Integration tests

The continuous integration runs a set of [gherkin integration tests](https://github.com/open-feature/test-harness/blob/main/features) using the [flagd provider](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/flagd), [flagd](https://github.com/open-feature/flagd) and [the flagd test module](https://github.com/open-feature/go-sdk-contrib/tree/main/tests/flagd).
If you'd like to run them locally, first pull the `test-harness` git submodule
```
git submodule update --init --recursive
```
then start the flagd testbed with 
```
docker run -p 8013:8013 -v $PWD/test-harness/testing-flags.json:/testing-flags.json ghcr.io/open-feature/flagd-testbed:latest
```
 and finally run
```
make integration-test
```

#### Fuzzing

[Go supports fuzzing natively as of 1.18](https://go.dev/security/fuzz/).
The fuzzing suite is implemented as an integration of `go-sdk` with the [flagd provider](https://github.com/open-feature/go-sdk-contrib/tree/main/providers/flagd) and [flagd](https://github.com/open-feature/flagd).
The fuzzing tests are found in [./integration/evaluation_fuzz_test.go](./integration/evaluation_fuzz_test.go), they are dependent on the flagd testbed running, you can start it with
```
docker run -p 8013:8013 ghcr.io/open-feature/flagd-testbed:latest
```
then, to execute a fuzzing test, run the following
```
go test -fuzz=FuzzBooleanEvaluation ./integration/evaluation_fuzz_test.go
```
substituting the name of the fuzz as appropriate.

### Releases

This repo uses Release Please to release packages. Release Please sets up a running PR that tracks all changes for the library components, and maintains the versions according to conventional commits, generated when PRs are merged. When Release Please's running PR is merged, any changed artifacts are published.

#### SBOM

The release workflow generates a SBOM (using [cyclonedx](https://github.com/CycloneDX/cyclonedx-gomod)) and pushes it to the release. It can be found as an asset named `bom.json` within a release.

## Contacting us

We hold regular meetings which you can see [here](https://github.com/open-feature/community/#meetings-and-events).

We are also present in the `#openfeature` channel in the [CNCF slack](https://slack.cncf.io/).

## Contributors

Thanks so much to our contributors.

<a href="https://github.com/open-feature/go-sdk/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=open-feature/go-sdk" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## License

Apache License 2.0
