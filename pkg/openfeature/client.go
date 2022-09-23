package openfeature

import (
	"context"
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
	BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error)
	StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error)
	FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error)
	IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error)
	ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error)
	BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
	StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
	FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
	IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
	ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error)
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
	logger            func() logr.Logger
}

// NewClient returns a new Client. Name is a unique identifier for this client
func NewClient(name string) *Client {
	return &Client{
		metadata:          ClientMetadata{name: name},
		hooks:             []Hook{},
		evaluationContext: EvaluationContext{},
		logger:            globalLogger,
	}
}

// WithLogger sets the logger of the client
func (c *Client) WithLogger(l logr.Logger) *Client {
	c.logger = func() logr.Logger { return l }
	return c
}

// Metadata returns the client's metadata
func (c Client) Metadata() ClientMetadata {
	return c.metadata
}

// AddHooks appends to the client's collection of any previously added hooks
func (c *Client) AddHooks(hooks ...Hook) {
	c.hooks = append(c.hooks, hooks...)
	c.logger().V(info).Info("appended hooks to client", "client", c.Metadata().name, "hooks", hooks)
}

// SetEvaluationContext sets the client's evaluation context
func (c *Client) SetEvaluationContext(evalCtx EvaluationContext) {
	c.evaluationContext = evalCtx
	c.logger().V(info).Info(
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

// Option applies a change to EvaluationOptions
type Option func(*EvaluationOptions)

// WithHooks applies provided hooks.
func WithHooks(hooks ...Hook) Option {
	return func(options *EvaluationOptions) {
		options.hooks = hooks
	}
}

// WithHookHints applies provided hook hints.
func WithHookHints(hookHints HookHints) Option {
	return func(options *EvaluationOptions) {
		options.hookHints = hookHints
	}
}

// BooleanValue performs a flag evaluation that returns a boolean.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Boolean, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(bool)
	if !ok {
		err := errors.New("evaluated value is not a boolean")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "bool",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// StringValue performs a flag evaluation that returns a string.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, String, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(string)
	if !ok {
		err := errors.New("evaluated value is not a string")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "string",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// FloatValue performs a flag evaluation that returns a float64.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Float, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(float64)
	if !ok {
		err := errors.New("evaluated value is not a float64")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "float64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// IntValue performs a flag evaluation that returns an int64.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Int, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return defaultValue, err
	}

	value, ok := evalDetails.Value.(int64)
	if !ok {
		err := errors.New("evaluated value is not an int64")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "int64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		return defaultValue, err
	}

	return value, nil
}

// ObjectValue performs a flag evaluation that returns an object.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Object, defaultValue, evalCtx, *evalOptions)
	return evalDetails.Value, err
}

// BooleanValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, Boolean, defaultValue, evalCtx, *evalOptions)
}

// StringValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, String, defaultValue, evalCtx, *evalOptions)
}

// FloatValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, Float, defaultValue, evalCtx, *evalOptions)
}

// IntValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, Int, defaultValue, evalCtx, *evalOptions)
}

// ObjectValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c Client) ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (EvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, Object, defaultValue, evalCtx, *evalOptions)
}

func (c Client) evaluate(
	ctx context.Context, flag string, flagType Type, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions,
) (EvaluationDetails, error) {
	c.logger().V(debug).Info(
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
		c.logger().Error(
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
		resolution = api.provider.ObjectEvaluation(ctx, flag, defaultValue, flatCtx)
	case Boolean:
		defValue := defaultValue.(bool)
		res := api.provider.BooleanEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case String:
		defValue := defaultValue.(string)
		res := api.provider.StringEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case Float:
		defValue := defaultValue.(float64)
		res := api.provider.FloatEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	case Int:
		defValue := defaultValue.(int64)
		res := api.provider.IntEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ResolutionDetail = res.ResolutionDetail
		resolution.Value = res.Value
	}

	err = resolution.Error()
	if err != nil {
		c.logger().Error(
			err, "flag resolution", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(), "errorCode", err,
		)
		err = fmt.Errorf("error code: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		evalDetails.InterfaceResolutionDetail.ResolutionDetail = resolution.ResolutionDetail
		return evalDetails, err
	}
	if resolution.Value != nil {
		evalDetails.InterfaceResolutionDetail = resolution
	}

	if err := c.afterHooks(hookCtx, providerInvocationClientApiHooks, evalDetails, options); err != nil {
		c.logger().Error(
			err, "after hook", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(),
		)
		err = fmt.Errorf("after hook: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		return evalDetails, err
	}

	c.logger().V(debug).Info("evaluated flag", "flag", flag, "details", evalDetails, "type", flagType)
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
	c.logger().V(debug).Info("executing before hooks")
	defer c.logger().V(debug).Info("executed before hooks")

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
	c.logger().V(debug).Info("executing after hooks")
	defer c.logger().V(debug).Info("executed after hooks")

	for _, hook := range hooks {
		if err := hook.After(hookCtx, evalDetails, options.hookHints); err != nil {
			return err
		}
	}

	return nil
}

func (c Client) errorHooks(hookCtx HookContext, hooks []Hook, err error, options EvaluationOptions) {
	c.logger().V(debug).Info("executing error hooks")
	defer c.logger().V(debug).Info("executed error hooks")

	for _, hook := range hooks {
		hook.Error(hookCtx, err, options.hookHints)
	}
}

func (c Client) finallyHooks(hookCtx HookContext, hooks []Hook, options EvaluationOptions) {
	c.logger().V(debug).Info("executing finally hooks")
	defer c.logger().V(debug).Info("executed finally hooks")

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
