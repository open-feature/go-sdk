package openfeature

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"unicode/utf8"
)

// ClientMetadata provides a client's metadata
type ClientMetadata struct {
	domain string
}

// NewClientMetadata constructs ClientMetadata
// Allows for simplified hook test cases while maintaining immutability
func NewClientMetadata(domain string) ClientMetadata {
	return ClientMetadata{
		domain: domain,
	}
}

// Domain returns the client's domain
func (cm ClientMetadata) Domain() string {
	return cm.domain
}

// Client implements the behaviour required of an openfeature client
type Client struct {
	api               evaluationImpl
	clientEventing    clientEvent
	metadata          ClientMetadata
	hooks             []Hook
	evaluationContext EvaluationContext
	domain            string

	mx sync.RWMutex
}

// interface guard to ensure that Client implements IClient
var _ iClient = (*Client)(nil)

// NewClient returns a new [Client].
// By default it returns the client for the default domain. The default domain [Client] is the [IClient] instance that
// wraps around an unnamed [FeatureProvider].
// To get the domain specific client use [WithDomain] option with a unique identifier for this client.
func NewClient(opts ...CallOption) *Client {
	c := newCallOption(opts...)
	return newClient(c.domain, api, eventing)
}

func newClient(domain string, apiRef evaluationImpl, eventRef clientEvent) *Client {
	return &Client{
		domain:            domain,
		api:               apiRef,
		clientEventing:    eventRef,
		metadata:          ClientMetadata{domain: domain},
		hooks:             []Hook{},
		evaluationContext: EvaluationContext{},
	}
}

// State returns the state of the associated provider
func (c *Client) State() State {
	return c.clientEventing.State(c.domain)
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
}

// AddHandler allows to add Client level event handler
func (c *Client) AddHandler(eventType EventType, callback EventCallback) {
	c.clientEventing.AddClientHandler(c.metadata.Domain(), eventType, callback)
}

// RemoveHandler allows to remove Client level event handler
func (c *Client) RemoveHandler(eventType EventType, callback EventCallback) {
	c.clientEventing.RemoveClientHandler(c.metadata.Domain(), eventType, callback)
}

// SetEvaluationContext sets the client's evaluation context
func (c *Client) SetEvaluationContext(evalCtx EvaluationContext) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.evaluationContext = evalCtx
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

type EvaluationDetails[T FlagTypes] struct {
	FlagKey  string
	FlagType Type
	ResolutionDetail
	Value T
}

type (
	// BooleanEvaluationDetails represents the result of the flag evaluation process for boolean flags.
	BooleanEvaluationDetails = EvaluationDetails[bool]
	// StringEvaluationDetails represents the result of the flag evaluation process for string flags.
	StringEvaluationDetails = EvaluationDetails[string]
	// FloatEvaluationDetails represents the result of the flag evaluation process for float64 flags.
	FloatEvaluationDetails = EvaluationDetails[float64]
	// IntEvaluationDetails represents the result of the flag evaluation process for int64 flags.
	IntEvaluationDetails = EvaluationDetails[int64]
	// ObjectEvaluationDetails represents the result of the flag evaluation process for Object flags.
	ObjectEvaluationDetails = EvaluationDetails[any]

	// HookEvaluationDetails represents the result for hooks usage.
	HookEvaluationDetails = EvaluationDetails[FlagTypes]
)

type ResolutionDetail struct {
	Variant      string
	Reason       Reason
	ErrorCode    ErrorCode
	ErrorMessage string
	FlagMetadata FlagMetadata
}

// FlagMetadata is a structure which supports definition of arbitrary properties, with keys of type string, and values
// of type boolean, string, int64 or float64. This structure is populated by a provider for use by an Application
// Author (via the Evaluation API) or an Application Integrator (via hooks).
type FlagMetadata map[string]any

// GetString fetch string value from FlagMetadata.
// Returns an error if the key does not exist, or, the value is of the wrong type
func (f FlagMetadata) GetString(key string) (string, error) {
	v, ok := f[key]
	if !ok {
		return "", fmt.Errorf("key %s does not exist in FlagMetadata", key)
	}
	switch t := v.(type) {
	case string:
		return v.(string), nil
	default:
		return "", fmt.Errorf("wrong type for key %s, expected string, got %T", key, t)
	}
}

// GetBool fetch bool value from FlagMetadata.
// Returns an error if the key does not exist, or, the value is of the wrong type
func (f FlagMetadata) GetBool(key string) (bool, error) {
	v, ok := f[key]
	if !ok {
		return false, fmt.Errorf("key %s does not exist in FlagMetadata", key)
	}
	switch t := v.(type) {
	case bool:
		return v.(bool), nil
	default:
		return false, fmt.Errorf("wrong type for key %s, expected bool, got %T", key, t)
	}
}

