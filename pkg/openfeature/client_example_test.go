package openfeature_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/open-feature/golang-sdk/pkg/openfeature"
)

func ExampleNewClient() {
	client := openfeature.NewClient("example-client")
	fmt.Printf("Client Name: %s", client.Metadata().Name())
	// Output: Client Name: example-client
}

func ExampleClient_BooleanValue() {
	client := openfeature.NewClient("example-client")
	value, err := client.BooleanValue(
		"test-flag", true, openfeature.EvaluationContext{}, openfeature.EvaluationOptions{},
	)
	if err != nil {
		log.Fatal("error while getting boolean value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: true
}

func ExampleClient_StringValue() {
	client := openfeature.NewClient("example-client")
	value, err := client.StringValue(
		"test-flag", "openfeature", openfeature.EvaluationContext{}, openfeature.EvaluationOptions{},
	)
	if err != nil {
		log.Fatal("error while getting string value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: openfeature
}

func ExampleClient_NumberValue() {
	client := openfeature.NewClient("example-client")
	value, err := client.NumberValue(
		"test-flag", 0.55, openfeature.EvaluationContext{}, openfeature.EvaluationOptions{},
	)
	if err != nil {
		log.Fatal("error while getting number value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: 0.55
}

func ExampleClient_ObjectValue() {
	client := openfeature.NewClient("example-client")
	value, err := client.ObjectValue(
		"test-flag", map[string]string{"foo": "bar"}, openfeature.EvaluationContext{}, openfeature.EvaluationOptions{},
	)
	if err != nil {
		log.Fatal("error while getting object value : ", err)
	}

	str, _ := json.Marshal(value)
	fmt.Printf("test-flag value: %v", string(str))
	// Output: test-flag value: {"foo":"bar"}
}
