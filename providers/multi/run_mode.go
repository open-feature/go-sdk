package multi

import (
	"context"
	"sync"

	of "go.openfeature.dev/openfeature/v2"
)

// runModeFn is a function type that defines how flag evaluations are executed across multiple providers.
// It returns an iterator that yields provider names and their corresponding resolution details.
type runModeFn[T of.FlagTypes] func(ctx context.Context, providers []namedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) ResolutionIterator[T]

// runModeParallel evaluates a flag across multiple providers concurrently.
// It launches a goroutine for each provider and yields results as they complete.
// Evaluation stops early if the iterator consumer stops consuming results or if the context is cancelled.
func runModeParallel[T of.FlagTypes](ctx context.Context, providers []namedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) ResolutionIterator[T] {
	type namedResult struct {
		name       string
		resolution *of.GenericResolutionDetail[T]
	}
	return func(yield func(string, *of.GenericResolutionDetail[T]) bool) {
		resolutions := make(chan *namedResult, len(providers))
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			var wg sync.WaitGroup
			for _, provider := range providers {
				wg.Go(func() {
					resolution := evaluate(ctx, provider, provider.Name(), flag, defaultValue, flatCtx)
					select {
					case <-ctx.Done():
						return
					case resolutions <- &namedResult{name: provider.Name(), resolution: resolution}:
					}
				})
			}
			wg.Wait()
			close(resolutions)
		}()

		for result := range resolutions {
			select {
			case <-ctx.Done():
				return
			default:
				if !yield(result.name, result.resolution) {
					return
				}
			}
		}
	}
}

// runModeSequential evaluates a flag across multiple providers one at a time in order.
// It yields each provider's result sequentially and stops if the iterator consumer stops consuming results
// or if the context is cancelled.
func runModeSequential[T of.FlagTypes](ctx context.Context, providers []namedProvider, flag string, defaultValue T, flatCtx of.FlattenedContext) ResolutionIterator[T] {
	return func(yield func(string, *of.GenericResolutionDetail[T]) bool) {
		for _, provider := range providers {
			select {
			case <-ctx.Done():
				return
			default:
				resolution := evaluate(ctx, provider, provider.Name(), flag, defaultValue, flatCtx)
				if !yield(provider.Name(), resolution) {
					return
				}
			}
		}
	}
}