// GetInt fetch int64 value from FlagMetadata.
// Returns an error if the key does not exist, or, the value is of the wrong type
func (f FlagMetadata) GetInt(key string) (int64, error) {
	v, ok := f[key]
	if !ok {
		return 0, fmt.Errorf("key %s does not exist in FlagMetadata", key)
	}
	switch t := v.(type) {
	case int:
		return int64(v.(int)), nil
	case int8:
		return int64(v.(int8)), nil
	case int16:
		return int64(v.(int16)), nil
	case int32:
		return int64(v.(int32)), nil
	case int64:
		return v.(int64), nil
	default:
		return 0, fmt.Errorf("wrong type for key %s, expected integer, got %T", key, t)
	}
}

// GetFloat fetch float64 value from FlagMetadata.
// Returns an error if the key does not exist, or, the value is of the wrong type
func (f FlagMetadata) GetFloat(key string) (float64, error) {
	v, ok := f[key]
	if !ok {
		return 0, fmt.Errorf("key %s does not exist in FlagMetadata", key)
	}
	switch t := v.(type) {
	case float32:
		return float64(v.(float32)), nil
	case float64:
		return v.(float64), nil
	default:
		return 0, fmt.Errorf("wrong type for key %s, expected float, got %T", key, t)
	}
}

// Option applies a change to EvaluationOptions
type Option func(*EvaluationOptions)

// EvaluationOptions should contain a list of hooks to be executed for a flag evaluation
type EvaluationOptions struct {
	hooks     []Hook
	hookHints HookHints
}

// HookHints returns evaluation options' hook hints
func (e EvaluationOptions) HookHints() HookHints {
	return e.hookHints
}

// Hooks returns evaluation options' hooks
func (e EvaluationOptions) Hooks() []Hook {
	return e.hooks
}

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

// BooleanValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) BooleanValueDetails(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) (BooleanEvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	c.mx.RLock()
	evalDetails, err := evaluate(ctx, c, flag, Boolean, defaultValue, evalCtx, *evalOptions)
	c.mx.RUnlock()

	return evalDetails, err
}

// StringValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) StringValueDetails(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) (StringEvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	c.mx.RLock()
	evalDetails, err := evaluate(ctx, c, flag, String, defaultValue, evalCtx, *evalOptions)
	c.mx.RUnlock()

	return evalDetails, err
}

// FloatValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) FloatValueDetails(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) (FloatEvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	c.mx.RLock()
	evalDetails, err := evaluate(ctx, c, flag, Float, defaultValue, evalCtx, *evalOptions)
	c.mx.RUnlock()

	return evalDetails, err
}

// IntValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) IntValueDetails(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) (IntEvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	c.mx.RLock()
	evalDetails, err := evaluate(ctx, c, flag, Int, defaultValue, evalCtx, *evalOptions)
	c.mx.RUnlock()

	return evalDetails, err
}

// ObjectValueDetails performs a flag evaluation that returns an evaluation details struct.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) ObjectValueDetails(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) (ObjectEvaluationDetails, error) {
	evalOptions := &EvaluationOptions{}
	for _, option := range options {
		option(evalOptions)
	}

	c.mx.RLock()
	evalDetails, err := evaluate(ctx, c, flag, Object, defaultValue, evalCtx, *evalOptions)
	c.mx.RUnlock()
	return evalDetails, err
}

// Boolean performs a flag evaluation that returns a boolean. Any error
// encountered during the evaluation will result in the default value being
// returned. To explicitly handle errors, use [Client.BooleanValueDetails]
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) Boolean(ctx context.Context, flag string, defaultValue bool, evalCtx EvaluationContext, options ...Option) bool {
	value, err := c.BooleanValueDetails(ctx, flag, defaultValue, evalCtx, options...)
	if err != nil {
		return defaultValue
	}
	return value.Value
}

// String performs a flag evaluation that returns a string. Any error
// encountered during the evaluation will result in the default value being
// returned. To explicitly handle errors, use [Client.StringValueDetails]
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) String(ctx context.Context, flag string, defaultValue string, evalCtx EvaluationContext, options ...Option) string {
	value, err := c.StringValueDetails(ctx, flag, defaultValue, evalCtx, options...)
	if err != nil {
		return defaultValue
	}
	return value.Value
}

// Float performs a flag evaluation that returns a float64. Any error
// encountered during the evaluation will result in the default value being
// returned. To explicitly handle errors, use [Client.FloatValueDetails]
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) Float(ctx context.Context, flag string, defaultValue float64, evalCtx EvaluationContext, options ...Option) float64 {
	value, err := c.FloatValueDetails(ctx, flag, defaultValue, evalCtx, options...)
	if err != nil {
		return defaultValue
	}
	return value.Value
}

