package openfeature

import "github.com/open-feature/go-sdk/openfeature"

// Deprecated: use github.com/open-feature/go-sdk/openfeature.ErrorCode, instead.
type ErrorCode = openfeature.ErrorCode

const (
	// ProviderNotReadyCode - the value was resolved before the provider was ready.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	ProviderNotReadyCode = openfeature.ProviderNotReadyCode
	// FlagNotFoundCode - the flag could not be found.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	FlagNotFoundCode = openfeature.FlagNotFoundCode
	// ParseErrorCode - an error was encountered parsing data, such as a flag configuration.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	ParseErrorCode = openfeature.ParseErrorCode
	// TypeMismatchCode - the type of the flag value does not match the expected type.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	TypeMismatchCode = openfeature.TypeMismatchCode
	// TargetingKeyMissingCode - the provider requires a targeting key and one was not provided in the evaluation context.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	TargetingKeyMissingCode = openfeature.TargetingKeyMissingCode
	// InvalidContextCode - the evaluation context does not meet provider requirements.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	InvalidContextCode = openfeature.InvalidContextCode
	// GeneralCode - the error was for a reason not enumerated above.
	//
	// Deprecated: use github.com/open-feature/go-sdk/openfeature.UnknownReason, instead.
	GeneralCode = openfeature.GeneralCode
)

// ResolutionError is an enumerated error code with an optional message
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.ResolutionError, instead.
type ResolutionError = openfeature.ResolutionError

// NewProviderNotReadyResolutionError constructs a resolution error with code PROVIDER_NOT_READY
//
// Explanation - The value was resolved before the provider was ready.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewProviderNotReadyResolutionError, instead.
func NewProviderNotReadyResolutionError(msg string) ResolutionError {
	return openfeature.NewProviderNotReadyResolutionError(msg)
}

// NewFlagNotFoundResolutionError constructs a resolution error with code FLAG_NOT_FOUND
//
// Explanation - The flag could not be found.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewFlagNotFoundResolutionError, instead.
func NewFlagNotFoundResolutionError(msg string) ResolutionError {
	return openfeature.NewFlagNotFoundResolutionError(msg)
}

// NewParseErrorResolutionError constructs a resolution error with code PARSE_ERROR
//
// Explanation - An error was encountered parsing data, such as a flag configuration.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewParseErrorResolutionError, instead.
func NewParseErrorResolutionError(msg string) ResolutionError {
	return openfeature.NewParseErrorResolutionError(msg)
}

// NewTypeMismatchResolutionError constructs a resolution error with code TYPE_MISMATCH
//
// Explanation - The type of the flag value does not match the expected type.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewTypeMismatchResolutionError, instead.
func NewTypeMismatchResolutionError(msg string) ResolutionError {
	return openfeature.NewTypeMismatchResolutionError(msg)
}

// NewTargetingKeyMissingResolutionError constructs a resolution error with code TARGETING_KEY_MISSING
//
// Explanation - The provider requires a targeting key and one was not provided in the evaluation context.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewTargetingKeyMissingResolutionError, instead.
func NewTargetingKeyMissingResolutionError(msg string) ResolutionError {
	return openfeature.NewTargetingKeyMissingResolutionError(msg)
}

// NewInvalidContextResolutionError constructs a resolution error with code INVALID_CONTEXT
//
// Explanation - The evaluation context does not meet provider requirements.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewInvalidContextResolutionError, instead.
func NewInvalidContextResolutionError(msg string) ResolutionError {
	return openfeature.NewInvalidContextResolutionError(msg)
}

// NewGeneralResolutionError constructs a resolution error with code GENERAL
//
// Explanation - The error was for a reason not enumerated above.
//
// Deprecated: use github.com/open-feature/go-sdk/openfeature.NewGeneralResolutionError, instead.
func NewGeneralResolutionError(msg string) ResolutionError {
	return openfeature.NewGeneralResolutionError(msg)
}
