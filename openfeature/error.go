package openfeature

import (
	"errors"
	"fmt"
)

// ProviderInitError represents an error that occurs during provider initialization.
type ProviderInitError struct {
	ErrorCode ErrorCode // Field to store the specific error code
	Message   string    // Custom error message
}

// Error implements the error interface for ProviderInitError.
func (e *ProviderInitError) Error() string {
	return fmt.Sprintf("ProviderInitError: %s (code: %s)", e.Message, e.ErrorCode)
}

// ProviderNotReadyError signifies that an operation failed because the provider is in a NOT_READY state.
var ProviderNotReadyError = errors.New("provider not yet initialized")

// ProviderFatalError signifies that an operation failed because the provider is in a FATAL state.
var ProviderFatalError = errors.New("provider is in an irrecoverable error state")
