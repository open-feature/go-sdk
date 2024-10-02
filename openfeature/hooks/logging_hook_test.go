package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"log/slog"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

func TestLoggingHookInitializesCorrectly(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	hook := LoggingHook{
		includeEvaluationContext: true,
		logger:                   logger,
	}

	if hook.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, hook.logger)
	}

	if !hook.includeEvaluationContext {
		t.Errorf("Expected includeEvaluationContext to be true, got %v", hook.includeEvaluationContext)
	}
}

func TestLoggingHookHandlesNilLoggerGracefully(t *testing.T) {
	hook := LoggingHook{
		includeEvaluationContext: false,
		logger:                   nil,
	}

	if hook.logger != nil {
		t.Errorf("Expected logger to be nil, got %v", hook.logger)
	}

	if hook.includeEvaluationContext {
		t.Errorf("Expected includeEvaluationContext to be false, got %v", hook.includeEvaluationContext)
	}
}

func TestLoggingHookLogsMessagesAsExpected(t *testing.T) {
	// handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	// handler := &testHandler{}
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	hook := LoggingHook{
		includeEvaluationContext: false,
		logger:                   logger,
	}

	// Check if resultMap contains all key-value pairs from expected
	testLoggingHookLogsMessagesAsExpected(hook, logger, t, buf)
}

func TestLoggingHookLogsMessagesAsExpectedIncludeEvaluationContext(t *testing.T) {
	// handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	// handler := &testHandler{}
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	hook := LoggingHook{
		includeEvaluationContext: true,
		logger:                   logger,
	}

	// Check if resultMap contains all key-value pairs from expected
	testLoggingHookLogsMessagesAsExpected(hook, logger, t, buf)
}

func testLoggingHookLogsMessagesAsExpected(hook LoggingHook, logger *slog.Logger, t *testing.T, buf bytes.Buffer) {
	if hook.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, hook.logger)
	}

	if !hook.includeEvaluationContext {
		t.Errorf("Expected includeEvaluationContext to be true, got %v", hook.includeEvaluationContext)
	}

	memoryProvider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
		"boolFlag": {
			Key:            "boolFlag",
			State:          memprovider.Enabled,
			DefaultVariant: "true",
			Variants: map[string]interface{}{
				"true":  true,
				"false": false,
			},
			ContextEvaluator: nil,
		},
	})

	openfeature.SetProviderAndWait(memoryProvider)
	openfeature.AddHooks(&hook)
	client := openfeature.NewClient("test-app")

	t.Run("test boolean success", func(t *testing.T) {
		res, err := client.BooleanValue(
			context.Background(),
			"boolFlag",
			true,
			openfeature.NewEvaluationContext(
				"target1",
				map[string]interface{}{
					"color": "green",
				},
			),
		)
		if err != nil {
			t.Error("expected nil error")
		}
		if res != true {
			t.Errorf("incorect evaluation, expected %t, got %t", true, res)
		}

		slog.Info("logs", "handler.logs", buf.Bytes())

		var ms map[string]map[string]any = make(map[string]map[string]any)
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}

			slog.Info("object", "m", m)

			ms[m["msg"].(string)] = m
		}

		var expected = map[string]map[string]any{
			"Before stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"flag_key":      "boolFlag",
			},
		}

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
	})
}
