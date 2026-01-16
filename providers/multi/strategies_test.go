package multi

import (
	of "go.openfeature.dev/openfeature/v2"
	"go.uber.org/mock/gomock"
)

func createMockProviders(ctrl *gomock.Controller, count int) []*of.MockProvider {
	providerMocks := make([]*of.MockProvider, 0, count)
	for range count {
		provider := of.NewMockProvider(ctrl)
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
