Welcome!

There are a few things to consider before contributing to the sdk.

Firstly, there's [a code of conduct](https://github.com/open-feature/.github/blob/main/CODE_OF_CONDUCT.md).
TLDR: be respectful.

Vendor specific details are intentionally not included in this module in order to be lightweight and agnostic.
If there are changes needed to enable vendor specific behaviour in code or other extension points, check out [the spec](https://github.com/open-feature/spec).

Any contributions are expected to include unit tests. These can be validated with `make test` or the automated github workflow will run them on PR creation.

The go version in the `go.mod` is the currently supported version of go.

When writing a test to cover a spec requirement use the test naming convention `TestRequirement_x_y_z` where `x_y_z` is the numbering of the spec requirement (e.g. spec requirement `1.1.1` demands a test with name `TestRequirement_1_1_1`). Also include the description of the test requirement as a comment of the test.

Thanks! Issues and pull requests following these guidelines are welcome.
