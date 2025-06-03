package openfeature_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/open-feature/go-sdk/openfeature"
)

func ExampleNewClient() {
	client := openfeature.NewClient("example-client")
	fmt.Printf("Client Domain: %s", client.Metadata().Domain())
	// Output: Client Domain: example-client
}

func ExampleClient_BooleanValue() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("example-client")
	value, err := client.BooleanValue(
		context.TODO(), "test-flag", true, openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatal("error while getting boolean value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: true
}

func ExampleClient_StringValue() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("example-client")
	value, err := client.StringValue(
		context.TODO(), "test-flag", "openfeature", openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatal("error while getting string value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: openfeature
}

func ExampleClient_FloatValue() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("example-client")
	value, err := client.FloatValue(
		context.TODO(), "test-flag", 0.55, openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatalf("error while getting float value: %v", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: 0.55
}

func ExampleClient_IntValue() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("example-client")
	value, err := client.IntValue(
		context.TODO(), "test-flag", 3, openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatalf("error while getting int value: %v", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: 3
}

func ExampleClient_ObjectValue() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("example-client")
	value, err := client.ObjectValue(
		context.TODO(), "test-flag", map[string]string{"foo": "bar"}, openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatal("error while getting object value : ", err)
	}

	str, _ := json.Marshal(value)
	fmt.Printf("test-flag value: %v", string(str))
	// Output: test-flag value: {"foo":"bar"}
}

func ExampleClient_Boolean() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	if client.Boolean(ctx, "myflag", true, openfeature.EvaluationContext{}) {
		fmt.Println("myflag is true")
	} else {
		fmt.Println("myflag is false")
	}

	// Output: myflag is true
}

func ExampleClient_String() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	fmt.Println(client.String(ctx, "myflag", "default", openfeature.EvaluationContext{}))

	// Output: default
}

func ExampleClient_Float() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	fmt.Println(client.Float(ctx, "myflag", 0.5, openfeature.EvaluationContext{}))

	// Output: 0.5
}

func ExampleClient_Int() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	fmt.Println(client.Int(ctx, "myflag", 5, openfeature.EvaluationContext{}))

	// Output: 5
}

func ExampleClient_Object() {
	if err := openfeature.SetNamedProviderAndWait("example-client", openfeature.NoopProvider{}); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	fmt.Println(client.Object(ctx, "myflag", map[string]string{"foo": "bar"}, openfeature.EvaluationContext{}))

	// Output: map[foo:bar]
}

func ExampleClient_Track() {
	ctx := context.TODO()
	client := openfeature.NewClient("example-client")

	evaluationContext := openfeature.EvaluationContext{}

	// example tracking event recording that a subject reached a page associated with a business goal
	client.Track(ctx, "visited-promo-page", evaluationContext, openfeature.TrackingEventDetails{})

	// example tracking event recording that a subject performed an action associated with a business goal, with the tracking event details having a particular numeric value
	client.Track(ctx, "clicked-checkout", evaluationContext, openfeature.NewTrackingEventDetails(99.77))

	// example tracking event recording that a subject performed an action associated with a business goal, with the tracking event details having a particular numeric value
	client.Track(ctx, "clicked-checkout", evaluationContext, openfeature.NewTrackingEventDetails(99.77).Add("currencyCode", "USD"))

	// Output:
}
