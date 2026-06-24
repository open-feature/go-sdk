package isolated_test

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk/openfeature/internal/factory"
	"github.com/open-feature/go-sdk/openfeature/isolated"
)

// TestNewAPI_ReturnsDistinctInstances verifies that each call to NewAPI
// returns a new, independent instance (spec 1.8.1).
func TestNewAPI_ReturnsDistinctInstances(t *testing.T) {
	a := isolated.NewAPI()
	b := isolated.NewAPI()
	t.Cleanup(func() {
		ctx := context.Background() //nolint:usetesting
		_ = a.Shutdown(ctx)
		_ = b.Shutdown(ctx)
	})

	if a == nil || b == nil {
		t.Fatal("NewAPI returned nil")
	}
	if a == b {
		t.Error("two NewAPI calls returned the same instance")
	}
}

// TestNewAPI_NotSameAsSingleton verifies that the isolated factory does not
// return the global singleton.
func TestNewAPI_NotSameAsSingleton(t *testing.T) {
	a := isolated.NewAPI()
	t.Cleanup(func() {
		ctx := context.Background() //nolint:usetesting
		_ = a.Shutdown(ctx)
	})

	if a == factory.CurrentAPI() {
		t.Error("isolated.NewAPI() returned the global singleton")
	}
}
