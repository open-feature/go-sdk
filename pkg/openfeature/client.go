package openfeature

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"unicode/utf8"

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
	BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error)
	StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error)
	FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error)
	IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error)
	ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error)
}

// ClientMetadata provides a client's metadata
type ClientMetadata struct {
	name string
}

// NewClientMetadata constructs ClientMetadata
// Allows for simplified hook test cases while maintaining immutability
func NewClientMetadata(name string) ClientMetadata {
	return ClientMetadata{
		name: name,
	}
}

// Name returns the client's name
func (cm ClientMetadata) Name() string {
	return cm.name
}

// Client implements the behaviour required of an openfeature client
type Client struct {
	mx                sync.RWMutex
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
	c.mx.Lock()
	defer c.mx.Unlock()
	c.logger = func() logr.Logger { return l }
	return c
}

// Metadata returns the client's metadata
func (c *Client) Metadata() ClientMetadata {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.metadata
}

// AddHooks appends to the client's collection of any previously added hooks
func (c *Client) AddHooks(hooks ...Hook) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.hooks = append(c.hooks, hooks...)
	c.logger().V(info).Info("appended hooks to client", "client", c.metadata.name, "hooks", hooks)
}

// SetEvaluationContext sets the client's evaluation context
func (c *Client) SetEvaluationContext(evalCtx EvaluationContext) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.evaluationContext = evalCtx
	c.logger().V(info).Info(
		"set client evaluation context", "client", c.metadata.name, "evaluationContext", evalCtx,
	)
}

// EvaluationContext returns the client's evaluation context
func (c *Client) EvaluationContext() EvaluationContext {
	c.mx.RLock()
	defer c.mx.RUnlock()
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
	ResolutionDetail
}

type BooleanEvaluationDetails struct {
	Value bool
	EvaluationDetails
}

type StringEvaluationDetails struct {
	Value string
	EvaluationDetails
}

type FloatEvaluationDetails struct {
	Value float64
	EvaluationDetails
}

type IntEvaluationDetails struct {
	Value int64
	EvaluationDetails
}

type InterfaceEvaluationDetails struct {
	Value interface{}
	EvaluationDetails
}

type ResolutionDetail struct {
	Variant      string
	Reason       Reason
	ErrorCode    ErrorCode
	ErrorMessage string
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
func (c *Client) BooleanValue(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (bool, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

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
func (c *Client) StringValue(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (string, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

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
func (c *Client) FloatValue(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (float64, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

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
func (c *Client) IntValue(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (int64, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

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
func (c *Client) ObjectValue(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (interface{}, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

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
func (c *Client) BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Boolean, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return BooleanEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}, err
	}

	value, ok := evalDetails.Value.(bool)
	if !ok {
		err := errors.New("evaluated value is not a boolean")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "boolean",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		boolEvalDetails := BooleanEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}
		boolEvalDetails.EvaluationDetails.ErrorCode = TypeMismatchCode
		boolEvalDetails.EvaluationDetails.ErrorMessage = err.Error()

		return boolEvalDetails, err
	}

	return BooleanEvaluationDetails{
		Value:             value,
		EvaluationDetails: evalDetails.EvaluationDetails,
	}, nil
}

// StringValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, String, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return StringEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}, err
	}

	value, ok := evalDetails.Value.(string)
	if !ok {
		err := errors.New("evaluated value is not a string")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "string",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		strEvalDetails := StringEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}
		strEvalDetails.EvaluationDetails.ErrorCode = TypeMismatchCode
		strEvalDetails.EvaluationDetails.ErrorMessage = err.Error()

		return strEvalDetails, err
	}

	return StringEvaluationDetails{
		Value:             value,
		EvaluationDetails: evalDetails.EvaluationDetails,
	}, nil
}

// FloatValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Float, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return FloatEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}, err
	}

	value, ok := evalDetails.Value.(float64)
	if !ok {
		err := errors.New("evaluated value is not a float64")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "float64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		floatEvalDetails := FloatEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}
		floatEvalDetails.EvaluationDetails.ErrorCode = TypeMismatchCode
		floatEvalDetails.EvaluationDetails.ErrorMessage = err.Error()

		return floatEvalDetails, err
	}

	return FloatEvaluationDetails{
		Value:             value,
		EvaluationDetails: evalDetails.EvaluationDetails,
	}, nil
}

// IntValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	evalDetails, err := c.evaluate(ctx, flag, Int, defaultValue, evalCtx, *evalOptions)
	if err != nil {
		return IntEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}, err
	}

	value, ok := evalDetails.Value.(int64)
	if !ok {
		err := errors.New("evaluated value is not an int64")
		c.logger().Error(
			err, "invalid flag resolution type", "expectedType", "int64",
			"gotType", fmt.Sprintf("%T", evalDetails.Value),
		)
		intEvalDetails := IntEvaluationDetails{
			Value:             defaultValue,
			EvaluationDetails: evalDetails.EvaluationDetails,
		}
		intEvalDetails.EvaluationDetails.ErrorCode = TypeMismatchCode
		intEvalDetails.EvaluationDetails.ErrorMessage = err.Error()

		return intEvalDetails, err
	}

	return IntEvaluationDetails{
		Value:             value,
		EvaluationDetails: evalDetails.EvaluationDetails,
	}, nil
}

// ObjectValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
// - ctx is the standard go context struct used to manage requests (e.g. timeouts)
// - flag is the key that uniquely identifies a particular flag
// - defaultValue is returned if an error occurs
// - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
// - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) ObjectValueDetails(ctx context.Context, flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...Option) (InterfaceEvaluationDetails, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	return c.evaluate(ctx, flag, Object, defaultValue, evalCtx, *evalOptions)
}

func (c *Client) evaluate(
	ctx context.Context, flag string, flagType Type, defaultValue interface{}, evalCtx EvaluationContext, options EvaluationOptions,
) (InterfaceEvaluationDetails, error) {
	c.logger().V(debug).Info(
		"evaluating flag", "flag", flag, "type", flagType.String(), "defaultValue", defaultValue,
		"evaluationContext", evalCtx, "evaluationOptions", options,
	)

	evalDetails := InterfaceEvaluationDetails{
		Value: defaultValue,
		EvaluationDetails: EvaluationDetails{
			FlagKey:  flag,
			FlagType: flagType,
		},
	}

	if !utf8.Valid([]byte(flag)) {
		return evalDetails, NewParseErrorResolutionError("flag key is not a UTF-8 encoded string")
	}

	// ensure that the same provider & hooks are used across this transaction to avoid unexpected behaviour
	api.RLock()
	provider := api.prvder
	globalHooks := api.hks
	globalCtx := api.evalCtx
	api.RUnlock()

	evalCtx = mergeContexts(evalCtx, c.evaluationContext, globalCtx)                                                           // API (global) -> client -> invocation
	apiClientInvocationProviderHooks := append(append(append(globalHooks, c.hooks...), options.hooks...), provider.Hooks()...) // API, Client, Invocation, Provider
	providerInvocationClientApiHooks := append(append(append(provider.Hooks(), options.hooks...), c.hooks...), globalHooks...) // Provider, Invocation, Client, API

	var err error
	hookCtx := HookContext{
		flagKey:           flag,
		flagType:          flagType,
		defaultValue:      defaultValue,
		clientMetadata:    c.metadata,
		providerMetadata:  provider.Metadata(),
		evaluationContext: evalCtx,
	}

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
		resolution = provider.ObjectEvaluation(ctx, flag, defaultValue, flatCtx)
	case Boolean:
		defValue := defaultValue.(bool)
		res := provider.BooleanEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case String:
		defValue := defaultValue.(string)
		res := provider.StringEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case Float:
		defValue := defaultValue.(float64)
		res := provider.FloatEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case Int:
		defValue := defaultValue.(int64)
		res := provider.IntEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	}

	err = resolution.Error()
	if err != nil {
		c.logger().Error(
			err, "flag resolution", "flag", flag, "defaultValue", defaultValue,
			"evaluationContext", evalCtx, "evaluationOptions", options, "type", flagType.String(), "errorCode", err,
			"errMessage", resolution.ResolutionError.message,
		)
		err = fmt.Errorf("error code: %w", err)
		c.errorHooks(hookCtx, providerInvocationClientApiHooks, err, options)
		evalDetails.ResolutionDetail = resolution.ResolutionDetail()
		evalDetails.Reason = ErrorReason
		return evalDetails, err
	}
	evalDetails.Value = resolution.Value
	evalDetails.ResolutionDetail = resolution.ResolutionDetail()

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

func flattenContext(evalCtx EvaluationContext) FlattenedContext {
	flatCtx := FlattenedContext{}
	if evalCtx.attributes != nil {
		flatCtx = evalCtx.Attributes()
	}
	if evalCtx.targetingKey != "" {
		flatCtx[TargetingKey] = evalCtx.targetingKey
	}
	return flatCtx
}

func (c *Client) beforeHooks(
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

func (c *Client) afterHooks(
	hookCtx HookContext, hooks []Hook, evalDetails InterfaceEvaluationDetails, options EvaluationOptions,
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

func (c *Client) errorHooks(hookCtx HookContext, hooks []Hook, err error, options EvaluationOptions) {
	c.logger().V(debug).Info("executing error hooks")
	defer c.logger().V(debug).Info("executed error hooks")

	for _, hook := range hooks {
		hook.Error(hookCtx, err, options.hookHints)
	}
}

func (c *Client) finallyHooks(hookCtx HookContext, hooks []Hook, options EvaluationOptions) {
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

	// create copy to prevent mutation of given EvaluationContext
	mergedCtx := EvaluationContext{
		attributes:   evaluationContexts[0].Attributes(),
		targetingKey: evaluationContexts[0].targetingKey,
	}

	for i := 1; i < len(evaluationContexts); i++ {
		if mergedCtx.targetingKey == "" && evaluationContexts[i].targetingKey != "" {
			mergedCtx.targetingKey = evaluationContexts[i].targetingKey
		}

		for k, v := range evaluationContexts[i].attributes {
			_, ok := mergedCtx.attributes[k]
			if !ok {
				mergedCtx.attributes[k] = v
			}
		}
	}

	return mergedCtx
}
