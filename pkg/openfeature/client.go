package openfeature

// IClient defines the behaviour required of an openfeature client
type IClient interface {
	Metadata() ClientMetadata
	AddHooks(hooks ...Hook)
	BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (bool, error)
	StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (string, error)
	NumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (float64, error)
	ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (interface{}, error)
	BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
	StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
	NumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
	ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error)
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

// Type represents the type of a flg
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

// BooleanValue return boolean evaluation for flag
func (c Client) BooleanValue(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (bool, error) {
	resolution := api.provider.BooleanEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return value, err
}

// StringValue return string evaluation for flag
func (c Client) StringValue(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (string, error) {
	resolution := api.provider.StringEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return value, err
}

// NumberValue return number evaluation for flag
func (c Client) NumberValue(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (float64, error) {
	resolution := api.provider.NumberEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return value, err
}

// ObjectValue return object evaluation for flag
func (c Client) ObjectValue(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (interface{}, error) {
	resolution := api.provider.ObjectEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return value, err
}

// BooleanValueDetails return boolean evaluation for flag
func (c Client) BooleanValueDetails(flag string, defaultValue bool, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error) {
	resolution := api.provider.BooleanEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return EvaluationDetails{
		FlagKey:  flag,
		FlagType: Boolean,
		ResolutionDetail: ResolutionDetail{
			Value:     value,
			ErrorCode: resolution.ErrorCode,
			Reason:    resolution.Reason,
			Variant:   resolution.Variant,
		},
	}, err

}

// StringValueDetails return string evaluation for flag
func (c Client) StringValueDetails(flag string, defaultValue string, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error) {
	resolution := api.provider.StringEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return EvaluationDetails{
		FlagKey:  flag,
		FlagType: String,
		ResolutionDetail: ResolutionDetail{
			Value:     value,
			ErrorCode: resolution.ErrorCode,
			Reason:    resolution.Reason,
			Variant:   resolution.Variant,
		},
	}, err
}

// NumberValueDetails return number evaluation for flag
func (c Client) NumberValueDetails(flag string, defaultValue float64, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error) {
	resolution := api.provider.NumberEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return EvaluationDetails{
		FlagKey:  flag,
		FlagType: Number,
		ResolutionDetail: ResolutionDetail{
			Value:     value,
			ErrorCode: resolution.ErrorCode,
			Reason:    resolution.Reason,
			Variant:   resolution.Variant,
		},
	}, err
}

// ObjectValueDetails return object evaluation for flag
func (c Client) ObjectValueDetails(flag string, defaultValue interface{}, evalCtx EvaluationContext, options ...EvaluationOption) (EvaluationDetails, error) {
	resolution := api.provider.ObjectEvaluation(flag, defaultValue, evalCtx, options...)
	value := resolution.Value
	err := resolution.Error()
	if err != nil {
		value = defaultValue
	}
	return EvaluationDetails{
		FlagKey:  flag,
		FlagType: Object,
		ResolutionDetail: ResolutionDetail{
			Value:     value,
			ErrorCode: resolution.ErrorCode,
			Reason:    resolution.Reason,
			Variant:   resolution.Variant,
		},
	}, err
}
