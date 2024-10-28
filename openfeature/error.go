package openfeature

import "fmt"

// ProviderInitError represents an error that occurs during provider initialization.
type ProviderInitError struct {
	ErrorCode ErrorCode // Field to store the specific error code
	Message   string    // Custom error message
}

// Error implements the error interface for ProviderInitError.
func (e *ProviderInitError) Error() string {
	return fmt.Sprintf("ProviderInitError: %s (code: %s)", e.Message, e.ErrorCode)
}
