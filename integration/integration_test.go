package integration_test

import (
	"context"
	"errors"

	"github.com/open-feature/go-sdk/pkg/openfeature"
)

// ctxStorageKey is the key used to pass test data across context.Context
type ctxStorageKey struct{}

// ctxClientKey is the key used to pass the openfeature client across context.Context
type ctxClientKey struct{}

func firstBooleanEvaluationDetails(ctx context.Context) (openfeature.BooleanEvaluationDetails, error) {
	store, ok := ctx.Value(ctxStorageKey{}).(map[string]openfeature.BooleanEvaluationDetails)
	if !ok {
		return openfeature.BooleanEvaluationDetails{}, errors.New("no flag resolution result")
	}

	var got openfeature.BooleanEvaluationDetails
	for _, evalDetails := range store {
		got = evalDetails
		break
	}

	return got, nil
}

func firstStringEvaluationDetails(ctx context.Context) (openfeature.StringEvaluationDetails, error) {
	store, ok := ctx.Value(ctxStorageKey{}).(map[string]openfeature.StringEvaluationDetails)
	if !ok {
		return openfeature.StringEvaluationDetails{}, errors.New("no flag resolution result")
	}

	var got openfeature.StringEvaluationDetails
	for _, evalDetails := range store {
		got = evalDetails
		break
	}

	return got, nil
}

func firstIntegerEvaluationDetails(ctx context.Context) (openfeature.IntEvaluationDetails, error) {
	store, ok := ctx.Value(ctxStorageKey{}).(map[string]openfeature.IntEvaluationDetails)
	if !ok {
		return openfeature.IntEvaluationDetails{}, errors.New("no flag resolution result")
	}

	var got openfeature.IntEvaluationDetails
	for _, evalDetails := range store {
		got = evalDetails
		break
	}

	return got, nil
}

func firstFloatEvaluationDetails(ctx context.Context) (openfeature.FloatEvaluationDetails, error) {
	store, ok := ctx.Value(ctxStorageKey{}).(map[string]openfeature.FloatEvaluationDetails)
	if !ok {
		return openfeature.FloatEvaluationDetails{}, errors.New("no flag resolution result")
	}

	var got openfeature.FloatEvaluationDetails
	for _, evalDetails := range store {
		got = evalDetails
		break
	}

	return got, nil
}

func firstInterfaceEvaluationDetails(ctx context.Context) (openfeature.InterfaceEvaluationDetails, error) {
	store, ok := ctx.Value(ctxStorageKey{}).(map[string]openfeature.InterfaceEvaluationDetails)
	if !ok {
		return openfeature.InterfaceEvaluationDetails{}, errors.New("no flag resolution result")
	}

	var got openfeature.InterfaceEvaluationDetails
	for _, evalDetails := range store {
		got = evalDetails
		break
	}

	return got, nil
}
