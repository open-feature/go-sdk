package e2e_test

import (
	"context"
	"strings"
	"testing"
	"time"

	flagd "github.com/open-feature/go-sdk-contrib/providers/flagd/pkg"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

func setupFuzzClient(f *testing.F) *openfeature.Client {
	f.Helper()

	provider := flagd.NewProvider(flagd.WithPort(8013), flagd.WithoutCache())
	err := openfeature.SetProvider(provider)
	if err != nil {
		f.Errorf("error setting up provider %v", err)
	}

	select {
	case <-provider.IsReady():
	case <-time.After(500 * time.Millisecond):
		f.Fatal("provider not ready after 500 milliseconds")
	}

	return openfeature.NewClient("fuzzing")
}

func FuzzBooleanEvaluation(f *testing.F) {
	client := setupFuzzClient(f)

	f.Add("foo", false)
	f.Add("FoO", true)
	f.Add("FoO234", false)
	f.Add("FoO2\b34", true)
	f.Fuzz(func(t *testing.T, flagKey string, defaultValue bool) {
		res, err := client.BooleanValueDetails(context.Background(), flagKey, defaultValue, openfeature.EvaluationContext{})
		if err != nil {
			if res.ErrorCode == openfeature.FlagNotFoundCode {
				return
			}
			if strings.Contains(err.Error(), string(openfeature.ParseErrorCode)) {
				return
			}
			t.Error(err)
		}
	})
}

func FuzzStringEvaluation(f *testing.F) {
	client := setupFuzzClient(f)

	f.Add("foo", "bar")
	f.Add("FoO", "BaR")
	f.Add("FoO234", "Ba1232")
	f.Add("FoO2\b34", "BaaR\b2312")
	f.Fuzz(func(t *testing.T, flagKey string, defaultValue string) {
		res, err := client.StringValueDetails(context.Background(), flagKey, defaultValue, openfeature.EvaluationContext{})
		if err != nil {
			if res.ErrorCode == openfeature.FlagNotFoundCode {
				return
			}
			if strings.Contains(err.Error(), string(openfeature.ParseErrorCode)) {
				return
			}
			t.Error(err)
		}
	})
}

func FuzzIntEvaluation(f *testing.F) {
	client := setupFuzzClient(f)

	f.Add("foo", int64(1))
	f.Add("FoO", int64(99))
	f.Add("FoO234", int64(100029))
	f.Add("FoO2\b34", int64(-1))
	f.Fuzz(func(t *testing.T, flagKey string, defaultValue int64) {
		res, err := client.IntValueDetails(context.Background(), flagKey, defaultValue, openfeature.EvaluationContext{})
		if err != nil {
			if res.ErrorCode == openfeature.FlagNotFoundCode {
				return
			}
			if strings.Contains(err.Error(), string(openfeature.ParseErrorCode)) {
				return
			}
			t.Error(err)
		}
	})
}

func FuzzFloatEvaluation(f *testing.F) {
	client := setupFuzzClient(f)

	f.Add("foo", float64(1))
	f.Add("FoO", 99.9)
	f.Add("FoO234", 0.00004)
	f.Add("FoO2\b34", -1.9203)
	f.Fuzz(func(t *testing.T, flagKey string, defaultValue float64) {
		res, err := client.FloatValueDetails(context.Background(), flagKey, defaultValue, openfeature.EvaluationContext{})
		if err != nil {
			if res.ErrorCode == openfeature.FlagNotFoundCode {
				return
			}
			if strings.Contains(err.Error(), string(openfeature.ParseErrorCode)) {
				return
			}
			t.Error(err)
		}
	})
}

func FuzzObjectEvaluation(f *testing.F) {
	client := setupFuzzClient(f)

	f.Add("foo", "{}")
	f.Add("FoO", "true")
	f.Add("FoO234", "-1.23")
	f.Add("FoO2\b34", "1")
	f.Fuzz(func(t *testing.T, flagKey string, defaultValue string) { // interface{} is not supported, using a string
		res, err := client.ObjectValueDetails(context.Background(), flagKey, defaultValue, openfeature.EvaluationContext{})
		if err != nil {
			if res.ErrorCode == openfeature.FlagNotFoundCode {
				return
			}
			if strings.Contains(err.Error(), string(openfeature.ParseErrorCode)) {
				return
			}
			t.Error(err)
		}
	})
}
