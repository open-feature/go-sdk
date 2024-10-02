package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
	"golang.org/x/exp/slog"
)

// LoggingHook initializes correctly with valid parameters
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

// LoggingHook handles nil logger gracefully
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

type testHandler struct {
	logs []string
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer
	slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}).Handle(ctx, r)
	h.logs = append(h.logs, buf.String())
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func TestLoggingHookLogsMessagesAsExpected(t *testing.T) {
	// handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	// handler := &testHandler{}
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

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

		// slog.Info("ms", "ms", ms)

		var expected = map[string]map[string]any{
			"Before stage": {
				"provider_name": "InMemoryProvider",
				"domain":        "test-app",
				"flag_key":      "boolFlag",
			},
		}

		// Check if resultMap contains all key-value pairs from expected
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
					_, exists := resultInnerMap[EVALUATION_CONTEXT_KEY]
					if !exists {
						t.Errorf("Inner key %s not found in resultMap[%s]", EVALUATION_CONTEXT_KEY, key)
					}
				}
			}
		}
	})
}

// LoggingHook includes evaluation context when specified
func TestLoggingHookIncludesEvaluationContextWhenSpecified(t *testing.T) {
	// Initialize LoggingHook with includeEvaluationContext set to true

	// Mock slog.Logger

	// Call the function that includes evaluation context

	// Assert the evaluation context is included in the logging
}

// LoggingHook does not include evaluation context when not specified
func TestLoggingHookDoesNotIncludeEvaluationContextWhenNotSpecified(t *testing.T) {
	// Initialize LoggingHook with includeEvaluationContext set to false

	// Mock slog.Logger

	// Call the function that does not include evaluation context

	// Assert the evaluation context is not included in the logging
}

// func TestLoggingHookEmptyLogger(t *testing.T) {
// 	// Initialize LoggingHook with empty logger
// 	hook := LoggingHook{
// 		includeEvaluationContext: true,
// 		logger:                   nil,
// 	}

// 	// Perform the test logic here
// 	// Assert that the hook handles empty logger gracefully
// }

// func TestLoggingHookInvalidLoggerConfig(t *testing.T) {
// 	// Initialize LoggingHook with invalid logger configurations
// 	hook := LoggingHook{
// 		includeEvaluationContext: false,
// 		logger:                   &slog.Logger{}, // Invalid logger configuration
// 	}

// 	// Perform the test logic here
// 	// Assert that the hook handles invalid logger configurations
// }
