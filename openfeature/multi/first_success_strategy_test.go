package multi

import (
	"context"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func configureFirstSuccessProvider[R any](provider *of.MockFeatureProvider, resultVal R, state bool, error int) {
	var rErr of.ResolutionError
	var variant string
	var reason of.Reason
	switch error {
	case TestErrorError:
		rErr = of.NewGeneralResolutionError("test error")
		reason = of.ErrorReason
	case TestErrorNotFound:
		rErr = of.NewFlagNotFoundResolutionError("test not found")
		reason = of.DefaultReason
	}

	if state {
		variant = "on"
	} else {
		variant = "off"
	}
	details := of.ProviderResolutionDetail{
		ResolutionError: rErr,
		Reason:          reason,
		Variant:         variant,
		FlagMetadata:    make(of.FlagMetadata),
	}

	provider.EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"}).MaxTimes(1)
	switch any(resultVal).(type) {
	case bool:
		provider.EXPECT().BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
			return of.BoolResolutionDetail{
				Value:                    any(resultVal).(bool),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case string:
		provider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
			return of.StringResolutionDetail{
				Value:                    any(resultVal).(string),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case int64:
		provider.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
			return of.IntResolutionDetail{
				Value:                    any(resultVal).(int64),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case float64:
		provider.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
			return of.FloatResolutionDetail{
				Value:                    any(resultVal).(float64),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	default:
		provider.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
			return of.InterfaceResolutionDetail{
				Value:                    resultVal,
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	}
}

func Test_FirstSuccessStrategyEvaluation(t *testing.T) {
	tests := []struct {
		kind       of.Type
		successVal FlagTypes
		defaultVal FlagTypes
	}{
		{kind: of.Boolean, successVal: true, defaultVal: false},
		{kind: of.String, successVal: "success", defaultVal: "default"},
		{kind: of.Int, successVal: int64(150), defaultVal: int64(0)},
		{kind: of.Float, successVal: float64(15.5), defaultVal: float64(0)},
		{kind: of.Object, successVal: struct{ Field string }{Field: "test"}, defaultVal: struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			t.Run("single success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider, tt.successVal, true, TestErrorNone)

				strategy := newFirstSuccessStrategy([]*NamedProvider{
					{
						Name:            "test-provider",
						FeatureProvider: provider,
					},
				})
				result := strategy(context.Background(), testFlag, tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
			})

			t.Run("first success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider1, tt.successVal, true, TestErrorNone)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider2, tt.defaultVal, false, TestErrorError)

				strategy := newFirstSuccessStrategy([]*NamedProvider{
					{
						Name:            "success-provider",
						FeatureProvider: provider1,
					},
					{
						Name:            "failure-provider",
						FeatureProvider: provider2,
					},
				})

				result := strategy(context.Background(), testFlag, tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
			})

			t.Run("second success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider1, tt.successVal, true, TestErrorNone)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider2, tt.defaultVal, false, TestErrorError)

				strategy := newFirstSuccessStrategy([]*NamedProvider{
					{
						Name:            "success-provider",
						FeatureProvider: provider1,
					},
					{
						Name:            "failure-provider",
						FeatureProvider: provider2,
					},
				})

				result := strategy(context.Background(), testFlag, tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
			})

			t.Run("all errors", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider1, tt.defaultVal, false, TestErrorError)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider2, tt.defaultVal, false, TestErrorNotFound)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureFirstSuccessProvider(provider3, tt.defaultVal, false, TestErrorError)

				strategy := newFirstSuccessStrategy([]*NamedProvider{
					{
						Name:            "provider1",
						FeatureProvider: provider1,
					},
					{
						Name:            "provider2",
						FeatureProvider: provider2,
					},
					{
						Name:            "provider3",
						FeatureProvider: provider3,
					},
				})

				result := strategy(context.Background(), testFlag, tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.defaultVal, result.Value)
				assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.Equal(t, of.ErrorReason, result.Reason)
				assert.NotNil(t, result.Error())
			})
		})
	}
}
