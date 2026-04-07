package multi

import (
	"errors"
	"fmt"
	"sync"
)

type (
	// ProviderError is an error wrapper that includes the provider name.
	ProviderError struct {
		// err is the original error that was returned from a provider
		err error
		// ProviderName is the name of the provider that returned the included error
		ProviderName string
	}

	// AggregateError is a map that contains up to one error per provider within the multiprovider.
	AggregateError []ProviderError
)

// Compile-time interface compliance checks
var (
	_ error = (*ProviderError)(nil)
	_ error = (AggregateError)(nil)
)

// Error implements the error interface for ProviderError.
func (e *ProviderError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("Provider %s: <nil>", e.ProviderName)
	}
	return fmt.Sprintf("Provider %s: %s", e.ProviderName, e.err.Error())
}

// Unwrap allows access to the original error, if any.
func (e *ProviderError) Unwrap() error {
	return e.err
}

// NewAggregateError creates a new AggregateError from a slice of [ProviderError] instances
func NewAggregateError(providerErrors []ProviderError) AggregateError {
	return providerErrors
}

func (ae AggregateError) Error() string {
	if len(ae) == 0 {
		return ""
	}

	errs := make([]error, 0, len(ae))
	for i := range ae {
		errs = append(errs, &ae[i])
	}
	return errors.Join(errs...).Error()
}

// multiErrGroup collects all errors from concurrent goroutines.
type multiErrGroup struct {
	wg     sync.WaitGroup
	mu     sync.Mutex
	errors []error
}

// Go starts a function in a goroutine.
func (g *multiErrGroup) Go(fn func() error) {
	g.wg.Go(func() {
		if err := fn(); err != nil {
			g.mu.Lock()
			g.errors = append(g.errors, err)
			g.mu.Unlock()
		}
	})
}

// Wait waits for all goroutines to complete.
// Returns a combined error or nil if none.
func (g *multiErrGroup) Wait() error {
	g.wg.Wait()
	return errors.Join(g.errors...)
}
