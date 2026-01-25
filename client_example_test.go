package openfeature_test

import (
	"context"
	"fmt"
	"log"

	"go.openfeature.dev/openfeature/v2"
)

func ExampleNewClient() {
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))
	fmt.Printf("Client Domain: %s", client.Metadata().Domain())
	// Output: Client Domain: example-client
}

func ExampleClient_Boolean() {
	if err := openfeature.SetProviderAndWait(context.TODO(), openfeature.NoopProvider{}, openfeature.WithDomain("example-client")); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	if client.Boolean(ctx, "myflag", true, openfeature.EvaluationContext{}) {
		fmt.Println("myflag is true")
	} else {
		fmt.Println("myflag is false")
	}

	// Output: myflag is true
}

func ExampleClient_String() {
	if err := openfeature.SetProviderAndWait(context.TODO(), openfeature.NoopProvider{}, openfeature.WithDomain("example-client")); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	fmt.Println(client.String(ctx, "myflag", "default", openfeature.EvaluationContext{}))

	// Output: default
}

func ExampleClient_Float() {
	if err := openfeature.SetProviderAndWait(context.TODO(), openfeature.NoopProvider{}, openfeature.WithDomain("example-client")); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	fmt.Println(client.Float(ctx, "myflag", 0.5, openfeature.EvaluationContext{}))

	// Output: 0.5
}

func ExampleClient_Int() {
	if err := openfeature.SetProviderAndWait(context.TODO(), openfeature.NoopProvider{}, openfeature.WithDomain("example-client")); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	fmt.Println(client.Int(ctx, "myflag", 5, openfeature.EvaluationContext{}))

	// Output: 5
}

func ExampleClient_Object() {
	if err := openfeature.SetProviderAndWait(context.TODO(), openfeature.NoopProvider{}, openfeature.WithDomain("example-client")); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	fmt.Println(client.Object(ctx, "myflag", map[string]string{"foo": "bar"}, openfeature.EvaluationContext{}))

	// Output: map[foo:bar]
}

func ExampleClient_Track() {
	ctx := context.TODO()
	client := openfeature.NewClient(openfeature.WithDomain("example-client"))

	evaluationContext := openfeature.EvaluationContext{}

	// example tracking event recording that a subject reached a page associated with a business goal
	client.Track(ctx, "visited-promo-page", evaluationContext, openfeature.TrackingEventDetails{})

	// example tracking event recording that a subject performed an action associated with a business goal, with the tracking event details having a particular numeric value
	client.Track(ctx, "clicked-checkout", evaluationContext, openfeature.NewTrackingEventDetails(99.77))

	// example tracking event recording that a subject performed an action associated with a business goal, with the tracking event details having a particular numeric value
	client.Track(ctx, "clicked-checkout", evaluationContext, openfeature.NewTrackingEventDetails(99.77).Add("currencyCode", "USD"))

	// Output:
}
