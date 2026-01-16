# AGENTS.md - OpenFeature Go SDK

This document provides guidelines for agents working on this codebase.

## Build Commands

```bash
# Run unit tests (short mode, skips slow e2e tests)
make test
# Equivalent: go test --short -tags testtools -cover -timeout 1m ./...

# Run end-to-end tests (includes e2e tests with race detection)
make e2e-test
# Equivalent: git submodule update --init --recursive && go test -tags testtools -race -cover -timeout 1m ./e2e/...

# Run linter
make lint
# Equivalent: golangci-lint run ./...

# Auto-fix linting issues
make fix
# Equivalent: golangci-lint run ./... --fix

# Generate mocks (requires mockgen)
make mockgen

# Generate API documentation
make docs
```

### Running a Single Test

```bash
# Run specific test in current package
go test -tags testtools -run TestMyFunction ./...

# Run test in specific file/package
go test -tags testtools -run TestMyFunction ./path/to/package/...

# Run with verbose output
go test -tags testtools -v -run TestMyFunction ./...
```

## Developer Certificate of Origin (DCO)

This project requires all commits to include a Signed-off-by line to certify adherence to the [Developer Certificate of Origin](https://developercertificate.org/).

### Signing Off Commits

Add the `-s` flag to your commit command:

```bash
git commit -s -m "your commit message"
```

This adds a line like `Signed-off-by: Name <email@example.com>` to your commit message.

### Amending to Add DCO

If you forgot to sign off a commit:

```bash
git commit --amend -s
```

### Verifying DCO Status

Check if your commits are signed:

```bash
git log --pretty=format:%H,%s | head -5
```

Or use the DCO bot to check PRs (the bot will flag unsigned commits).

## Conventional Commits

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools

### Examples

```
feat(provider): add new flag evaluation method

fix(api): resolve nil pointer in evaluation context

docs: update installation instructions
```

## Code Style Guidelines

### Formatting & Imports

- Use `gofumpt` for formatting (stricter than gofmt)
- Use `gci` for import sorting (stdlib first, then external)
- Run `make fix` before committing to auto-format

### Linting

This project uses golangci-lint with the following linters:

- **staticcheck**: All checks enabled (SAxxxx rules)
- **errcheck**: Error checking enabled
- **govet**: All checks except fieldalignment and shadow
- **usetesting**: Enforces testing best practices
- **nolintlint**: Requires specific linter annotations (no unused nolint directives)
- **modernize**: Suggests modern Go idioms and language features
- **copyloopvar**: Detects loop variable copying bugs
- **intrange**: Suggests using `for range` for integer loops

Excluded rules:

- G101 (hardcoded credentials detection - too noisy)
- G404 (weak random number generation - used in tests)

### Naming Conventions

- **Variables**: camelCase (e.g., `clientMetadata`, `evalCtx`)
- **Exported Types/Functions**: PascalCase (e.g., `Client`, `NewClient`)
- **Interfaces**: PascalCase, typically I-prefixed (e.g., `IClient`, `IFeatureProvider`) or role-based (e.g., `FeatureProvider`)
- **Constants**: PascalCase

### Error Handling

- Use `fmt.Errorf` with `%w` for error wrapping
- Return errors with descriptive messages (lowercase, no punctuation)
- Use sentinel errors from this package when applicable
- Check errors explicitly: `if err != nil { ... }`

### Context Usage

- Context is the first parameter in functions that need it
- Use `context.Background()` for top-level entry points
- Pass context through call chains (don't store in structs)
- Support cancellation where operations may be slow

### Mutex Patterns

- Use `sync.RWMutex` for read-heavy workloads
- `RLock/RUnlock` for reads, `Lock/Unlock` for writes
- Always defer unlock after locking
- Never hold locks across function calls or goroutines

### Documentation

- Document all exported types and functions with // comments
- Use proper Go doc comments starting with type/function name
- Example: `// Client implements the behaviour required of an openfeature client`
- Include usage examples in example test files (\*\_example_test.go)

### Go Examples

- Go examples are a critical part of documentation and testing
- Place in `*_example_test.go` files next to the code they document
- Naming convention: `Example<Type>_<Method>` for method examples, `Example<Type>` for type examples
- Example function signature: `func ExampleClient_Boolean()`
- Use `// Output:` comment to specify expected output (verified by `go test`)
- Use `fmt.Println` for output to verify example behavior
- External examples use package `openfeature_test` to test public API usage
- Examples serve dual purpose: test correctness and generate godoc examples
- Run examples with: `go test -tags testtools -run Example ./...`

### Testing

- Use standard `testing` package with `testtools` build tag
- Use `t.Cleanup()` for cleanup operations
- Use table-driven tests where appropriate
- Place tests in `*_test.go` files next to implementation
- E2E tests are in `/e2e` directory with `testtools` tag

### Modern Go Guidelines

#### Standard Library Packages

- **`slices` package**: Use modern slice operations instead of manual implementations
- **`maps` package**: Use modern map operations

#### Modern Language Features

- **For loop variable scoping**: Variables declared in `for` loops are created anew for each iteration
- **Ranging over integers**: Use `for i := range n` instead of `for i := 0; i < n; i++`
- **Generics**: Use type parameters for reusable, type-safe functions and types
- **Type inference**: Leverage automatic type inference with `:=` when possible

#### Modernization Tools

- **`go fix`**: Automatically modernize code to use current Go features

  ```bash
  go fix ./...
  ```

- **`modernize` analyzer**: Integrated in golangci-lint (enabled by default)
  - Run `make fix` to apply modernize suggestions automatically

#### Best Practices

- **Slice initialization**: Prefer `var s []T` over `s := []T{}` or `make([]T, 0)` for nil slices

  ```go
  // Good
  var results []string
  results = append(results, "item")

  // Avoid
  results := []string{}
  ```

- **Preallocation**: Preallocate slice capacity when size is known

  ```go
  // Good
  items := make([]Item, 0, len(source))
  for _, s := range source {
      items = append(items, Item(s))
  }
  ```

- **Error handling**: Use `fmt.Errorf` with `%w` for wrapping errors
- **Context**: Pass context as first parameter, support cancellation
- **Concurrency**: Use `sync/errgroup` for managing goroutines with error handling
- **Testing**: Use table-driven tests with `t.Cleanup()` for resource cleanup

### General

- Go version: 1.25
- Module: `go.openfeature.dev/openfeature/v2`
- Mock generation: `make mockgen`
- Run `make test && make lint` before committing
