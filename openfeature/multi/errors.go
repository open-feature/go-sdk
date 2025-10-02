package multi

import (
	"errors"
	"fmt"
)

type (
	// ProviderError is an error wrapper that includes the provider name.
	ProviderError struct {
		// Err is the original error that was returned from a provider
		Err error
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

func (e *ProviderError) Error() string {
	return fmt.Sprintf("Provider %s: %s", e.ProviderName, e.Err.Error())
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
