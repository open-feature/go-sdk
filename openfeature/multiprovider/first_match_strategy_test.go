package multiprovider

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"strconv"
	"testing"
)

func Test_FirstMatchStrategy_BooleanEvaluation(t *testing.T) {
	t.Run("Single Provider Match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{Value: true, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.BooleanEvaluation(context.Background(), "test-string", false, of.FlattenedContext{})
		assert.True(t, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Default Resolution", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("not found in any provider"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.BooleanEvaluation(context.Background(), "test-string", false, of.FlattenedContext{})
		assert.False(t, result.Value)
		assert.Equal(t, of.DefaultReason, result.Reason)
		assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})

	t.Run("Evaluation stops after match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mocks := createMockProviders(ctrl, 5)
		mocks[0].EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("Flag not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		mocks[1].EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{Value: true, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[1].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.BooleanEvaluation(context.Background(), "test-flag", false, of.FlattenedContext{})
		assert.True(t, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mocks := createMockProviders(ctrl, 5)
		expectedErr := of.NewGeneralResolutionError("something went wrong")
		mocks[0].EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: expectedErr,
					Reason:          of.ErrorReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.BooleanEvaluation(context.Background(), "test-string", false, of.FlattenedContext{})
		assert.False(t, result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})
}

func Test_FirstMatchStrategy_StringEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("Single Provider Match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{Value: "test", ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.StringEvaluation(context.Background(), "test-string", "", of.FlattenedContext{})
		assert.Equal(t, "test", result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Default Resolution", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{
				Value: "",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.StringEvaluation(context.Background(), "test-string", "", of.FlattenedContext{})
		assert.Equal(t, "", result.Value)
		assert.Equal(t, of.DefaultReason, result.Reason)
		assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})

	t.Run("Evaluation stops after match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		mocks[0].EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{
				Value: "",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("Flag not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		mocks[1].EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{Value: "test", ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[1].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}

		strategy := NewFirstMatchStrategy(providers)
		result := strategy.StringEvaluation(context.Background(), "test-flag", "", of.FlattenedContext{})
		assert.Equal(t, "test", result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		expectedErr := of.NewGeneralResolutionError("something went wrong")
		mocks[0].EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{
				Value: "",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: expectedErr,
					Reason:          of.ErrorReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.StringEvaluation(context.Background(), "test-string", "", of.FlattenedContext{})
		assert.Equal(t, "", result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})
}

func Test_FirstMatchStrategy_IntEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("Single Provider Match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{Value: 123, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.IntEvaluation(context.Background(), "test-string", 0, of.FlattenedContext{})
		assert.Equal(t, int64(123), result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Default Resolution", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.IntEvaluation(context.Background(), "test-string", 0, of.FlattenedContext{})
		assert.Equal(t, int64(0), result.Value)
		assert.Equal(t, of.DefaultReason, result.Reason)
		assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})

	t.Run("Evaluation stops after match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		mocks[0].EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("Flag not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		mocks[1].EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{Value: 123, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[1].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}

		strategy := NewFirstMatchStrategy(providers)
		result := strategy.IntEvaluation(context.Background(), "test-flag", 0, of.FlattenedContext{})
		assert.Equal(t, int64(123), result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		expectedErr := of.NewGeneralResolutionError("something went wrong")
		mocks[0].EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{
				Value: 123,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: expectedErr,
					Reason:          of.ErrorReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.IntEvaluation(context.Background(), "test-string", 123, of.FlattenedContext{})
		assert.Equal(t, int64(123), result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})
}

func Test_FirstMatchStrategy_FloatEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("Single Provider Match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{Value: 123, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.FloatEvaluation(context.Background(), "test-string", 0, of.FlattenedContext{})
		assert.Equal(t, float64(123), result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Default Resolution", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.FloatEvaluation(context.Background(), "test-string", 0, of.FlattenedContext{})
		assert.Equal(t, float64(0), result.Value)
		assert.Equal(t, of.DefaultReason, result.Reason)
		assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})

	t.Run("Evaluation stops after match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		mocks[0].EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("Flag not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		mocks[1].EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{Value: 123, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[1].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}

		strategy := NewFirstMatchStrategy(providers)
		result := strategy.FloatEvaluation(context.Background(), "test-flag", 0, of.FlattenedContext{})
		assert.Equal(t, float64(123), result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		expectedErr := of.NewGeneralResolutionError("something went wrong")
		mocks[0].EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{
				Value: 123.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: expectedErr,
					Reason:          of.ErrorReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.FloatEvaluation(context.Background(), "test-string", 123, of.FlattenedContext{})
		assert.Equal(t, 123.0, result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})
}

func Test_FirstMatchStrategy_ObjectEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("Single Provider Match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{Value: struct{ Field int }{Field: 123}, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.ObjectEvaluation(context.Background(), "test-string", struct{}{}, of.FlattenedContext{})
		assert.Equal(t, struct{ Field int }{Field: 123}, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Default Resolution", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 1)
		mocks[0].EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{
				Value: struct{}{},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("not found"),
					Reason:          of.DefaultReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.ObjectEvaluation(context.Background(), "test-string", struct{}{}, of.FlattenedContext{})
		assert.Equal(t, struct{}{}, result.Value)
		assert.Equal(t, of.DefaultReason, result.Reason)
		assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})

	t.Run("Evaluation stops after match", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		mocks[0].EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewFlagNotFoundResolutionError("Flag not found"),
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		mocks[1].EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{Value: struct{ Field int }{Field: 123}, ProviderResolutionDetail: of.ProviderResolutionDetail{FlagMetadata: make(of.FlagMetadata)}})
		mocks[1].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}

		strategy := NewFirstMatchStrategy(providers)
		result := strategy.ObjectEvaluation(context.Background(), "test-flag", struct{}{}, of.FlattenedContext{})
		assert.Equal(t, struct{ Field int }{Field: 123}, result.Value)
		assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
		assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
	})

	t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
		mocks := createMockProviders(ctrl, 5)
		expectedErr := of.NewGeneralResolutionError("something went wrong")
		mocks[0].EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{
				Value: struct{}{},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: expectedErr,
					Reason:          of.ErrorReason,
					FlagMetadata:    make(of.FlagMetadata),
				},
			})
		mocks[0].EXPECT().Metadata().Return(of.Metadata{Name: "mock provider"})
		providers := make([]*NamedProvider, 0, 5)
		for i, m := range mocks {
			providers = append(providers, &NamedProvider{
				Name:            strconv.Itoa(i),
				FeatureProvider: m,
			})
		}
		strategy := NewFirstMatchStrategy(providers)
		result := strategy.ObjectEvaluation(context.Background(), "test-string", struct{}{}, of.FlattenedContext{})
		assert.Equal(t, struct{}{}, result.Value)
		assert.Equal(t, of.ErrorReason, result.Reason)
		assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
		assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
		assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
	})
}
