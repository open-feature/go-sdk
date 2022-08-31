package openfeature

import (
	"errors"
	"fmt"
)

// IClient defines the behaviour required of an openfeature client
type IClient interface {
	Metadata() ClientMetadata
	AddHooks(hooks ...Hook)
	SetEvaluationContext(evalCtx EvaluationContext)
	EvaluationContext() EvaluationContext
	BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error)
	StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error)
	FloatValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error)
	IntValue(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (int64, error)
	ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error)
	BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	FloatValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	IntValueDetails(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
	ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error)
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
	metadata          ClientMetadata
	hooks             []Hook
	evaluationContext EvaluationContext
}

// NewClient returns a new Client. Name is a unique identifier for this client
func NewClient(name string) *Client {
	return &Client{
		metadata:          ClientMetadata{name: name},
		hooks:             []Hook{},
		evaluationContext: EvaluationContext{},
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

// SetEvaluationContext sets the client's evaluation context
func (c *Client) SetEvaluationContext(evalCtx EvaluationContext) {
	c.evaluationContext = evalCtx
}

// EvaluationContext returns the client's evaluation context
func (c *Client) EvaluationContext() EvaluationContext {
	return c.evaluationContext
}

// Type represents the type of a flag
type Type int64

const (
	Boolean Type = iota
	String
	Float
	Int
	Object
)

type EvaluationDetails struct {
	FlagKey  string
	FlagType Type
	ResolutionDetail
}

// BooleanValue return boolean evaluation for flag
func (c Client) BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (bool, error) {
	evalDetails, err := c.evaluate(flag, Boolean, defaultValue, evalCtx, options)
	if err != nil {
		return defaultValue, fmt.Errorf("evaluate: %w", err)
	}

	value, ok := evalDetails.Value.(bool)
	if !ok {
		return defaultValue, errors.New("evaluated value is not a boolean")
	}

	return value, nil
}

// StringValue return string evaluation for flag
func (c Client) StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (string, error) {
	evalDetails, err := c.evaluate(flag, String, defaultValue, evalCtx, options)
	if err != nil {
		return defaultValue, fmt.Errorf("evaluate: %w", err)
	}

	value, ok := evalDetails.Value.(string)
	if !ok {
		return defaultValue, errors.New("evaluated value is not a string")
	}

	return value, nil
}

// FloatValue return float evaluation for flag
func (c Client) FloatValue(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (float64, error) {
	evalDetails, err := c.evaluate(flag, Float, defaultValue, evalCtx, options)
	if err != nil {
		return defaultValue, fmt.Errorf("evaluate: %w", err)
	}

	value, ok := evalDetails.Value.(float64)
	if !ok {
		return defaultValue, errors.New("evaluated value is not a float64")
	}

	return value, nil
}

// IntValue return int evaluation for flag
func (c Client) IntValue(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (int64, error) {
	evalDetails, err := c.evaluate(flag, Int, defaultValue, evalCtx, options)
	if err != nil {
		return defaultValue, fmt.Errorf("evaluate: %w", err)
	}

	value, ok := evalDetails.Value.(int64)
	if !ok {
		return defaultValue, errors.New("evaluated value is not an int64")
	}

	return value, nil
}

// ObjectValue return object evaluation for flag
func (c Client) ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (interface{}, error) {
	evalDetails, err := c.evaluate(flag, Object, defaultValue, evalCtx, options)
	return evalDetails.Value, err
}

// BooleanValueDetails return boolean evaluation for flag
func (c Client) BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Boolean, defaultValue, evalCtx, options)
}

// StringValueDetails return string evaluation for flag
func (c Client) StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, String, defaultValue, evalCtx, options)
}

// FloatValueDetails return float evaluation for flag
func (c Client) FloatValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Float, defaultValue, evalCtx, options)
}

// IntValueDetails return int evaluation for flag
func (c Client) IntValueDetails(flag string, defaultValue int64, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Int, defaultValue, evalCtx, options)
}

