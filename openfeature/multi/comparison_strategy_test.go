package multi

import (
	"context"
	"fmt"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func configureComparisonProvider[R any](provider *of.MockFeatureProvider, resultVal R, state bool, error int, forceObj bool) {
	var rErr of.ResolutionError
	var variant string
	var reason of.Reason
	switch error {
	case TestErrorError:
		rErr = of.NewGeneralResolutionError("test error")
		reason = of.ErrorReason
	case TestErrorNotFound:
		rErr = of.NewFlagNotFoundResolutionError("not found")
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
	objFunc := func(p *of.MockFeatureProvider) {
		p.EXPECT().ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(c context.Context, flag string, defaultVal any, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
			return of.InterfaceResolutionDetail{
				Value:                    resultVal,
				ProviderResolutionDetail: details,
			}
		}).MaxTimes(1)
	}

	if forceObj {
		objFunc(provider)
		return
	}

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
		objFunc(provider)
	}
}

func Test_ComparisonStrategy_Evaluation(t *testing.T) {
	tests := []struct {
		kind       of.Type
		successVal FlagTypes
		defaultVal FlagTypes
	}{
		{of.Boolean, true, false},
		{of.String, "success", "default"},
		{of.Int, int64(1234), int64(0)},
		{of.Float, float64(12.34), float64(0)},
		{of.Object, struct{ Field string }{Field: "test"}, struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			successVal := tt.successVal
			defaultVal := tt.defaultVal
			t.Run("single success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider := of.NewMockFeatureProvider(ctrl)
				fallback := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider",
						FeatureProvider: provider,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "test-provider", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("two success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, successVal, true, TestErrorNone, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Equal(t, "test-provider1, test-provider2", result.FlagMetadata[MetadataSuccessfulProviderNames])
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("multiple success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				fallback.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, successVal, true, TestErrorNone, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, successVal, true, TestErrorNone, false)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider3, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
					&namedProvider{
						name:            "test-provider3",
						FeatureProvider: provider3,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Equal(t, "test-provider1, test-provider2, test-provider3", result.FlagMetadata[MetadataSuccessfulProviderNames])
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("multiple not found with single success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, defaultVal, true, TestErrorNotFound, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorNotFound, false)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider3, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
					&namedProvider{
						name:            "test-provider3",
						FeatureProvider: provider3,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Equal(t, "test-provider3", result.FlagMetadata[MetadataSuccessfulProviderNames])
				assert.Contains(t, result.FlagMetadata, MetadataFallbackUsed)
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("multiple not found with multiple success", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				fallback.EXPECT().IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, defaultVal, true, TestErrorNotFound, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorNotFound, false)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider3, successVal, true, TestErrorNone, false)
				provider4 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider4, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
					&namedProvider{
						name:            "test-provider3",
						FeatureProvider: provider3,
					},
					&namedProvider{
						name:            "test-provider4",
						FeatureProvider: provider4,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Equal(t, "test-provider3, test-provider4", result.FlagMetadata[MetadataSuccessfulProviderNames])
				assert.Contains(t, result.FlagMetadata, MetadataFallbackUsed)
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("comparison failure uses fallback", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(fallback, successVal, true, TestErrorNone, false)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, defaultVal, true, TestErrorNone, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorNone, false)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider3, successVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
					&namedProvider{
						name:            "test-provider3",
						FeatureProvider: provider3,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.NotContains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "fallback", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.True(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("not found all providers", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				fallback.EXPECT().FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, defaultVal, true, TestErrorNotFound, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorNotFound, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, defaultVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.NotContains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.Contains(t, result.FlagMetadata, MetadataFallbackUsed)
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("comparison failure with not found", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(fallback, successVal, true, TestErrorNone, false)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, defaultVal, true, TestErrorNotFound, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorNotFound, false)
				provider3 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider3, successVal, true, TestErrorNone, false)
				provider4 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider4, defaultVal, true, TestErrorNone, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
					&namedProvider{
						name:            "test-provider3",
						FeatureProvider: provider3,
					},
					&namedProvider{
						name:            "test-provider4",
						FeatureProvider: provider4,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.NotContains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "fallback", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.Contains(t, result.FlagMetadata, MetadataFallbackUsed)
				assert.True(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})

			t.Run("non FLAG_NOT_FOUND error causes default", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				fallback := of.NewMockFeatureProvider(ctrl)
				provider1 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider1, successVal, true, TestErrorError, false)
				provider2 := of.NewMockFeatureProvider(ctrl)
				configureComparisonProvider(provider2, defaultVal, true, TestErrorError, false)

				strategy := newComparisonStrategy([]NamedProvider{
					&namedProvider{
						name:            "test-provider1",
						FeatureProvider: provider1,
					},
					&namedProvider{
						name:            "test-provider2",
						FeatureProvider: provider2,
					},
				}, fallback, nil)

				result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
				assert.Equal(t, defaultVal, result.Value)
				assert.Equal(t, of.ErrorReason, result.Reason)
				assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
				assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
				assert.NotContains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
				assert.Contains(t, result.FlagMetadata, MetadataEvaluationError)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
			})
		})
	}
}

func Test_ComparisonStrategy_ObjectEvaluation(t *testing.T) {
	successVal := struct{ Name string }{Name: "test"}
	defaultVal := struct{}{}

	type orderableTestCase struct {
		typeName     string
		successValue any
		defaultValue any
	}

	orderableTests := []orderableTestCase{
		{
			typeName:     "int8",
			successValue: int8(5),
			defaultValue: int8(1),
		},
		{
			typeName:     "int16",
			successValue: int16(5),
			defaultValue: int16(1),
		},
		{
			typeName:     "int32",
			successValue: int32(5),
			defaultValue: int32(1),
		},
		// {
		// 	typeName:     "int64",
		// 	successValue: int64(5),
		// 	defaultValue: int64(1),
		// },
		{
			typeName:     "int",
			successValue: 5,
			defaultValue: 1,
		},
		{
			typeName:     "uint8",
			successValue: uint8(5),
			defaultValue: uint8(1),
		},
		{
			typeName:     "uint16",
			successValue: uint16(5),
			defaultValue: uint16(1),
		},
		{
			typeName:     "uint32",
			successValue: uint32(5),
			defaultValue: uint32(1),
		},
		{
			typeName:     "uint64",
			successValue: uint64(5),
			defaultValue: uint64(1),
		},
		{
			typeName:     "uint",
			successValue: uint(5),
			defaultValue: uint(1),
		},
		{
			typeName:     "uintptr",
			successValue: uintptr(5),
			defaultValue: uintptr(1),
		},
		{
			typeName:     "float32",
			successValue: float32(5.5),
			defaultValue: float32(1.1),
		},
		// {
		// 	typeName:     "float64",
		// 	successValue: 5.5,
		// 	defaultValue: 1.1,
		// },
		// {
		// 	typeName:     "string",
		// 	successValue: "success",
		// 	defaultValue: "default",
		// },
	}

	for _, testCase := range orderableTests {
		tc := testCase
		t.Run(fmt.Sprintf("with orderable type %s success", tc.typeName), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fallback := of.NewMockFeatureProvider(ctrl)
			provider1 := of.NewMockFeatureProvider(ctrl)
			configureComparisonProvider(provider1, testCase.successValue, true, TestErrorNone, true)
			provider2 := of.NewMockFeatureProvider(ctrl)
			configureComparisonProvider(provider2, testCase.successValue, true, TestErrorNone, true)

			strategy := newComparisonStrategy([]NamedProvider{
				&namedProvider{
					name:            "test-provider1",
					FeatureProvider: provider1,
				},
				&namedProvider{
					name:            "test-provider2",
					FeatureProvider: provider2,
				},
			}, fallback, nil)

			result := strategy(t.Context(), testFlag, tc.defaultValue, of.FlattenedContext{})
			assert.Equal(t, tc.successValue, result.Value)
			assert.NoError(t, result.Error())
			assert.Equal(t, ReasonAggregated, result.Reason)
			assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
			assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
			assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
			assert.Equal(t, "test-provider1, test-provider2", result.FlagMetadata[MetadataSuccessfulProviderNames])
			assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
		})

		t.Run(fmt.Sprintf("with orderable type %s no match fallback", tc.typeName), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fallback := of.NewMockFeatureProvider(ctrl)
			configureComparisonProvider(fallback, tc.successValue, true, TestErrorNone, true)
			provider1 := of.NewMockFeatureProvider(ctrl)
			configureComparisonProvider(provider1, tc.successValue, true, TestErrorNone, true)
			provider2 := of.NewMockFeatureProvider(ctrl)
			configureComparisonProvider(provider2, tc.defaultValue, true, TestErrorNone, true)

			strategy := newComparisonStrategy([]NamedProvider{
				&namedProvider{
					name:            "test-provider1",
					FeatureProvider: provider1,
				},
				&namedProvider{
					name:            "test-provider2",
					FeatureProvider: provider2,
				},
			}, fallback, nil)
			result := strategy(t.Context(), testFlag, tc.defaultValue, of.FlattenedContext{})
			assert.Equal(t, tc.successValue, result.Value)
			assert.NoError(t, result.Error())
			assert.Equal(t, ReasonAggregatedFallback, result.Reason)
			assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
			assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
			assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
			assert.Equal(t, "fallback", result.FlagMetadata[MetadataSuccessfulProviderName])
			assert.True(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
		})
	}

	t.Run("with comparable custom type success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, successVal, true, TestErrorNone, true)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, successVal, true, TestErrorNone, true)

		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, nil)

		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.NoError(t, result.Error())
		assert.Equal(t, ReasonAggregated, result.Reason)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
		assert.Equal(t, "test-provider1, test-provider2", result.FlagMetadata[MetadataSuccessfulProviderNames])
		assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})

	t.Run("with comparable custom type no match fallback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(fallback, successVal, true, TestErrorNone, true)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, successVal, true, TestErrorNone, true)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, defaultVal, true, TestErrorNone, true)

		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, nil)
		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.NoError(t, result.Error())
		assert.Equal(t, ReasonAggregatedFallback, result.Reason)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "fallback", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.True(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})

	t.Run("with comparable custom type force custom comparator", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(fallback, defaultVal, true, TestErrorNone, true)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, successVal, true, TestErrorNone, true)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, successVal, true, TestErrorNone, true)

		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, func(val []any) bool {
			return true
		})
		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.NoError(t, result.Error())
		assert.Equal(t, ReasonAggregated, result.Reason)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
		assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})

	t.Run("with non comparable types using custom comparator success", func(t *testing.T) {
		successVal := []string{"test1", "test2"}
		defaultVal := []string{"test3"}
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, successVal, true, TestErrorNone, false)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, successVal, true, TestErrorNone, false)

		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, func(val []any) bool {
			return true
		})

		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.NoError(t, result.Error())
		assert.Equal(t, ReasonAggregated, result.Reason)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
		assert.Equal(t, "test-provider1, test-provider2", result.FlagMetadata[MetadataSuccessfulProviderNames])
		assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})

	t.Run("with non comparable types using custom comparator no match fallback", func(t *testing.T) {
		successVal := []string{"test1", "test2"}
		defaultVal := []string{"test3"}
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(fallback, successVal, true, TestErrorNone, false)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, defaultVal, true, TestErrorNone, false)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, defaultVal, true, TestErrorNone, false)

		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, func(val []any) bool {
			return false
		})
		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, successVal, result.Value)
		assert.NoError(t, result.Error())
		assert.Equal(t, ReasonAggregatedFallback, result.Reason)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "fallback", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.True(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})

	t.Run("any provider error bypasses comparison", func(t *testing.T) {
		successVal := []string{"test1", "test2"}
		defaultVal := []string{"test3"}
		ctrl := gomock.NewController(t)
		fallback := of.NewMockFeatureProvider(ctrl)
		provider1 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider1, successVal, true, TestErrorNone, false)
		provider2 := of.NewMockFeatureProvider(ctrl)
		configureComparisonProvider(provider2, successVal, true, TestErrorError, false)
		strategy := newComparisonStrategy([]NamedProvider{
			&namedProvider{
				name:            "test-provider1",
				FeatureProvider: provider1,
			},
			&namedProvider{
				name:            "test-provider2",
				FeatureProvider: provider2,
			},
		}, fallback, nil)
		result := strategy(t.Context(), testFlag, defaultVal, of.FlattenedContext{})
		assert.Equal(t, defaultVal, result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, of.NewGeneralResolutionError(ErrAggregationNotAllowed.Error()), result.ResolutionError)
		assert.Contains(t, result.FlagMetadata, MetadataStrategyUsed)
		assert.Equal(t, StrategyComparison, result.FlagMetadata[MetadataStrategyUsed])
		assert.NotContains(t, result.FlagMetadata, MetadataSuccessfulProviderNames)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.False(t, result.FlagMetadata[MetadataFallbackUsed].(bool))
	})
}
