package openfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
)

// The `API`, and any state it maintains SHOULD exist as a global singleton,
// even in cases wherein multiple versions of the `API` are present at runtime.
func TestRequirement_1_1_1(t *testing.T) {
	defer t.Cleanup(initSingleton)

	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProvider.EXPECT().Metadata().AnyTimes()
	SetProvider(mockProvider)

	if api.provider() != mockProvider {
		t.Error("func SetProvider hasn't set the provider to the singleton")
	}
}

// The `API` MUST provide a function to set the global `provider` singleton,
// which accepts an API-conformant `provider` implementation.
func TestRequirement_1_1_2(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockProvider := NewMockFeatureProvider(ctrl)
	mockProviderName := "mock-provider"
	mockProvider.EXPECT().Metadata().Return(Metadata{Name: mockProviderName}).AnyTimes()
	SetProvider(mockProvider)

	if ProviderMetadata() != mockProvider.Metadata() {
		t.Error("globally set provider's metadata doesn't match the mock provider's metadata")
	}
}

// The `API` MUST provide a function to add `hooks` which accepts one or more API-conformant `hooks`,
// and appends them to the collection of any previously added hooks. When new hooks are added,
// previously added hooks are not removed.
func TestRequirement_1_1_3(t *testing.T) {
	defer t.Cleanup(initSingleton)
	ctrl := gomock.NewController(t)

	mockHook := NewMockHook(ctrl)

	AddHooks(mockHook)
	AddHooks(mockHook, mockHook)

	if len(api.hooks()) != 3 {
		t.Error("func AddHooks didn't append the list of hooks to the existing collection of hooks")
	}
}

// The API MUST provide a function for retrieving the metadata field of the configured `provider`.
func TestRequirement_1_1_4(t *testing.T) {
	defer t.Cleanup(initSingleton)
	defaultProvider := NoopProvider{}
	if ProviderMetadata() != defaultProvider.Metadata() {
		t.Error("default global provider's metadata isn't NoopProvider's metadata")
	}
}

// The `API` MUST provide a function for creating a `client` which accepts the following options:
// - name (optional): A logical string identifier for the client.
func TestRequirement_1_1_5(t *testing.T) {
	defer t.Cleanup(initSingleton)
	NewClient("test-client")
}

// The client creation function MUST NOT throw, or otherwise abnormally terminate.
func TestRequirement_1_1_6(t *testing.T) {
	defer t.Cleanup(initSingleton)
	type clientCreationFunc func(name string) *Client

	// asserting that our NewClient method matches this signature is enough to deduce that no error is returned
	var f clientCreationFunc = NewClient

	use(f) // to avoid the declared and not used error
}

func use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
