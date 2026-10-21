package multi

import (
	"context"
	"maps"
	"testing"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type isolationHook struct {
	of.UnimplementedHook
	visit  func(string, context.Context, of.HookContext, of.HookHints)
	update *of.EvaluationContext
}

func (h isolationHook) Before(ctx context.Context, hc of.HookContext, hints of.HookHints) (*of.EvaluationContext, error) {
	h.visit("before", ctx, hc, hints)
	return h.update, nil
}

func (h isolationHook) After(ctx context.Context, hc of.HookContext, _ of.InterfaceEvaluationDetails, hints of.HookHints) error {
	h.visit("after", ctx, hc, hints)
	return nil
}

func (h isolationHook) Error(ctx context.Context, hc of.HookContext, _ error, hints of.HookHints) {
	h.visit("error", ctx, hc, hints)
}

func (h isolationHook) Finally(ctx context.Context, hc of.HookContext, _ of.InterfaceEvaluationDetails, hints of.HookHints) {
	h.visit("finally", ctx, hc, hints)
}

type isolationProvider struct {
	of.NoopProvider
	hooks []of.Hook
	fail  bool
}

func (p isolationProvider) Hooks() []of.Hook { return p.hooks }

func (p isolationProvider) StringEvaluation(_ context.Context, flag, defaultValue string, flatCtx of.FlattenedContext) of.StringResolutionDetail {
	if p.fail && flag == "outer" {
		return of.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				ResolutionError: of.NewGeneralResolutionError("evaluation failed"),
			},
		}
	}
	return of.StringResolutionDetail{Value: flatCtx[of.TargetingKey].(string)}
}

func isolationClient(t testing.TB, provider of.FeatureProvider) *of.Client {
	t.Helper()
	api := isolated.NewAPI()
	require.NoError(t, api.SetProviderAndWait(t.Context(), provider))
	t.Cleanup(func() { assert.NoError(t, api.Shutdown(context.WithoutCancel(t.Context()))) })
	return api.NewClient(of.WithDomain("isolation"))
}

func evaluateSubject(ctx context.Context, client *of.Client, subject string) (string, error) {
	return client.StringValue(ctx, subject, "default", of.NewEvaluationContext(subject, map[string]any{"subject": subject}),
		of.WithHookHints(of.NewHookHints(map[string]any{"subject": subject})))
}

func awaitIsolation(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("evaluation did not complete")
	}
}

func recordIsolation(t *testing.T, records chan<- string) isolationHook {
	return isolationHook{visit: func(stage string, _ context.Context, hc of.HookContext, hints of.HookHints) {
		assert.Equal(t, hc.FlagKey(), hc.EvaluationContext().TargetingKey())
		assert.Equal(t, hc.FlagKey(), hc.EvaluationContext().Attribute("subject"))
		assert.Equal(t, hc.FlagKey(), hints.Value("subject"))
		assert.Equal(t, "isolation", hc.ClientMetadata().Domain())
		records <- hc.FlagKey() + ":" + stage
	}}
}

func assertIsolationRecords(t *testing.T, records <-chan string, outerStage string) {
	t.Helper()
	var got []string
	for len(records) > 0 {
		got = append(got, <-records)
	}
	assert.ElementsMatch(t, []string{
		"outer:before", "outer:" + outerStage, "outer:finally",
		"inner:before", "inner:after", "inner:finally",
	}, got)
}

