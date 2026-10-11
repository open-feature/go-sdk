## end-to-end tests

This folder contains e2e tests for go-sdk. Tests written here rely on `InMemoryProvider` to perform flag evaluations.
Some tests require the `spec` Git submodule and use behaviour driven tests defined with [Gherkin](https://cucumber.io/docs/gherkin/reference/) syntax.
The feature files come from [Appendix B](https://github.com/open-feature/spec/blob/main/specification/appendix-b-gherkin-suites.md) of the specification, at `spec/specification/assets/gherkin`.

