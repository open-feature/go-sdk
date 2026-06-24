package openfeature_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
)

// TestEvaluationAPINoUnexpectedExports ensures the [EvaluationAPI] public
// surface only exposes the intended methods. A failure here means the external
// consumer contract was broken — methods were added, removed, or renamed.
func TestEvaluationAPINoUnexpectedExports(t *testing.T) {
	wantMethods := []string{
		"AddHandler",
		"AddHooks",
		"NewClient",
		"RemoveHandler",
		"SetEvaluationContext",
		"SetProvider",
		"SetProviderAndWait",
		"Shutdown",
	}

	apiType := reflect.TypeFor[*openfeature.EvaluationAPI]()
	gotMethods := make([]string, apiType.NumMethod())
	for i := range apiType.NumMethod() {
		gotMethods[i] = apiType.Method(i).Name
	}
	slices.Sort(gotMethods)

	if !slices.Equal(wantMethods, gotMethods) {
		t.Errorf("EvaluationAPI public methods mismatch:\nwant: %v\ngot:  %v", wantMethods, gotMethods)
	}
}

// TestEvaluationAPIBreakingPreventer calls every [EvaluationAPI] public method
// with valid arguments. This guards the external consumer contract at both
// compile time (must compile) and runtime (no panics). A signature change or
// broken behavior fails this test.
func TestEvaluationAPIBreakingPreventer(t *testing.T) {
	api := isolated.NewAPI()
	ctx := t.Context()

	// SetProvider — default provider (async init)
	if err := api.SetProvider(ctx, openfeature.NoopProvider{}); err != nil {
		t.Errorf("SetProvider: unexpected error: %v", err)
	}

	// SetProviderWithDomain — named provider via WithDomain option (async init)
	if err := api.SetProvider(ctx, openfeature.NoopProvider{}, openfeature.WithDomain("domain-a")); err != nil {
		t.Errorf("SetProvider + WithDomain: unexpected error: %v", err)
	}

	// SetProviderAndWait — default provider (sync init)
	if err := api.SetProviderAndWait(ctx, openfeature.NoopProvider{}); err != nil {
		t.Errorf("SetProviderAndWait: unexpected error: %v", err)
	}

	// SetProviderAndWaitWithDomain — named provider via WithDomain option (sync init)
	if err := api.SetProviderAndWait(ctx, openfeature.NoopProvider{}, openfeature.WithDomain("domain-b")); err != nil {
		t.Errorf("SetProviderAndWait + WithDomain: unexpected error: %v", err)
	}

	// NewClient — default client (no domain)
	if client := api.NewClient(); client == nil {
		t.Error("NewClient: expected non-nil client")
	}

	// NewClientWithDomain — named client via WithDomain option
	if client := api.NewClient(openfeature.WithDomain("domain-a")); client == nil {
		t.Error("NewClient + WithDomain: expected non-nil client")
	}

	// SetEvaluationContext — global evaluation context
	api.SetEvaluationContext(openfeature.NewEvaluationContext("my-key", nil))

	// AddHooks — append API-level hooks
	api.AddHooks(openfeature.UnimplementedHook{})

	// AddHandler / RemoveHandler — API-level eventing
	handler := func(details openfeature.EventDetails) {}
	api.AddHandler(openfeature.ProviderReady, &handler)
	api.RemoveHandler(openfeature.ProviderReady, &handler)

	// Shutdown — clean up the isolated instance
	if err := api.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}
