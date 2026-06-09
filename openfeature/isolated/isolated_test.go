package isolated_test

import (
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
)

// TestNewAPI_ReturnsDistinctInstances verifies that each call to NewAPI
// returns a new, independent instance (spec 1.8.1).
func TestNewAPI_ReturnsDistinctInstances(t *testing.T) {
	a := isolated.NewAPI()
	b := isolated.NewAPI()
	t.Cleanup(func() {
		a.Shutdown()
		b.Shutdown()
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
	t.Cleanup(func() { a.Shutdown() })

	//nolint:staticcheck // test needs the singleton reference for comparison
	singleton, ok := openfeature.GetApiInstance().(*openfeature.EvaluationAPI)
	if !ok {
		t.Fatal("singleton is not *EvaluationAPI; cannot compare")
	}
	if a == singleton {
		t.Error("isolated.NewAPI() returned the global singleton")
	}
}
