package multiprovider

import (
	of "github.com/open-feature/go-sdk/openfeature"
	"go.uber.org/mock/gomock"
)

func createMockProviders(ctrl *gomock.Controller, count int) []*of.MockFeatureProvider {
	providerMocks := make([]*of.MockFeatureProvider, 0, count)
	for i := 0; i < count; i++ {
		provider := of.NewMockFeatureProvider(ctrl)
		providerMocks = append(providerMocks, provider)
	}

	return providerMocks
}

const testFlag = "test-flag"

const (
	TestErrorNone     = 0
	TestErrorNotFound = 1
	TestErrorError    = 2
)