func TestMultiProvider_EvaluationIsolation_Nested(t *testing.T) {
	for _, stage := range []string{"before", "after", "error", "finally"} {
		t.Run(stage, func(t *testing.T) {
			records := make(chan string, 16)
			var client *of.Client
			nested := isolationHook{visit: func(current string, ctx context.Context, hc of.HookContext, _ of.HookHints) {
				if current == stage && hc.FlagKey() == "outer" {
					value, err := evaluateSubject(ctx, client, "inner")
					assert.NoError(t, err)
					assert.Equal(t, "inner", value)
				}
			}}
			mp, err := NewProvider(StrategyFirstMatch, WithProvider("echo", isolationProvider{
				hooks: []of.Hook{nested, recordIsolation(t, records)},
				fail:  stage == "error",
			}))
			require.NoError(t, err)
			client = isolationClient(t, mp)
			done := make(chan struct{})
			go func() {
				defer close(done)
				value, err := evaluateSubject(t.Context(), client, "outer")
				if stage == "error" {
					assert.Error(t, err)
					assert.Equal(t, "default", value)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, "outer", value)
				}
			}()
			awaitIsolation(t, done)
			outerStage := "after"
			if stage == "error" {
				outerStage = "error"
			}
			assertIsolationRecords(t, records, outerStage)
		})
	}
}

func TestMultiProvider_EvaluationIsolation_Concurrent(t *testing.T) {
	records := make(chan string, 16)
	outerStarted := make(chan struct{})
	innerDone := make(chan struct{})
	mp, err := NewProvider(StrategyCustom,
		WithProvider("echo", isolationProvider{hooks: []of.Hook{recordIsolation(t, records)}}),
		WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
			return func(ctx context.Context, flag string, defaultValue FlagTypes, flatCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
				if flag == "outer" {
					close(outerStarted)
					<-innerDone
				}
				return Evaluate(ctx, providers[0], flag, defaultValue, flatCtx)
			}
		}),
	)
	require.NoError(t, err)
	client := isolationClient(t, mp)
	// Both calls reuse the same caller context. Their invocation state must still be separate.
	ctx := t.Context()
	outerDone := make(chan struct{})
	go func() {
		defer close(outerDone)
		value, err := evaluateSubject(ctx, client, "outer")
		assert.NoError(t, err)
		assert.Equal(t, "outer", value)
	}()
	awaitIsolation(t, outerStarted)
	value, err := evaluateSubject(ctx, client, "inner")
	assert.NoError(t, err)
	assert.Equal(t, "inner", value)
	close(innerDone)
	awaitIsolation(t, outerDone)
	assertIsolationRecords(t, records, "after")
}

func TestMultiProvider_EvaluationIsolation_DirectTypes(t *testing.T) {
	for _, tc := range []struct {
		flagType of.Type
		value    any
	}{
		{of.Boolean, true},
		{of.String, "default"},
		{of.Float, 1.5},
		{of.Int, int64(7)},
		{of.Object, map[string]any{"enabled": true}},
	} {
		t.Run(tc.flagType.String(), func(t *testing.T) {
			var stages []string
			hook := isolationHook{visit: func(stage string, _ context.Context, hc of.HookContext, hints of.HookHints) {
				stages = append(stages, stage)
				assert.Equal(t, "flag", hc.FlagKey())
				assert.Equal(t, tc.flagType, hc.FlagType())
				assert.Equal(t, tc.value, hc.DefaultValue())
				assert.Equal(t, "direct", hc.EvaluationContext().TargetingKey())
				assert.Equal(t, "value", hc.EvaluationContext().Attribute("attribute"))
				assert.Empty(t, hc.ClientMetadata().Domain())
				assert.Equal(t, "NoopProvider", hc.ProviderMetadata().Name)
				assert.Nil(t, hints.Value("subject"))
			}}
			mp, err := NewProvider(StrategyFirstMatch, WithProvider("noop", of.NoopProvider{}, hook))
			require.NoError(t, err)
			result := Evaluate(t.Context(), mp.Providers()[0], "flag", tc.value,
				of.FlattenedContext{of.TargetingKey: "direct", "attribute": "value"})
			assert.NoError(t, result.Error())
			assert.Equal(t, tc.value, result.Value)
			assert.Equal(t, []string{"before", "after", "finally"}, stages)
		})
	}
}