// Int performs a flag evaluation that returns an int64. Any error
// encountered during the evaluation will result in the default value being
// returned. To explicitly handle errors, use [Client.IntValueDetails]
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) Int(ctx context.Context, flag string, defaultValue int64, evalCtx EvaluationContext, options ...Option) int64 {
	value, err := c.IntValueDetails(ctx, flag, defaultValue, evalCtx, options...)
	if err != nil {
		return defaultValue
	}
	return value.Value
}

// Object performs a flag evaluation that returns an object. Any error
// encountered during the evaluation will result in the default value being
// returned. To explicitly handle errors, use [Client.ObjectValueDetails]
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - flag is the key that uniquely identifies a particular flag
//   - defaultValue is returned if an error occurs
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - options are optional additional evaluation options e.g. WithHooks & WithHookHints
func (c *Client) Object(ctx context.Context, flag string, defaultValue any, evalCtx EvaluationContext, options ...Option) any {
	value, err := c.ObjectValueDetails(ctx, flag, defaultValue, evalCtx, options...)
	if err != nil {
		return defaultValue
	}
	return value.Value
}

// Track performs an action for tracking for occurrence  of a particular action or application state.
//
// Parameters:
//   - ctx is the standard go context struct used to manage requests (e.g. timeouts)
//   - trackingEventName is the event name to track
//   - evalCtx is the evaluation context used in a flag evaluation (not to be confused with ctx)
//   - trackingEventDetails defines optional data pertinent to a particular
func (c *Client) Track(ctx context.Context, trackingEventName string, evalCtx EvaluationContext, details TrackingEventDetails) {
	provider, evalCtx := c.forTracking(ctx, evalCtx)
	provider.Track(ctx, trackingEventName, evalCtx, details)
}

// forTracking return the TrackingHandler and the combination of EvaluationContext from api, transaction, client and invocation.
//
// The returned evaluation context MUST be merged in the order, with duplicate values being overwritten:
//   - API (global; lowest precedence)
//   - transaction
//   - client
//   - invocation (highest precedence)
func (c *Client) forTracking(ctx context.Context, evalCtx EvaluationContext) (Tracker, EvaluationContext) {
	provider, _, globalEvalCtx := c.api.ForEvaluation(c.metadata.domain)
	evalCtx = mergeContexts(evalCtx, c.evaluationContext, EvaluationContextFromContext(ctx), globalEvalCtx)
	trackingProvider, ok := provider.(Tracker)
	if !ok {
		trackingProvider = NoopProvider{}
	}
	return trackingProvider, evalCtx
}

