package multi

import (
	"context"
	"iter"
	"sync"

	of "github.com/open-feature/go-sdk/openfeature"
)

// runModeFn is a function type that defines how flag evaluations are executed across multiple providers.
// It returns an iterator that yields provider names and their corresponding resolution details.
type runModeFn[T FlagTypes] func(ctx context.Context, providers []NamedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) iter.Seq2[string, of.GenericResolutionDetail[T]]

// runModeParallel evaluates a flag across multiple providers concurrently.
// It launches a goroutine for each provider and yields results as they complete.
// Evaluation stops early if the iterator consumer stops consuming results.
func runModeParallel[T FlagTypes](ctx context.Context, providers []NamedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) iter.Seq2[string, of.GenericResolutionDetail[T]] {
	type namedResult struct {
		name       string
		resolution of.GenericResolutionDetail[T]
	}
	return iter.Seq2[string, of.GenericResolutionDetail[T]](func(yield func(string, of.GenericResolutionDetail[T]) bool) {
		resolutions := make(chan *namedResult, len(providers))
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			var wg sync.WaitGroup
			for _, provider := range providers {
				wg.Add(1)
				go func(provider NamedProvider) {
					defer wg.Done()
					resolution := Evaluate(ctx, provider, flag, defaultValue, flatCtx)
					select {
					case <-ctx.Done():
						return
					default:
						resolutions <- &namedResult{name: provider.Name(), resolution: resolution}
					}
				}(provider)
			}
			wg.Wait()
			close(resolutions)
		}()

		for result := range resolutions {
			if !yield(result.name, result.resolution) {
				return
			}
		}
	})
}

// runModeSequential evaluates a flag across multiple providers one at a time in order.
// It yields each provider's result sequentially and stops if the iterator consumer stops consuming results.
func runModeSequential[T FlagTypes](ctx context.Context, providers []NamedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) iter.Seq2[string, of.GenericResolutionDetail[T]] {
	return iter.Seq2[string, of.GenericResolutionDetail[T]](func(yield func(string, of.GenericResolutionDetail[T]) bool) {
		for _, provider := range providers {
			resolution := Evaluate(ctx, provider, flag, defaultValue, flatCtx)
			if !yield(provider.Name(), resolution) {
				return
			}
		}
	})
}
