# OpenFeature SDK for Golang

[![a](https://img.shields.io/badge/slack-%40cncf%2Fopenfeature-brightgreen?style=flat&logo=slack)](https://cloud-native.slack.com/archives/C0344AANLA1)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-feature/golang-sdk)](https://goreportcard.com/report/github.com/open-feature/golang-sdk)
[![codecov](https://codecov.io/gh/open-feature/golang-sdk/branch/main/graph/badge.svg?token=FZ17BHNSU5)](https://codecov.io/gh/open-feature/golang-sdk)

This is the Golang implementation of [OpenFeature](https://openfeature.dev), a vendor-agnostic abstraction library for evaluating feature flags.

We support multiple data types for flags (floats, integers, strings, booleans, objects) as well as hooks, which can alter the lifecycle of a flag evaluation.

## Installation

```shell
go get github.com/open-feature/golang-sdk
```

## Usage

To configure the sdk you'll need to add a provider to the `openfeature` global singleton. From there, you can generate a `Client` which is usable by your code.
While you'll likely want a provider for your specific backend, we've provided a `NoopProvider`, which simply returns the default passed in.
```golang
package main

import (
	"github.com/open-feature/golang-sdk/pkg/openfeature"
)

func main() {
	openfeature.SetProvider(openfeature.NoopProvider{})
	client := openfeature.GetClient("app")
	value, err := client.BooleanValue("v2_enabled", false, nil)
}
```

## Development

### Installation and Dependencies

Install dependencies with `go get ./...`

We value having as few runtime dependencies as possible. The addition of any dependencies requires careful consideration and review.

### Testing

Run tests with `make test`.

## Contacting us
We hold regular meetings which you can see [here](https://github.com/open-feature/community/#meetings-and-events).

We are also present in the `#openfeature` channel in the [CNCF slack](https://slack.cncf.io/).

## Contributors

Thanks so much to our contributors.

<a href="https://github.com/open-feature/golang-sdk/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=open-feature/golang-sdk" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## License

Apache License 2.0
