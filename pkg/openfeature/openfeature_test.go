package openfeature

import (
	"github.com/golang/mock/gomock"
	"testing"
)

func TestRequirement1_1_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProviderName := "mock-provider"
	SetProvider(mockProvider)
	mockProvider.EXPECT().Metadata().Return(Metadata{Name: mockProviderName}).Times(2)

	if ProviderMetadata() != mockProvider.Metadata() {
		t.Errorf("Globally set provider's metadata doesn't match the mock provider's metadata")
	}
}

func TestRequirement1_1_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	defaultProvider := NoopProvider{}
	if ProviderMetadata() != defaultProvider.Metadata() {
		t.Errorf("Default global provider's metadata isn't NoopProvider's metadata")
	}
}
