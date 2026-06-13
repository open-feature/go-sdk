package isolated_test

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
)

func ExampleNewAPI() {
	ctx := context.Background()

	// Create an isolated API instance independent of the global singleton.
	api := isolated.NewAPI()
	defer func() { _ = api.Shutdown(ctx) }()

	// Set a default provider (waits for initialization).
	if err := api.SetProviderAndWait(ctx, openfeature.NoopProvider{}); err != nil {
		fmt.Println("SetProviderAndWait failed:", err)
		return
	}

	// Set a domain-scoped provider with WithDomain.
	if err := api.SetProviderAndWait(ctx, openfeature.NoopProvider{}, openfeature.WithDomain("my-domain")); err != nil {
		fmt.Println("SetProviderAndWait + WithDomain failed:", err)
		return
	}

	// Obtain a client bound to the domain-scoped provider.
	client := api.NewClient(openfeature.WithDomain("my-domain"))

	// Evaluate a boolean flag.
	value := client.Boolean(ctx, "my-flag", true, openfeature.EvaluationContext{})
	fmt.Printf("flag value: %v", value)
	// Output: flag value: true
}