func TestMultiProvider_EvaluationIsolation_StrategyContext(t *testing.T) {
	type requestKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), requestKey{}, "request"), 5*time.Second)
	defer cancel()
	globalContext := of.NewTargetlessEvaluationContext(map[string]any{"global": true})
	privateContext := of.NewTargetlessEvaluationContext(map[string]any{"private": true})
	var stages [2][]string
	makeHook := func(index int) isolationHook {
		return isolationHook{visit: func(stage string, hookCtx context.Context, hc of.HookContext, hints of.HookHints) {
			stages[index] = append(stages[index], stage)
			assert.Equal(t, "request", hookCtx.Value(requestKey{}))
			assert.Equal(t, ctx.Done(), hookCtx.Done())
			assert.Equal(t, ctx.Err(), hookCtx.Err())
			deadline, hasDeadline := hookCtx.Deadline()
			wantDeadline, wantHasDeadline := ctx.Deadline()
			assert.Equal(t, wantDeadline, deadline)
			assert.Equal(t, wantHasDeadline, hasDeadline)
			assert.Equal(t, "strategy", hc.EvaluationContext().TargetingKey())
			assert.Equal(t, true, hc.EvaluationContext().Attribute("global"))
			assert.Equal(t, "outer", hints.Value("subject"))
			assert.Equal(t, "isolation", hc.ClientMetadata().Domain())
			if index == 1 || stage == "before" {
				assert.Nil(t, hc.EvaluationContext().Attribute("private"))
			} else {
				assert.Equal(t, true, hc.EvaluationContext().Attribute("private"))
			}
		}}
	}
	firstHook := makeHook(0)
	firstHook.update = &privateContext
	inner, err := NewProvider(StrategyFirstMatch, WithProvider("first", isolationProvider{hooks: []of.Hook{firstHook}}))
	require.NoError(t, err)
	mp, err := NewProvider(StrategyCustom,
		WithGlobalHooks(isolationHook{visit: func(string, context.Context, of.HookContext, of.HookHints) {}, update: &globalContext}),
		WithProvider("nested", inner),
		WithProvider("second", isolationProvider{hooks: []of.Hook{makeHook(1)}}),
		WithCustomStrategy(func(providers []NamedProvider) StrategyFn[FlagTypes] {
			return func(ctx context.Context, flag string, defaultValue FlagTypes, flatCtx of.FlattenedContext) of.GenericResolutionDetail[FlagTypes] {
				flatCtx = maps.Clone(flatCtx)
				flatCtx[of.TargetingKey] = "strategy"
				first := Evaluate(ctx, providers[0], flag, defaultValue, flatCtx)
				assert.Equal(t, "strategy", first.Value)
				return Evaluate(ctx, providers[1], flag, defaultValue, flatCtx)
			}
		}),
	)
	require.NoError(t, err)
	client := isolationClient(t, mp)
	for _, cancelled := range []bool{false, true} {
		if cancelled {
			cancel()
		}
		stages = [2][]string{}
		value, err := evaluateSubject(ctx, client, "outer")
		require.NoError(t, err)
		assert.Equal(t, "strategy", value)
		for _, got := range stages {
			assert.Equal(t, []string{"before", "after", "finally"}, got)
		}
	}
}

func BenchmarkMultiProviderEvaluationIsolation(b *testing.B) {
	for _, useMulti := range []bool{false, true} {
		name := "client"
		var provider of.FeatureProvider = of.NoopProvider{}
		if useMulti {
			name = "multi"
			var err error
			provider, err = NewProvider(StrategyFirstMatch, WithProvider("noop", provider, of.UnimplementedHook{}))
			require.NoError(b, err)
		}
		b.Run(name, func(b *testing.B) {
			client := isolationClient(b, provider)
			ctx := b.Context()
			evalCtx := of.NewEvaluationContext("subject", nil)
			b.ReportAllocs()
			for b.Loop() {
				_, _ = client.BooleanValue(ctx, "flag", true, evalCtx)
			}
		})
	}
}