func evaluate[T FlagTypes](
	ctx context.Context, c *Client, flag string, flagType Type, defaultValue T, evalCtx EvaluationContext, options EvaluationOptions,
) (EvaluationDetails[T], error) {
	evalDetails := EvaluationDetails[T]{
		Value:    defaultValue,
		FlagKey:  flag,
		FlagType: flagType,
	}

	if !utf8.Valid([]byte(flag)) {
		return evalDetails, NewParseErrorResolutionError("flag key is not a UTF-8 encoded string")
	}

	// ensure that the same provider & hooks are used across this transaction to avoid unexpected behaviour
	provider, globalHooks, globalEvalCtx := c.api.ForEvaluation(c.metadata.domain)

	evalCtx = mergeContexts(evalCtx, c.evaluationContext, EvaluationContextFromContext(ctx), globalEvalCtx) // API (global) -> transaction -> client -> invocation
	hooks := slices.Concat(globalHooks, c.hooks, options.hooks, provider.Hooks())                           // API, Client, Invocation, Provider

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
		finallyHooks(ctx, hookCtx, hooks, evalDetails, options)
	}()

	ctx, evalCtx, err = beforeHooks(ctx, hookCtx, hooks, evalCtx, options)
	hookCtx.evaluationContext = evalCtx
	if err != nil {
		err = fmt.Errorf("before hook: %w", err)
		errorHooks(ctx, hookCtx, hooks, err, options)
		return evalDetails, err
	}

	// bypass short-circuit logic for the Noop provider; it is essentially stateless and a "special case"
	if _, ok := provider.(NoopProvider); !ok {
		// short circuit if provider is in NOT READY state
		if c.State() == NotReadyState {
			errorHooks(ctx, hookCtx, hooks, ErrProviderNotReady, options)
			return evalDetails, ErrProviderNotReady
		}

		// short circuit if provider is in FATAL state
		if c.State() == FatalState {
			errorHooks(ctx, hookCtx, hooks, ErrProviderFatal, options)
			return evalDetails, ErrProviderFatal
		}
	}

	flatCtx := evalCtx.Flattened()
	var resolution ObjectResolutionDetail
	switch defValue := any(defaultValue).(type) {
	case bool:
		res := provider.BooleanEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case string:
		res := provider.StringEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case float64:
		res := provider.FloatEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	case int64:
		res := provider.IntEvaluation(ctx, flag, defValue, flatCtx)
		resolution.ProviderResolutionDetail = res.ProviderResolutionDetail
		resolution.Value = res.Value
	default:
		resolution = provider.ObjectEvaluation(ctx, flag, defaultValue, flatCtx)
	}

	err = resolution.Error()
	if err != nil {
		err = fmt.Errorf("error code: %w", err)
		errorHooks(ctx, hookCtx, hooks, err, options)
		evalDetails.ResolutionDetail = resolution.ResolutionDetail()
		evalDetails.Reason = ErrorReason
		return evalDetails, err
	}

	if resolution.Value != nil {
		var ok bool
		evalDetails.Value, ok = resolution.Value.(T)
		if !ok {
			err := fmt.Errorf("evaluated value is not a %s", flagType)
			errorHooks(ctx, hookCtx, hooks, err, options)
			evalDetails.Value = defaultValue
			evalDetails.ErrorCode = TypeMismatchCode
			evalDetails.ErrorMessage = err.Error()
			return evalDetails, err
		}
	}

	evalDetails.ResolutionDetail = resolution.ResolutionDetail()

	if err := afterHooks(ctx, hookCtx, hooks, evalDetails, options); err != nil {
		err = fmt.Errorf("after hook: %w", err)
		errorHooks(ctx, hookCtx, hooks, err, options)
		return evalDetails, err
	}

	return evalDetails, nil
}

// beforeHooks executes the Before hook for each hook in the collection.
// Hooks are executed in forward order: API → Client → Invocation → Provider.
func beforeHooks(
	ctx context.Context, hookCtx HookContext, hooks []Hook, evalCtx EvaluationContext, options EvaluationOptions,
) (context.Context, EvaluationContext, error) {
	for _, hook := range hooks {
		tctx, err := hook.Before(ctx, hookCtx, options.hookHints)
		if tctx != nil {
			ctx = tctx
			if resultEvalCtx, ok := extractEvaluationContextFromContext(ctx); ok {
				hookCtx.evaluationContext = resultEvalCtx
			}
		}

		if err != nil {
			return ctx, mergeContexts(hookCtx.evaluationContext, evalCtx), err
		}
	}

	return ctx, mergeContexts(hookCtx.evaluationContext, evalCtx), nil
}

// afterHooks executes the After hook for each hook in the collection.
// Hooks are executed in reverse order: Provider → Invocation → Client → API.
func afterHooks[T FlagTypes](
	ctx context.Context, hookCtx HookContext, hooks []Hook, evalDetails EvaluationDetails[T], options EvaluationOptions,
) error {
	e := EvaluationDetails[FlagTypes]{
		FlagType:         evalDetails.FlagType,
		FlagKey:          evalDetails.FlagKey,
		Value:            evalDetails.Value,
		ResolutionDetail: evalDetails.ResolutionDetail,
	}
	// reverse order
	for _, hook := range slices.Backward(hooks) {
		if err := hook.After(ctx, hookCtx, e, options.hookHints); err != nil {
			return err
		}
	}

	return nil
}

// errorHooks executes the Error hook for each hook in the collection.
// Hooks are executed in reverse order: Provider → Invocation → Client → API.
func errorHooks(ctx context.Context, hookCtx HookContext, hooks []Hook, err error, options EvaluationOptions) {
	// reverse order
	for _, hook := range slices.Backward(hooks) {
		hook.Error(ctx, hookCtx, err, options.hookHints)
	}
}

// finallyHooks executes the Finally hook for each hook in the collection.
// Hooks are executed in reverse order: Provider → Invocation → Client → API.
func finallyHooks[T FlagTypes](ctx context.Context, hookCtx HookContext, hooks []Hook, evalDetails EvaluationDetails[T], options EvaluationOptions) {
	e := EvaluationDetails[FlagTypes]{
		FlagType:         evalDetails.FlagType,
		FlagKey:          evalDetails.FlagKey,
		Value:            evalDetails.Value,
		ResolutionDetail: evalDetails.ResolutionDetail,
	}
	// reverse order
	for _, hook := range slices.Backward(hooks) {
		hook.Finally(ctx, hookCtx, e, options.hookHints)
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
