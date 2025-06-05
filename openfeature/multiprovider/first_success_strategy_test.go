package multiprovider

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func Test_FirstSuccessStrategy_Name(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := of.NewMockFeatureProvider(ctrl)
	strategy := NewFirstSuccessStrategy([]*NamedProvider{{Name: "test", FeatureProvider: mock}}, 0*time.Second)
	assert.Equal(t, StrategyFirstSuccess, strategy.Name())
}

func Test_FirstSuccessStrategy_Timeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock1 := of.NewMockFeatureProvider(ctrl)
	configureFirstSuccessProvider(mock1, true, true, TestErrorNone, 2*time.Second)
	mock2 := of.NewMockFeatureProvider(ctrl)
	configureFirstSuccessProvider(mock2, true, true, TestErrorNone, 2*time.Second)
	strategy := NewFirstSuccessStrategy([]*NamedProvider{{"test1", mock1}, {"test2", mock2}}, 1*time.Microsecond)
	result := strategy.BooleanEvaluation(context.Background(), "test", false, of.FlattenedContext{})
	assert.False(t, result.Value)
	assert.Errorf(t, result.Error(), "GENERAL: context deadline exceeded")

}

func configureFirstSuccessProvider[R any](provider *of.MockFeatureProvider, resultVal R, state bool, error int, delay time.Duration) {
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
			time.Sleep(delay)
			return of.BoolResolutionDetail{
				Value:                    any(resultVal).(bool),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case string:
		provider.EXPECT().StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
			time.Sleep(delay)
			return of.StringResolutionDetail{
				Value:                    any(resultVal).(string),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case int64:
		provider.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
			time.Sleep(delay)
			return of.IntResolutionDetail{
				Value:                    any(resultVal).(int64),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	case float64:
		provider.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
			time.Sleep(delay)
			return of.FloatResolutionDetail{
				Value:                    any(resultVal).(float64),
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	default:
		provider.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
			time.Sleep(delay)
			return of.InterfaceResolutionDetail{
				Value:                    resultVal,
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	}
}

func Test_FirstSuccessStrategy_BooleanEvaluation(t *testing.T) {
	t.Run("single success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider, true, true, TestErrorNone, 0*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "test-provider",
				FeatureProvider: provider,
			},
		}, 2*time.Second)
		result := strategy.BooleanEvaluation(context.Background(), testFlag, false, of.FlattenedContext{})
		assert.True(t, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("first success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, true, true, TestErrorNone, 5*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, false, false, TestErrorError, 50*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.BooleanEvaluation(context.Background(), testFlag, false, of.FlattenedContext{})
		assert.True(t, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("second success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, true, true, TestErrorNone, 500*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, false, false, TestErrorError, 5*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.BooleanEvaluation(context.Background(), testFlag, false, of.FlattenedContext{})
		assert.True(t, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("all errors (not including flag not found)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, false, false, TestErrorError, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, false, false, TestErrorError, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, false, false, TestErrorError, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.BooleanEvaluation(context.Background(), testFlag, false, of.FlattenedContext{})
		assert.False(t, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.ErrorReason, result.Reason)
	})

	t.Run("all not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, false, false, TestErrorNotFound, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, false, false, TestErrorNotFound, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, false, false, TestErrorNotFound, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.BooleanEvaluation(context.Background(), testFlag, false, of.FlattenedContext{})
		assert.False(t, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.DefaultReason, result.Reason)
	})
}

func Test_FirstSuccessStrategy_StringEvaluation(t *testing.T) {
	successVal := "success"
	defaultVal := "default"
	t.Run("single success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider, successVal, true, TestErrorNone, 0*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "test-provider",
				FeatureProvider: provider,
			},
		}, 2*time.Second)
		result := strategy.StringEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("first success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 5*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 50*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.StringEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("second success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 500*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 5*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.StringEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("all errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorError, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorError, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.StringEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.ErrorReason, result.Reason)
	})

	t.Run("all not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorNotFound, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorNotFound, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorNotFound, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.StringEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.DefaultReason, result.Reason)
	})
}

func Test_FirstSuccessStrategy_IntEvaluation(t *testing.T) {
	successVal := int64(150)
	defaultVal := int64(0)
	t.Run("single success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider, successVal, true, TestErrorNone, 0*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "test-provider",
				FeatureProvider: provider,
			},
		}, 2*time.Second)
		result := strategy.IntEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("first success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 5*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 50*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.IntEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("second success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 500*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 5*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.IntEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("all errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorError, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorError, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.IntEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.ErrorReason, result.Reason)
	})

	t.Run("all not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorNotFound, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorNotFound, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorNotFound, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.IntEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.DefaultReason, result.Reason)
	})
}

func Test_FirstSuccessStrategy_FloatEvaluation(t *testing.T) {
	successVal := float64(15.5)
	defaultVal := float64(0)
	t.Run("single success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider, successVal, true, TestErrorNone, 0*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "test-provider",
				FeatureProvider: provider,
			},
		}, 2*time.Second)
		result := strategy.FloatEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("first success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 5*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 50*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.FloatEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("second success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 500*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 5*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.FloatEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("all errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorError, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorError, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.FloatEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.ErrorReason, result.Reason)
	})

	t.Run("all not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorNotFound, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorNotFound, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorNotFound, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.FloatEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.DefaultReason, result.Reason)
	})
}

func Test_FirstSuccessStrategy_ObjectEvaluation(t *testing.T) {
	successVal := struct{ Field string }{Field: "test"}
	defaultVal := struct{}{}
	t.Run("single success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider, successVal, true, TestErrorNone, 0*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "test-provider",
				FeatureProvider: provider,
			},
		}, 2*time.Second)
		result := strategy.ObjectEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("first success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 5*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 50*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.ObjectEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("second success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, successVal, true, TestErrorNone, 500*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 5*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
			{
				Name:            "success-provider",
				FeatureProvider: provider1,
			},
			{
				Name:            "failure-provider",
				FeatureProvider: provider2,
			},
		}, 2*time.Second)

		result := strategy.ObjectEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "success-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("all errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorError, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorError, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorError, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.ObjectEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.ErrorReason, result.Reason)
	})

	t.Run("all not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider1, defaultVal, false, TestErrorNotFound, 50*time.Millisecond)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider2, defaultVal, false, TestErrorNotFound, 40*time.Millisecond)
		provider3 := of.NewMockFeatureProvider(ctrl)
		configureFirstSuccessProvider(provider3, defaultVal, false, TestErrorNotFound, 30*time.Millisecond)

		strategy := NewFirstSuccessStrategy([]*NamedProvider{
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
		}, 2*time.Second)

		result := strategy.ObjectEvaluation(context.Background(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, StrategyFirstSuccess, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, of.DefaultReason, result.Reason)
	})
}
