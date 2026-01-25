package openfeature_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.openfeature.dev/openfeature/v2"
)

// ExampleSetProvider demonstrates asynchronous provider setup with timeout control.
func ExampleSetProvider() {
	// Create a test provider for demonstration
	provider := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := openfeature.SetProvider(ctx, provider)
	if err != nil {
		log.Printf("Failed to start provider setup: %v", err)
		return
	}

	// Provider continues initializing in background
	fmt.Println("Provider setup initiated")
	// Output: Provider setup initiated
}

// ExampleSetProviderAndWait demonstrates synchronous provider setup with error handling.
func ExampleSetProviderAndWait() {
	// Create a test provider for demonstration
	provider := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := openfeature.SetProviderAndWait(ctx, provider)
	if err != nil {
		log.Printf("Provider initialization failed: %v", err)
		return
	}

	// Provider is now ready to use
	fmt.Println("Provider is ready")
	// Output: Provider is ready
}

// ExampleSetProvider_withDomain multi-tenant provider setup.
func ExampleSetProvider_withDomain() {
	// Create test providers for different services
	userProvider := &openfeature.NoopProvider{}
	billingProvider := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := openfeature.SetProvider(ctx, userProvider, openfeature.WithDomain("user-service"))
	if err != nil {
		log.Printf("Failed to setup user service provider: %v", err)
		return
	}

	err = openfeature.SetProvider(ctx, billingProvider, openfeature.WithDomain("billing-service"))
	if err != nil {
		log.Printf("Failed to setup billing service provider: %v", err)
		return
	}

	// Create clients for different domains
	userClient := openfeature.NewClient(openfeature.WithDomain("user-service"))
	billingClient := openfeature.NewClient(openfeature.WithDomain("billing-service"))

	fmt.Printf("User client domain: %s\n", userClient.Metadata().Domain())
	fmt.Printf("Billing client domain: %s\n", billingClient.Metadata().Domain())
	// Output: User client domain: user-service
	// Billing client domain: billing-service
}

// ExampleSetProviderAndWait_withDomain critical service provider setup.
func ExampleSetProviderAndWait_withDomain() {
	// Create a test provider for demonstration
	criticalProvider := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for critical providers to be ready
	err := openfeature.SetProviderAndWait(ctx, criticalProvider, openfeature.WithDomain("critical-service"))
	if err != nil {
		log.Printf("Critical provider failed to initialize: %v", err)
		return
	}

	// Now safe to use the client
	client := openfeature.NewClient(openfeature.WithDomain("critical-service"))
	enabled := client.Boolean(context.Background(), "feature-x", false, openfeature.EvaluationContext{})

	fmt.Printf("Critical service ready, feature-x enabled: %v\n", enabled)
	// Output: Critical service ready, feature-x enabled: false
}

// ExampleStateHandler demonstrates how context-aware shutdown works automatically.
func ExampleStateHandler() {
	// Context-aware providers automatically use ShutdownWithContext when replaced
	provider1 := &openfeature.NoopProvider{}
	provider2 := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set first provider
	err := openfeature.SetProviderAndWait(ctx, provider1)
	if err != nil {
		log.Printf("Provider setup failed: %v", err)
		return
	}

	// Replace with second provider - this triggers context-aware shutdown of provider1 if it supports it
	err = openfeature.SetProviderAndWait(ctx, provider2)
	if err != nil {
		log.Printf("Provider replacement failed: %v", err)
		return
	}

	fmt.Println("Context-aware provider lifecycle completed")
	// Output: Context-aware provider lifecycle completed
}

// ExampleShutdown demonstrates graceful application shutdown with timeout control.
func ExampleShutdown() {
	// Set up providers
	provider1 := &openfeature.NoopProvider{}
	provider2 := &openfeature.NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up multiple providers
	err := openfeature.SetProviderAndWait(ctx, provider1)
	if err != nil {
		log.Printf("Provider setup failed: %v", err)
		return
	}

	err = openfeature.SetProviderAndWait(ctx, provider2, openfeature.WithDomain("service-a"))
	if err != nil {
		log.Printf("Named provider setup failed: %v", err)
		return
	}

	// Application is running...

	// When application is shutting down, use context-aware shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	err = openfeature.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Shutdown completed with errors: %v", err)
	} else {
		fmt.Println("All providers shut down successfully")
	}

	// Output: All providers shut down successfully
}
