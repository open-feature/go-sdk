package openfeature

import (
	"fmt"
)

// IClient defines the behaviour required of an openfeature client
type IClient interface {
	Metadata() ClientMetadata
	AddHooks(hooks ...Hook)
	GetBooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error)
	GetStringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error)
	GetNumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error)
	GetObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error)
	GetBooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	GetStringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	GetNumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	GetObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
}

// ClientMetadata provides a client's metadata
type ClientMetadata struct {
	name string
}

// Name returns the client's name
func (cm ClientMetadata) Name() string {
	return cm.name
}

// Client implements the behaviour required of an openfeature client
type Client struct {
	metadata ClientMetadata
	hooks    []Hook
}

// GetClient returns a new Client. Name is a unique identifier for this client
func GetClient(name string) *Client {
	return &Client{
		metadata: ClientMetadata{name: name},
		hooks:    []Hook{},
	}
}

// Metadata returns the client's metadata
func (c Client) Metadata() ClientMetadata {
	return c.metadata
}

// AddHooks appends to the client's collection of any previously added hooks
func (c *Client) AddHooks(hooks ...Hook) {
	c.hooks = append(c.hooks, hooks...)
}

// Type represents the type of a flag
type Type int64

const (
	Boolean Type = iota
	String
	Number
	Object
)

type EvaluationDetails struct {
	FlagKey  string
	FlagType Type
	ResolutionDetail
}

// GetBooleanValue return boolean evaluation for flag
func (c Client) GetBooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error) {
	evalDetails, err := c.evaluate(flag, Boolean, defaultValue, evalCtx, options)
	return evalDetails.Value.(bool), err
}

// GetStringValue return string evaluation for flag
func (c Client) GetStringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error) {
	evalDetails, err := c.evaluate(flag, String, defaultValue, evalCtx, options)
	return evalDetails.Value.(string), err
}

// GetNumberValue return number evaluation for flag
func (c Client) GetNumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error) {
	evalDetails, err := c.evaluate(flag, Number, defaultValue, evalCtx, options)
	return evalDetails.Value.(float64), err
}

// GetObjectValue return object evaluation for flag
func (c Client) GetObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error) {
	evalDetails, err := c.evaluate(flag, Object, defaultValue, evalCtx, options)
	return evalDetails.Value, err
}

// GetBooleanValueDetails return boolean evaluation for flag
func (c Client) GetBooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Boolean, defaultValue, evalCtx, options)
}

// GetStringValueDetails return string evaluation for flag
func (c Client) GetStringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, String, defaultValue, evalCtx, options)
}

// GetNumberValueDetails return number evaluation for flag
func (c Client) GetNumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Number, defaultValue, evalCtx, options)
}

// GetObjectValueDetails return object evaluation for flag
func (c Client) GetObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Object, defaultValue, evalCtx, options)
}

func (c Client) evaluate(
	flag string, flagType Type, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions,
) (EvaluationDetails, error) {
	var err error
	hookCtx := HookContext{
		flagKey:           flag,
		flagType:          flagType,
		defaultValue:      defaultValue,
		clientMetadata:    c.metadata,
		providerMetadata:  api.provider.Metadata(),
		evaluationContext: evalCtx,
	}
	evalDetails := EvaluationDetails{
		FlagKey:  flag,
		FlagType: flagType,
		ResolutionDetail: ResolutionDetail{
			Value: defaultValue,
		},
	}

	apiClientInvocationHooks := append(append(api.hooks, c.hooks...), options.hooks...) // API, Client, Invocation
	invocationClientApiHooks := append(append(options.hooks, c.hooks...), api.hooks...) // Invocation, Client, API
	defer func() {
		c.finallyHooks(hookCtx, invocationClientApiHooks, options)
	}()

	evalCtx, err = c.beforeHooks(hookCtx, apiClientInvocationHooks, evalCtx, options)
	hookCtx.evaluationContext = evalCtx
	if err != nil {
		err = fmt.Errorf("failed to execute before hook: %w", err)
		c.errorHooks(hookCtx, invocationClientApiHooks, err, options)
		return evalDetails, err
	}

	var resolution ResolutionDetail
	switch flagType {
	case Object:
		resolution = api.provider.GetObjectEvaluation(flag, defaultValue, evalCtx, options)
	case Boolean:
		defValue := defaultValue.(bool)
		res := api.provider.GetBooleanEvaluation(flag, defValue, evalCtx, options)
		resolution = res.ResolutionDetail
	case String:
		defValue := defaultValue.(string)
		res := api.provider.GetStringEvaluation(flag, defValue, evalCtx, options)
		resolution = res.ResolutionDetail
	case Number:
		defValue := defaultValue.(float64)
		res := api.provider.GetNumberEvaluation(flag, defValue, evalCtx, options)
		resolution = res.ResolutionDetail
	}

	err = resolution.Error()
	if err != nil {
		err = fmt.Errorf("failed to evaluate the flag: %w", err)
		c.errorHooks(hookCtx, invocationClientApiHooks, err, options)
		return evalDetails, err
	}
	if resolution.Value != nil {
		evalDetails.ResolutionDetail = resolution
	}

	if err := c.afterHooks(hookCtx, invocationClientApiHooks, evalDetails, options); err != nil {
		err = fmt.Errorf("failed to execute after hook: %w", err)
		c.errorHooks(hookCtx, invocationClientApiHooks, err, options)
		return evalDetails, err
	}

	return evalDetails, nil
}

func (c Client) beforeHooks(
	hookCtx HookContext, hooks []Hook, evalCtx EvaluationContext, options EvaluationOptions,
) (EvaluationContext, error) {
	for _, hook := range hooks {
		resultEvalCtx, err := hook.Before(hookCtx, options.hookHints)
		if resultEvalCtx != nil {
			hookCtx.evaluationContext = *resultEvalCtx
		}
		if err != nil {
			return mergeContexts(evalCtx, hookCtx.evaluationContext), err
		}
	}

	return mergeContexts(evalCtx, hookCtx.evaluationContext), nil
}

func (c Client) afterHooks(
	hookCtx HookContext, hooks []Hook, evalDetails EvaluationDetails, options EvaluationOptions,
) error {
	for _, hook := range hooks {
		if err := hook.After(hookCtx, evalDetails, options.hookHints); err != nil {
			return err
		}
	}

	return nil
}

func (c Client) errorHooks(hookCtx HookContext, hooks []Hook, err error, options EvaluationOptions) {
	for _, hook := range hooks {
		hook.Error(hookCtx, err, options.hookHints)
	}
}

func (c Client) finallyHooks(hookCtx HookContext, hooks []Hook, options EvaluationOptions) {
	for _, hook := range hooks {
		hook.Finally(hookCtx, options.hookHints)
	}
}

// merges attributes from the given EvaluationContexts with the primary EvaluationContext taking precedence in case
// of any conflicts
func mergeContexts(primary, secondary EvaluationContext) EvaluationContext {
	mergedCtx := primary

	for k, v := range secondary.Attributes {
		_, ok := primary.Attributes[k]
		if !ok {
			mergedCtx.Attributes[k] = v
		}
	}

	return mergedCtx
}
