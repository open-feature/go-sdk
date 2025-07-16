package multiprovider

import (
	"context"
	"strconv"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_FirstMatchStrategy_Evaluation(t *testing.T) {
	tests := []struct {
		kind       of.Type
		successVal FlagTypes
		defaultVal FlagTypes
	}{
		{kind: of.Boolean, successVal: true, defaultVal: false},
		{kind: of.Int, successVal: int64(123), defaultVal: int64(0)},
		{kind: of.String, successVal: "stringValue", defaultVal: ""},
		{kind: of.Float, successVal: float64(123.45), defaultVal: float64(0.0)},
		{kind: of.Object, successVal: struct{ Field int }{Field: 123}, defaultVal: struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			ctrl := gomock.NewController(t)

			t.Run("Single Provider Match", func(t *testing.T) {
				mocks := createMockProviders(ctrl, 1)
				configureFirstMatchProviderMock(mocks[0], tt.successVal, TestErrorNone, "mock provider")
				providers := make([]*NamedProvider, 0, 5)
				for i, m := range mocks {
					providers = append(providers, &NamedProvider{
						Name:            strconv.Itoa(i),
						FeatureProvider: m,
					})
				}
				strategy := NewFirstMatchStrategy(providers)
				result := strategy(context.Background(), "test-string", tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, providers[0].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
			})

			t.Run("Default Resolution", func(t *testing.T) {
				mocks := createMockProviders(ctrl, 1)
				configureFirstMatchProviderMock(mocks[0], tt.defaultVal, TestErrorNotFound, "mock provider")
				providers := make([]*NamedProvider, 0, 5)
				for i, m := range mocks {
					providers = append(providers, &NamedProvider{
						Name:            strconv.Itoa(i),
						FeatureProvider: m,
					})
				}
				strategy := NewFirstMatchStrategy(providers)
				result := strategy(context.Background(), "test-string", tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.defaultVal, result.Value)
				assert.Equal(t, of.DefaultReason, result.Reason)
				assert.Equal(t, of.NewFlagNotFoundResolutionError("not found in any provider").Error(), result.ResolutionError.Error())
				assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
			})

			t.Run("Evaluation stops after match", func(t *testing.T) {
				mocks := createMockProviders(ctrl, 5)
				configureFirstMatchProviderMock(mocks[0], tt.defaultVal, TestErrorNotFound, "mock provider 1")
				configureFirstMatchProviderMock(mocks[1], tt.successVal, TestErrorNone, "mock provider 2")
				providers := make([]*NamedProvider, 0, 5)
				for i, m := range mocks {
					providers = append(providers, &NamedProvider{
						Name:            strconv.Itoa(i),
						FeatureProvider: m,
					})
				}

				strategy := NewFirstMatchStrategy(providers)
				result := strategy(context.Background(), "test-flag", tt.defaultVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Contains(t, result.FlagMetadata, MetadataSuccessfulProviderName)
				assert.Equal(t, providers[1].Name, result.FlagMetadata[MetadataSuccessfulProviderName])
			})

			t.Run("Evaluation stops after first error that is not a FLAG_NOT_FOUND error", func(t *testing.T) {
				mocks := createMockProviders(ctrl, 5)
				expectedErr := of.NewGeneralResolutionError("test error")
				providers := make([]*NamedProvider, 0, 5)
				for i, m := range mocks {
					providers = append(providers, &NamedProvider{
						Name:            strconv.Itoa(i),
						FeatureProvider: m,
					})
					switch {
					case i < 3:
						configureFirstMatchProviderMock(mocks[i], tt.successVal, TestErrorNotFound, "mock provider fail")
					case i == 3:
						configureFirstMatchProviderMock(mocks[i], tt.successVal, TestErrorError, "mock provider")
					}

				}
				strategy := NewFirstMatchStrategy(providers)
				result := strategy(context.Background(), "test-string", tt.successVal, of.FlattenedContext{})
				assert.Equal(t, tt.successVal, result.Value)
				assert.Equal(t, of.ErrorReason, result.Reason)
				assert.Equal(t, expectedErr.Error(), result.ResolutionError.Error())
				assert.Equal(t, "none", result.FlagMetadata[MetadataSuccessfulProviderName])
				assert.Equal(t, StrategyFirstMatch, result.FlagMetadata[MetadataStrategyUsed])
			})
		})
	}
}

func configureFirstMatchProviderMock[R FlagTypes](mock *of.MockFeatureProvider, value R, error int, providerName string) {
	var rErr of.ResolutionError
	var reason of.Reason
	switch error {
	case TestErrorError:
		rErr = of.NewGeneralResolutionError("test error")
		reason = of.ErrorReason
	case TestErrorNotFound:
		rErr = of.NewFlagNotFoundResolutionError("test not found")
		reason = of.DefaultReason
	}

	details := of.ProviderResolutionDetail{
		ResolutionError: rErr,
		Reason:          reason,
		FlagMetadata:    make(of.FlagMetadata),
	}
	switch v := any(value).(type) {
	case bool:
		mock.EXPECT().
			BooleanEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.BoolResolutionDetail{Value: v, ProviderResolutionDetail: details})
	case string:
		mock.EXPECT().
			StringEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.StringResolutionDetail{Value: v, ProviderResolutionDetail: details})
	case int64:
		mock.EXPECT().
			IntEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.IntResolutionDetail{Value: v, ProviderResolutionDetail: details})
	case float64:
		mock.EXPECT().
			FloatEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.FloatResolutionDetail{Value: v, ProviderResolutionDetail: details})
	default:
		mock.EXPECT().
			ObjectEvaluation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(of.InterfaceResolutionDetail{Value: v, ProviderResolutionDetail: details})
	}
	mock.EXPECT().Metadata().Return(of.Metadata{Name: providerName})
}
