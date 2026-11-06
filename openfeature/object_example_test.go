package openfeature_test

import (
	"context"
	"fmt"
	"log"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

type RolloutConfig struct {
	Text       string `json:"text"`
	Percentage int    `json:"percentage"`
}

// This example evaluates an object flag directly into a struct.
func ExampleGetObjectValue() {
	provider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
		"rollout-config": {
			Key:            "rollout-config",
			State:          memprovider.Enabled,
			DefaultVariant: "on",
			Variants: map[string]any{
				"on": map[string]any{"text": "hello", "percentage": 25},
			},
		},
	})
	if err := openfeature.SetNamedProviderAndWait("object-example", &provider); err != nil {
		log.Fatalf("error setting up provider %v", err)
	}
	client := openfeature.NewClient("object-example")

	// The flag value is deserialized straight into RolloutConfig.
	config, err := openfeature.GetObjectValue(
		context.TODO(), client, "rollout-config", RolloutConfig{Text: "N/A", Percentage: 75}, openfeature.EvaluationContext{},
	)
	if err != nil {
		log.Fatal("error while getting object value : ", err)
	}

	fmt.Printf("text: %s, percentage: %d", config.Text, config.Percentage)
	// Output: text: hello, percentage: 25
}
