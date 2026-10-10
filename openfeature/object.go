package openfeature

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetObject performs an object flag evaluation and returns its value deserialized
// into T. Any error encountered during the evaluation or the deserialization results
// in defaultValue being returned. To explicitly handle errors, use [GetObjectValue]
// or [GetObjectValueDetails].
func GetObject[T any](
	ctx context.Context, client *Client, flag string,
	defaultValue T, evalCtx EvaluationContext, options ...Option,
) T {
	value, _ := GetObjectValue(ctx, client, flag, defaultValue, evalCtx, options...)

	return value
}

// GetObjectValue performs an object flag evaluation and returns its value
// deserialized into T.
func GetObjectValue[T any](
	ctx context.Context, client *Client, flag string,
	defaultValue T, evalCtx EvaluationContext, options ...Option,
) (T, error) {
	details, err := GetObjectValueDetails(ctx, client, flag, defaultValue, evalCtx, options...)

	return details.Value, err
}

// GetObjectValueDetails performs an object flag evaluation that returns an
// evaluation details struct whose value is deserialized into T.
//
// A value that is already a T is used as-is, otherwise it is round-tripped through
// JSON. A value that cannot be deserialized into T is abnormal execution:
// defaultValue is returned with an error reason and a TYPE_MISMATCH error code.
func GetObjectValueDetails[T any](
	ctx context.Context, client *Client, flag string,
	defaultValue T, evalCtx EvaluationContext, options ...Option,
) (GenericEvaluationDetails[T], error) {
	evalDetails, err := client.ObjectValueDetails(ctx, flag, defaultValue, evalCtx, options...)

	typedDetails := GenericEvaluationDetails[T]{
		Value:             defaultValue,
		EvaluationDetails: evalDetails.EvaluationDetails,
	}
	if err != nil {
		return typedDetails, err
	}

	// The provider already returned a T; no conversion required.
	if value, ok := evalDetails.Value.(T); ok {
		typedDetails.Value = value

		return typedDetails, nil
	}

	// The provider returned a deserialized structure (e.g. map[string]any).
	// Round-trip it through JSON to reach T.
	value, convErr := convertObject[T](evalDetails.Value)
	if convErr != nil {
		err := fmt.Errorf("evaluated value is not a %T: %w", defaultValue, convErr)
		typedDetails.Reason = ErrorReason
		typedDetails.ErrorCode = TypeMismatchCode
		typedDetails.ErrorMessage = err.Error()

		return typedDetails, err
	}

	typedDetails.Value = value

	return typedDetails, nil
}

// convertObject deserializes value into T by round-tripping it through JSON.
func convertObject[T any](value any) (T, error) {
	var converted T

	data, err := json.Marshal(value)
	if err != nil {
		return converted, err
	}

	if err := json.Unmarshal(data, &converted); err != nil {
		return converted, err
	}

	return converted, nil
}
