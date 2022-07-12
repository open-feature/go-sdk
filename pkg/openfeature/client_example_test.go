package openfeature_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/open-feature/golang-sdk/pkg/openfeature"
)

func ExampleGetClient() {
	client := openfeature.GetClient("example-client")
	fmt.Printf("Client Name: %s", client.Metadata().Name())
	// Output: Client Name: example-client
}

func ExampleClient_GetBooleanValue() {
	client := openfeature.GetClient("example-client")
	value, err := client.GetBooleanValue("test-flag", true, nil)
	if err != nil {
		log.Fatal("error while getting boolean value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: true
}

func ExampleClient_GetStringValue() {
	client := openfeature.GetClient("example-client")
	value, err := client.GetStringValue("test-flag", "openfeature", nil)
	if err != nil {
		log.Fatal("error while getting string value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: openfeature
}

func ExampleClient_GetNumberValue() {
	client := openfeature.GetClient("example-client")
	value, err := client.GetNumberValue("test-flag", 0.55, nil)
	if err != nil {
		log.Fatal("error while getting number value : ", err)
	}

	fmt.Printf("test-flag value: %v", value)
	// Output: test-flag value: 0.55
}

func ExampleClient_GetObjectValue() {
	client := openfeature.GetClient("example-client")
	value, err := client.GetObjectValue("test-flag", map[string]string{"foo": "bar"}, nil)
	if err != nil {
		log.Fatal("error while getting object value : ", err)
	}

	str, _ := json.Marshal(value)
	fmt.Printf("test-flag value: %v", string(str))
	// Output: test-flag value: {"foo":"bar"}
}
