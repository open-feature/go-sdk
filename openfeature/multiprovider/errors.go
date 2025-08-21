package multiprovider

import (
	"errors"
	"fmt"
)

type (
	// ProviderError is an error wrapper that species the provider name.
	ProviderError struct {
		Err          error
		ProviderName string
	}

	// AggregateError is a map that contains up to one error per provider within the multiprovider.
	AggregateError []ProviderError
)

var (
	_ error = (*ProviderError)(nil)
	_ error = (AggregateError)(nil)
)

func (e *ProviderError) Error() string {
	return fmt.Sprintf("Provider %s: %s", e.ProviderName, e.Err.Error())
}

// NewAggregateError Creates a new AggregateError
func NewAggregateError(providerErrors []ProviderError) AggregateError {
	return providerErrors
}

func (ae AggregateError) Error() string {
	size := len(ae)
	switch size {
	case 0:
		return ""
	case 1:
		for _, err := range ae {
			return err.Error()
		}
	}

	errs := make([]error, 0, size)
	for i := range ae {
		errs = append(errs, &ae[i])
	}
	return errors.Join(errs...).Error()
}
