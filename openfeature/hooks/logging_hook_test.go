package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

func TestCreateLoggingHookWithDefaultLoggerAndContextInclusion(t *testing.T) {
	hook, err := NewLoggingHook(true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hook == nil {
		t.Fatal("expected a valid LoggingHook, got nil")
	}
}

func TestLoggingHookInitializesCorrectly(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	hook, err := NewCustomLoggingHook(true, logger)
	if err != nil {
		t.Error("expected no error")
	}

	if hook.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, hook.logger)
	}

	if !hook.includeEvaluationContext {
		t.Errorf("Expected includeEvaluationContext to be true, got %v", hook.includeEvaluationContext)
	}
}

func TestLoggingHookHandlesNilLoggerGracefully(t *testing.T) {
	hook, err := NewCustomLoggingHook(false, nil)
	if err != nil {
		t.Error("expected no error")
	}

	if hook.logger != nil {
		t.Errorf("Expected logger to be nil, got %v", hook.logger)
	}

	if hook.includeEvaluationContext {
		t.Errorf("Expected includeEvaluationContext to be false, got %v", hook.includeEvaluationContext)
	}
}

func TestLoggingHookLogsMessagesAsExpected(t *testing.T) {
	buf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	hook, err := NewCustomLoggingHook(false, logger)
	if err != nil {
		t.Error("expected no error")
	}

	// Check if resultMap contains all key-value pairs from expected
	testLoggingHookLogsMessagesAsExpected(*hook, logger, t, buf)
}

func TestLoggingHookLogsMessagesAsExpectedIncludeEvaluationContext(t *testing.T) {
	var buf = new(bytes.Buffer)
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	hook, err := NewCustomLoggingHook(true, logger)
	if err != nil {
		t.Error("expected no error")
	}

	// Check if resultMap contains all key-value pairs from expected
	testLoggingHookLogsMessagesAsExpected(*hook, logger, t, buf)
}

func testLoggingHookLogsMessagesAsExpected(hook LoggingHook, logger *slog.Logger, t *testing.T, buf *bytes.Buffer) {
	if hook.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, hook.logger)
	}

	memoryProvider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          memprovider.Enabled,
			DefaultVariant: "true",
			Variants: map[string]any{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	err := openfeature.SetNamedProviderAndWait("test-app", memoryProvider)
	if err != nil {
		t.Error("error setting provider", err)
	}
	openfeature.AddHooks(&hook)
	client := openfeature.GetApiInstance().GetNamedClient("test-app")

	t.Run("test boolean success", func(t *testing.T) {
		res, err := client.BooleanValue(
			context.Background(),
			"boolFlag",
			false,
			openfeature.NewEvaluationContext(
				"target1",
				map[string]any{
					"color": "green",
				},
			),
		)
		if err != nil {
			t.Error("expected nil error")
		}
		if res != true {
			t.Errorf("incorrect evaluation, expected %t, got %t", true, res)
		}

		ms := prepareOutput(buf, t)

		expected := map[string]map[string]any{
			"Before stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"stage":         "before",
			},
			"After stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"flag_key":      "boolFlag",
				"stage":         "after",
			},
		}

		compare(expected, ms, t, hook)
	})

	t.Run("test boolean error", func(t *testing.T) {
		res, err := client.BooleanValue(
			context.Background(),
			"non-existing",
			false,
			openfeature.NewEvaluationContext(
				"target1",
				map[string]any{
					"color": "green",
				},
			),
		)
		if err == nil {
			t.Error("expected error")
		}
		if res != false {
			t.Errorf("incorrect evaluation, expected %t, got %t", false, res)
		}

		ms := prepareOutput(buf, t)

		expected := map[string]map[string]any{
			"Before stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"stage":         "before",
			},
			"Error stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"flag_key":      "non-existing",
				"stage":         "error",
			},
		}

		compare(expected, ms, t, hook)
	})
}

func prepareOutput(buf *bytes.Buffer, t *testing.T) map[string]map[string]any {
	var ms = make(map[string]map[string]any)
	for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatal(err)
		}
		ms[m["msg"].(string)] = m
	}
	return ms
}

func compare(expected map[string]map[string]any, ms map[string]map[string]any, t *testing.T, hook LoggingHook) {
	for key, expectedInnerMap := range expected {
		resultInnerMap, exists := ms[key]
		if !exists {
			t.Errorf("Key %s not found in resultMap", key)
			continue
		}
		for innerKey, expectedValue := range expectedInnerMap {
			resultValue, exists := resultInnerMap[innerKey]
			if !exists {
				t.Errorf("Inner key %s not found in resultMap[%s]", innerKey, key)
				continue
			}
			if resultValue != expectedValue {
				t.Errorf("Value for resultMap[%s][%s] does not match. Expected %v, got %v", key, innerKey, expectedValue, resultValue)
			}

			if hook.includeEvaluationContext {
				evaluationContext, exists := resultInnerMap[EVALUATION_CONTEXT_KEY]
				if !exists {
					t.Errorf("Inner key %s not found in resultMap[%s]", EVALUATION_CONTEXT_KEY, key)
				}
				attributes, attributesExists := evaluationContext.(map[string]any)["Attributes"]
				if !attributesExists {
					t.Errorf("attributes do not exist")
				}
				color, colorExists := attributes.(map[string]any)["color"]
				if !colorExists {
					t.Errorf("color not exist")
				}
				if color != "green" {
					t.Errorf("expected green color in evaluationContext")
				}
				if evaluationContext.(map[string]any)["TargetingKey"] != "target1" {
					t.Errorf("expected TargetingKey in evaluationContext")
				}
			}
		}
	}
}
