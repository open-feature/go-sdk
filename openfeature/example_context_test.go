package openfeature_test

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleSetProviderWithContext demonstrates asynchronous provider setup with timeout control.
func ExampleSetProviderWithContext() {
	// Create a test provider for demonstration
	provider := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := SetProviderWithContext(ctx, provider)
	if err != nil {
		log.Printf("Failed to start provider setup: %v", err)
		return
	}

	// Provider continues initializing in background
	fmt.Println("Provider setup initiated")
	// Output: Provider setup initiated
}

// ExampleSetProviderWithContextAndWait demonstrates synchronous provider setup with error handling.
func ExampleSetProviderWithContextAndWait() {
	// Create a test provider for demonstration
	provider := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := SetProviderWithContextAndWait(ctx, provider)
	if err != nil {
		log.Printf("Provider initialization failed: %v", err)
		return
	}

	// Provider is now ready to use
	fmt.Println("Provider is ready")
	// Output: Provider is ready
}

// ExampleSetNamedProviderWithContext demonstrates multi-tenant provider setup.
func ExampleSetNamedProviderWithContext() {
	// Create test providers for different services
	userProvider := &NoopProvider{}
	billingProvider := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := SetNamedProviderWithContext(ctx, "user-service", userProvider)
	if err != nil {
		log.Printf("Failed to setup user service provider: %v", err)
		return
	}

	err = SetNamedProviderWithContext(ctx, "billing-service", billingProvider)
	if err != nil {
		log.Printf("Failed to setup billing service provider: %v", err)
		return
	}

	// Create clients for different domains
	userClient := NewClient("user-service")
	billingClient := NewClient("billing-service")

	fmt.Printf("User client domain: %s\n", userClient.domain)
	fmt.Printf("Billing client domain: %s\n", billingClient.domain)
	// Output: User client domain: user-service
	// Billing client domain: billing-service
}

// ExampleSetNamedProviderWithContextAndWait demonstrates critical service provider setup.
func ExampleSetNamedProviderWithContextAndWait() {
	// Create a test provider for demonstration
	criticalProvider := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for critical providers to be ready
	err := SetNamedProviderWithContextAndWait(ctx, "critical-service", criticalProvider)
	if err != nil {
		log.Printf("Critical provider failed to initialize: %v", err)
		return
	}

	// Now safe to use the client
	client := NewClient("critical-service")
	enabled, _ := client.BooleanValue(context.Background(), "feature-x", false, EvaluationContext{})

	fmt.Printf("Critical service ready, feature-x enabled: %v\n", enabled)
	// Output: Critical service ready, feature-x enabled: false
}

// ExampleContextAwareStateHandler demonstrates how context-aware shutdown works automatically.
func ExampleContextAwareStateHandler() {
	// Context-aware providers automatically use ShutdownWithContext when replaced
	provider1 := &NoopProvider{}
	provider2 := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set first provider
	err := SetProviderWithContextAndWait(ctx, provider1)
	if err != nil {
		log.Printf("Provider setup failed: %v", err)
		return
	}

	// Replace with second provider - this triggers context-aware shutdown of provider1 if it supports it
	err = SetProviderWithContextAndWait(ctx, provider2)
	if err != nil {
		log.Printf("Provider replacement failed: %v", err)
		return
	}

	fmt.Println("Context-aware provider lifecycle completed")
	// Output: Context-aware provider lifecycle completed
}

// ExampleShutdownWithContext demonstrates graceful application shutdown with timeout control.
func ExampleShutdownWithContext() {
	// Set up providers
	provider1 := &NoopProvider{}
	provider2 := &NoopProvider{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up multiple providers
	err := SetProviderWithContextAndWait(ctx, provider1)
	if err != nil {
		log.Printf("Provider setup failed: %v", err)
		return
	}

	err = SetNamedProviderWithContextAndWait(ctx, "service-a", provider2)
	if err != nil {
		log.Printf("Named provider setup failed: %v", err)
		return
	}

	// Application is running...

	// When application is shutting down, use context-aware shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	err = ShutdownWithContext(shutdownCtx)
	if err != nil {
		log.Printf("Shutdown completed with errors: %v", err)
	} else {
		fmt.Println("All providers shut down successfully")
	}

	// Output: All providers shut down successfully
}
