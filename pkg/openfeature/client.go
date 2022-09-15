package openfeature

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
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
	logger            logr.Logger
}

// NewClient returns a new Client. Name is a unique identifier for this client
func NewClient(name string) *Client {
	return &Client{
		metadata:          ClientMetadata{name: name},
		hooks:             []Hook{},
		evaluationContext: EvaluationContext{},
		logger:            api.logger,
	}
}

// WithLogger sets the logger of the client
func (c *Client) WithLogger(l logr.Logger) *Client {
	c.logger = l
	return c
}

// Metadata returns the client's metadata
func (c Client) Metadata() ClientMetadata {
	return c.metadata
}

// AddHooks appends to the client's collection of any previously added hooks
func (c *Client) AddHooks(hooks ...Hook) {
	c.hooks = append(c.hooks, hooks...)
	c.logger.V(info).Info("appended hooks to client", "client", c.Metadata().name, "hooks", hooks)
}

// SetEvaluationContext sets the client's evaluation context
func (c *Client) SetEvaluationContext(evalCtx EvaluationContext) {
	c.evaluationContext = evalCtx
	c.logger.V(info).Info(
		"set client evaluation context", "client", c.Metadata().name, "evaluationContext", evalCtx,
	)
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

func (t Type) String() string {
	return typeToString[t]
}

var typeToString = map[Type]string{
	Boolean: "bool",
	String:  "string",
	Float:   "float",
	Int:     "int",
	Object:  "object",
}

type EvaluationDetails struct {
	FlagKey  string
	FlagType Type
	InterfaceResolutionDetail
}

// BooleanValue returns boolean evaluation for flag.
// If options are provided, only first given is used.
func (c Client) BooleanValue(flag string, defaultValue bool, options ...EvaluationOptions) (bool, error) {
	return c.booleanValue(flag, defaultValue, EvaluationContext{}, options...)
}

// BooleanValueWithContext returns boolean evaluation for flag with context.
// If options are provided, only first given is used.
func (c Client) BooleanValueWithContext(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOptions) (bool, error) {
	return c.booleanValue(flag, defaultValue, evalCtx, options...)
}

func (c Client) booleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOptions) (bool, error) {
	optns := EvaluationOptions{}
	if len(options) > 0 {
		optns = options[0]
	}

	evalDetails, err := c.evaluate(flag, Boolean, defaultValue, evalCtx, optns)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(bool)
	if !ok {
		err := errors.New("evaluated value is not a boolean")
		c.logger.Error(
			err, "invalid flag resolution type", "expectedType", "bool",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// StringValue returns string evaluation for flag.
// If options are provided, only first given is used.
func (c Client) StringValue(flag string, defaultValue string, options ...EvaluationOptions) (string, error) {
	return c.stringValue(flag, defaultValue, EvaluationContext{}, options...)
}

// StringValueWithContext returns string evaluation for flag with context.
// If options are provided, only first given is used.
func (c Client) StringValueWithContext(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOptions) (string, error) {
	return c.stringValue(flag, defaultValue, evalCtx, options...)
}

func (c Client) stringValue(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOptions) (string, error) {
	optns := EvaluationOptions{}
	if len(options) > 0 {
		optns = options[0]
	}

	evalDetails, err := c.evaluate(flag, String, defaultValue, evalCtx, optns)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(string)
	if !ok {
		err := errors.New("evaluated value is not a string")
		c.logger.Error(
			err, "invalid flag resolution type", "expectedType", "string",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// FloatValue returns float evaluation for flag.
// If options are provided, only first given is used.
func (c Client) FloatValue(flag string, defaultValue float64, options ...EvaluationOptions) (float64, error) {
	return c.floatValue(flag, defaultValue, EvaluationContext{}, options...)
}

// FloatValueWithContext returns float evaluation for flag with context.
// If options are provided, only first given is used.
func (c Client) FloatValueWithContext(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOptions) (float64, error) {
	return c.floatValue(flag, defaultValue, evalCtx, options...)
}

func (c Client) floatValue(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOptions) (float64, error) {
	optns := EvaluationOptions{}
	if len(options) > 0 {
		optns = options[0]
	}

	evalDetails, err := c.evaluate(flag, Float, defaultValue, evalCtx, optns)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(float64)
	if !ok {
		err := errors.New("evaluated value is not a float64")
		c.logger.Error(
			err, "invalid flag resolution type", "expectedType", "float64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// IntValue returns int evaluation for flag.
// If options are provided, only first given is used.
func (c Client) IntValue(flag string, defaultValue int64, options ...EvaluationOptions) (int64, error) {
	return c.intValue(flag, defaultValue, EvaluationContext{}, options...)
}

// IntValueWithContext returns int evaluation for flag with context.
// If options are provided, only first given is used.
func (c Client) IntValueWithContext(flag string, defaultValue int64, evalCtx EvaluationContext, options ...EvaluationOptions) (int64, error) {
	return c.intValue(flag, defaultValue, evalCtx, options...)
}

func (c Client) intValue(flag string, defaultValue int64, evalCtx EvaluationContext, options ...EvaluationOptions) (int64, error) {
	optns := EvaluationOptions{}
	if len(options) > 0 {
		optns = options[0]
	}

	evalDetails, err := c.evaluate(flag, Int, defaultValue, evalCtx, optns)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(int64)
	if !ok {
		err := errors.New("evaluated value is not an int64")
		c.logger.Error(
			err, "invalid flag resolution type", "expectedType", "int64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// ObjectValue returns object evaluation for flag.
// If options are provided, only first given is used.
func (c Client) ObjectValue(flag string, defaultValue interface{}, options ...EvaluationOptions) (interface{}, error) {
	return c.objectValue(flag, defaultValue, EvaluationContext{}, options...)
}

// ObjectValueWithContext returns object evaluation for flag with context.
// If options are provided, only first given is used.
func (c Client) ObjectValueWithContext(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOptions) (interface{}, error) {
	return c.objectValue(flag, defaultValue, evalCtx, options...)
}

func (c Client) objectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOptions) (interface{}, error) {
	optns := EvaluationOptions{}
	if len(options) > 0 {
		optns = options[0]
	}

	evalDetails, err := c.evaluate(flag, Object, defaultValue, evalCtx, optns)
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
	c.logger.V(debug).Info(
		"evaluating flag", "flag", flag, "type", flagType.String(), "defaultValue", defaultValue,
		"evaluationContext", evalCtx, "evaluationOptions", options,
	)
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
		InterfaceResolutionDetail: InterfaceResolutionDetail{
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
		c.logger.Error(
			err, "before hook", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(),
		)
		err = fmt.Errorf("before hook: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}

	flatCtx := flattenContext(evalCtx)
	var resolution InterfaceResolutionDetail
	switch flagType {
	case Object:
		resolution = api.provider.ObjectEvaluation(flag, defaultValue, flatCtx)
	case Boolean:
		defValue := defaultValue.(bool)
		res := api.provider.BooleanEvaluation(flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case String:
		defValue := defaultValue.(string)
		res := api.provider.StringEvaluation(flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case Float:
		defValue := defaultValue.(float64)
		res := api.provider.FloatEvaluation(flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case Int:
		defValue := defaultValue.(int64)
		res := api.provider.IntEvaluation(flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	}

	err = resolution.Error()
	if err != nil {
		c.logger.Error(
			err, "flag resolution", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(), "errorCode", err,
		)
		err = fmt.Errorf("error code: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}
	if resolution.Value != nil {
		evalDetails.InterfaceResolutionDetail = resolution
	}

	if err := c.afterHooks(hookCtx, providerInvocationClientApiHooks, evalDetails, options); err != nil {
		c.logger.Error(
			err, "after hook", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(),
		)
		err = fmt.Errorf("after hook: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}

	c.logger.V(debug).Info("evaluated flag", "flag", flag, "details", evalDetails, "type", flagType)
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
	c.logger.V(debug).Info("executing before hooks")
	defer c.logger.V(debug).Info("executed before hooks")

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
	c.logger.V(debug).Info("executing after hooks")
	defer c.logger.V(debug).Info("executed after hooks")

	for _, hook := range hooks {
		if err := hook.After(hookCtx, evalDetails, options.hookHints); err != nil {
			return err
		}
	}

	return nil
}

func (c Client) errorHooks(hookCtx HookContext, hooks []Hook, err error, options EvaluationOptions) {
	c.logger.V(debug).Info("executing error hooks")
	defer c.logger.V(debug).Info("executed error hooks")

	for _, hook := range hooks {
		hook.Error(hookCtx, err, options.hookHints)
	}
}

func (c Client) finallyHooks(hookCtx HookContext, hooks []Hook, options EvaluationOptions) {
	c.logger.V(debug).Info("executing finally hooks")
	defer c.logger.V(debug).Info("executed finally hooks")

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