// ObjectValueDetails return object evaluation for flag
func (c Client) ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions) (EvaluationDetails, error) {
	return c.evaluate(flag, Object, defaultValue, evalCtx, options)
}

func (c Client) evaluate(
	flag string, flagType Type, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions,
) (EvaluationDetails, error) {
	evalCtx = mergeContexts(evalCtx, c.evaluationContext, api.evaluationContext) // API (global) -> client -> invocation

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

	apiClientInvocationProviderHooks := append(append(append(api.hooks, c.hooks...), options.hooks...), api.provider.Hooks()...) // API, Client, Invocation, Provider
	providerInvocationClientApiHooks := append(append(append(api.provider.Hooks(), options.hooks...), c.hooks...), api.hooks...) // Provider, Invocation, Client, API
	defer func() {
		c.finallyHooks(hookCtx, providerInvocationClientApiHooks, options)
	}()

	evalCtx, err = c.beforeHooks(hookCtx, apiClientInvocationProviderHooks, evalCtx, options)
	hookCtx.evaluationContext = evalCtx
	if err != nil {
		err = fmt.Errorf("execute before hook: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}

	flatCtx := flattenContext(evalCtx)
	var resolution ResolutionDetail
	switch flagType {
	case Object:
		resolution = api.provider.ObjectEvaluation(flag, defaultValue, flatCtx)
	case Boolean:
		defValue := defaultValue.(bool)
		res := api.provider.BooleanEvaluation(flag, defValue, flatCtx)
		resolution = res.ResolutionDetail
	case String:
		defValue := defaultValue.(string)
		res := api.provider.StringEvaluation(flag, defValue, flatCtx)
		resolution = res.ResolutionDetail
	case Float:
		defValue := defaultValue.(float64)
		res := api.provider.FloatEvaluation(flag, defValue, flatCtx)
		resolution = res.ResolutionDetail
	case Int:
		defValue := defaultValue.(int64)
		res := api.provider.IntEvaluation(flag, defValue, flatCtx)
		resolution = res.ResolutionDetail
	}

	err = resolution.Error()
	if err != nil {
		err = fmt.Errorf("evaluate the flag: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}
	if resolution.Value != nil {
		evalDetails.ResolutionDetail = resolution
	}

	if err := c.afterHooks(hookCtx, providerInvocationClientApiHooks, evalDetails, options); err != nil {
		err = fmt.Errorf("execute after hook: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}

	return evalDetails, nil
}

func flattenContext(evalCtx EvaluationContext) map[string]interface{} {
	flatCtx := map[string]interface{}{}
	if evalCtx.Attributes != nil {
		flatCtx = evalCtx.Attributes
	}
	if evalCtx.TargetingKey != "" {
		flatCtx[TargetingKey] = evalCtx.TargetingKey
	}
	return flatCtx
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
			return mergeContexts(hookCtx.evaluationContext, evalCtx), err
		}
	}

	return mergeContexts(hookCtx.evaluationContext, evalCtx), nil
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

// merges attributes from the given EvaluationContexts with the nth EvaluationContext taking precedence in case
// of any conflicts with the (n+1)th EvaluationContext
func mergeContexts(evaluationContexts ...EvaluationContext) EvaluationContext {
	if len(evaluationContexts) == 0 {
		return EvaluationContext{}
	}

	mergedCtx := evaluationContexts[0]

	for i := 1; i < len(evaluationContexts); i++ {
		if mergedCtx.TargetingKey == "" && evaluationContexts[i].TargetingKey != "" {
			mergedCtx.TargetingKey = evaluationContexts[i].TargetingKey
		}

		for k, v := range evaluationContexts[i].Attributes {
			_, ok := mergedCtx.Attributes[k]
			if !ok {
				mergedCtx.Attributes[k] = v
			}
		}
	}

	return mergedCtx
}
