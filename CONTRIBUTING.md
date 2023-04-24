## Welcome!

Thank you very much for contributing to this project. Any issues and pull requests following these guidelines are welcome.

### Code of conduct

There's [a code of conduct](https://github.com/open-feature/.github/blob/main/CODE_OF_CONDUCT.md).
TLDR: be respectful.

### Vendor specific details

Vendor specific details are intentionally not included in this module in order to be lightweight and agnostic.
If there are changes needed to enable vendor specific behaviour in code or other extension points, check out [the spec](https://github.com/open-feature/spec).

## Development

### Installation and Dependencies

Install dependencies with `go get ./...`

We value having as few runtime dependencies as possible. The addition of any dependencies requires careful consideration and review.

### Testing

Any contributions are expected to include unit tests. These can be validated with `make test` or the automated github workflow will run them on PR creation.

The go version in the `go.mod` is the currently supported version of go.

When writing a test to cover a spec requirement use the test naming convention `TestRequirement_x_y_z` where `x_y_z` is the numbering of the spec requirement (e.g. spec requirement `1.1.1` demands a test with name `TestRequirement_1_1_1`). Also include the description of the test requirement as a comment of the test.

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

## Contacting us

We hold regular meetings which you can see [here](https://github.com/open-feature/community/#meetings-and-events).

We are also present in the `#openfeature` channel in the [CNCF slack](https://slack.cncf.io/).
