package openfeaturetest

import (
	"context"
	"fmt"
	"github.com/open-feature/go-sdk/openfeature"
	"runtime"
	"sync"
	"testing"
)

const testNameKey = "testName"

// NewTestAwareProvider creates a new `TestAwareProvider`
func NewTestAwareProvider() TestAwareProvider {
	return TestAwareProvider{
		providers: &sync.Map{},
	}
}

// TestAwareProvider can be used in parallel unit tests. It holds a map of unit test name to `openfeature.FeatureProvider`s.
// Before executing the test, specify the actual (in memory) provider that's going to  be used for the specific test using the
// `SetProvider` method.
type TestAwareProvider struct {
	openfeature.NoopProvider
	providers *sync.Map
}

// SetProvider sets a given `FeatureProvider` for a given test.
func (tp TestAwareProvider) SetProvider(test *testing.T, fp openfeature.FeatureProvider) {
	storeGoroutineLocal(test.Name())
	tp.providers.Store(test.Name(), fp)
}

// Cleanup deletes the test provider bound to the current test and should be executed after each test execution
// e.g. using a defer statement.
func (tp TestAwareProvider) Cleanup() {
	tp.providers.Delete(getGoroutineLocal())
	deleteGoroutineLocal()
}

func (tp TestAwareProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	return tp.getProvider().BooleanEvaluation(ctx, flag, defaultValue, flCtx)
}

func (tp TestAwareProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	return tp.getProvider().StringEvaluation(ctx, flag, defaultValue, flCtx)
}

func (tp TestAwareProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	return tp.getProvider().FloatEvaluation(ctx, flag, defaultValue, flCtx)
}

func (tp TestAwareProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	return tp.getProvider().IntEvaluation(ctx, flag, defaultValue, flCtx)
}

func (tp TestAwareProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, flCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return tp.getProvider().ObjectEvaluation(ctx, flag, defaultValue, flCtx)
}

func (tp TestAwareProvider) Hooks() []openfeature.Hook {
	return tp.NoopProvider.Hooks()
}

func (tp TestAwareProvider) Metadata() openfeature.Metadata {
	return tp.NoopProvider.Metadata()
}

func (tp TestAwareProvider) getProvider() openfeature.FeatureProvider {
	// Retrieve the test name from the goroutine-local storage.
	testName, ok := getGoroutineLocal().(string)
	if !ok {
		panic("unable to detect test name")
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

func storeGoroutineLocal(value interface{}) {
	gID := getGoroutineID()
	goroutineLocalData.Store(fmt.Sprintf("%d_%v", gID, testNameKey), value)
}

func getGoroutineLocal() interface{} {
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
