package openfeature

import (
	"github.com/golang/mock/gomock"
	"testing"
)

func TestRequirement_1_1_1(t *testing.T) {
	defer t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	SetProvider(mockProvider)

	if api.provider != mockProvider {
		t.Error("func SetProvider hasn't set the provider to the singleton")
	}
}

func TestRequirement_1_1_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProviderName := "mock-provider"
	SetProvider(mockProvider)
	mockProvider.EXPECT().Metadata().Return(Metadata{Name: mockProviderName}).Times(2)

	if ProviderMetadata() != mockProvider.Metadata() {
		t.Error("globally set provider's metadata doesn't match the mock provider's metadata")
	}
}

func TestRequirement_1_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	AddHooks(mockHook)
	AddHooks(mockHook, mockHook)

	if len(api.hooks) != 3 {
		t.Error("func AddHooks didn't append the list of hooks to the existing collection of hooks")
	}
}

func TestRequirement_1_1_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	defaultProvider := NoopProvider{}
	if ProviderMetadata() != defaultProvider.Metadata() {
		t.Error("default global provider's metadata isn't NoopProvider's metadata")
	}
}

func TestRequirement_1_1_5(t *testing.T) {
	defer t.Cleanup(initSingleton)
	GetClient("test-client")
}

func TestRequirement_1_1_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	type clientCreationFunc func(name string)*Client

	var f clientCreationFunc
	f = GetClient // asserting that our GetClient method matches this signature is enough to deduce that no error is returned

	use(f) // to avoid the declared and not used error
}

func use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
