package multi

import (
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	imp "github.com/open-feature/go-sdk/openfeature/memprovider"
	"github.com/stretchr/testify/require"
)

func TestMultiProvider_TrackConcurrentWithShutdown(t *testing.T) {
	provider, err := NewProvider(StrategyFirstMatch,
		WithProvider("memory", imp.NewInMemoryProvider(map[string]imp.InMemoryFlag{})))
	require.NoError(t, err)
	ctx := t.Context()
	require.NoError(t, provider.InitWithContext(ctx, of.EvaluationContext{}))

	started := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		close(started)
		for range 10000 {
			provider.Track(ctx, "event", of.EvaluationContext{}, of.TrackingEventDetails{})
		}
	}()
	<-started
	require.NoError(t, provider.ShutdownWithContext(ctx))
	<-finished
}
