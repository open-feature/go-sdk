// Package testing provides a test-aware feature flag provider for OpenFeature.
package testing

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

const testNameKey = "testName"

// NewTestProvider creates a new `TestAwareProvider`
func NewTestProvider() TestProvider {
	return TestProvider{
		providers: &sync.Map{},
	}
}

// TestProvider is the recommended way to defined flags within the scope of a test.
// It uses the InMemoryProvider, with flags scoped per test.
// Before executing a test, specify the flag values to be used for the specific test using the UsingFlags function
type TestProvider struct {
	openfeature.NoopProvider
	providers *sync.Map
}

type TestFramework = interface{ Name() string }

// UsingFlags sets flags for the scope of a test.
func (tp TestProvider) UsingFlags(test TestFramework, flags map[string]memprovider.InMemoryFlag) {
	storeGoroutineLocal(test.Name())
	tp.providers.Store(test.Name(), memprovider.NewInMemoryProvider(flags))
}

// Cleanup deletes the flags provider bound to the current test and should be executed after each test execution
// e.g. using a defer statement.
func (tp TestProvider) Cleanup() {
	tp.providers.Delete(getGoroutineLocal())
	deleteGoroutineLocal()
}

func (tp TestProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	return tp.getProvider().BooleanEvaluation(ctx, flag, defaultValue, flatCtx)
}

func (tp TestProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	return tp.getProvider().StringEvaluation(ctx, flag, defaultValue, flatCtx)
}

func (tp TestProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	return tp.getProvider().FloatEvaluation(ctx, flag, defaultValue, flatCtx)
}

func (tp TestProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	return tp.getProvider().IntEvaluation(ctx, flag, defaultValue, flatCtx)
}

func (tp TestProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return tp.getProvider().ObjectEvaluation(ctx, flag, defaultValue, flatCtx)
}

func (tp TestProvider) Hooks() []openfeature.Hook {
	return tp.NoopProvider.Hooks()
}

func (tp TestProvider) Metadata() openfeature.Metadata {
	return tp.NoopProvider.Metadata()
}

func (tp TestProvider) getProvider() openfeature.FeatureProvider {
	// Retrieve the test name from the goroutine-local storage.
	testName, ok := getGoroutineLocal().(string)
	if !ok {
		panic("unable to detect test name; be sure to call `UsingFlags`	in the scope of a test (in T.run)!")
	}

	// Load the feature provider corresponding to the test name.
	provider, ok := tp.providers.Load(testName)
	if !ok {
		panic("unable to find feature provider for given test name: " + testName)
	}

	// Assert that the loaded provider is of type openfeature.FeatureProvider.
	featureProvider, ok := provider.(openfeature.FeatureProvider)
	if !ok {
		panic("invalid type for feature provider for given test name: " + testName)
	}

	return featureProvider
}

var goroutineLocalData sync.Map

func storeGoroutineLocal(value any) {
	gID := getGoroutineID()
	goroutineLocalData.Store(fmt.Sprintf("%d_%v", gID, testNameKey), value)
}

func getGoroutineLocal() any {
	gID := getGoroutineID()
	value, _ := goroutineLocalData.Load(fmt.Sprintf("%d_%v", gID, testNameKey))
	return value
}

func deleteGoroutineLocal() {
	gID := getGoroutineID()
	goroutineLocalData.Delete(fmt.Sprintf("%d_%v", gID, testNameKey))
}

func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stackLine := string(buf[:n])
	var gID uint64
	_, err := fmt.Sscanf(stackLine, "goroutine %d ", &gID)
	if err != nil {
		panic("unable to extract GID from stack trace")
	}
	return gID
}
