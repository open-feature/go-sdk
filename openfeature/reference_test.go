package openfeature

import (
	"testing"
)

func TestProviderReferenceEquals(t *testing.T) {
	type myProvider struct {
		NoopProvider
		field string
	}

	type otherProvider struct {
		NoopProvider
		field string
	}

	type nonComparableProvider struct {
		NoopProvider
		data map[string]string
	}

	p1 := myProvider{}
	p2 := myProvider{}

	tests := []struct {
		name     string
		pr1      providerReference
		pr2      providerReference
		expected bool
	}{
		{
			name:     "both pointers, different instances",
			pr1:      newProviderRef(&p1),
			pr2:      newProviderRef(&p2),
			expected: false,
		},
		{
			name:     "both pointers, same instance",
			pr1:      newProviderRef(&p1),
			pr2:      newProviderRef(&p1),
			expected: true,
		},
		{
			name:     "different pointers, different instance",
			pr1:      newProviderRef(p1),
			pr2:      newProviderRef(&p1),
			expected: false,
		},
		{
			name:     "no pointers, same instance",
			pr1:      newProviderRef(p1),
			pr2:      newProviderRef(p1),
			expected: true,
		},
		{
			name:     "no pointers, different equal instances",
			pr1:      newProviderRef(myProvider{field: "A"}),
			pr2:      newProviderRef(myProvider{field: "A"}),
			expected: true,
		},
		{
			name:     "no pointers, different not equal instances",
			pr1:      newProviderRef(myProvider{field: "A"}),
			pr2:      newProviderRef(myProvider{field: "B"}),
			expected: false,
		},
		{
			name: "nil typeOf",
			pr1: providerReference{
				featureProvider:   &p1,
				typeOf:            nil,
				shutdownSemaphore: make(chan any, 1),
			},
			pr2:      newProviderRef(&p1),
			expected: false,
		},
		{
			name:     "different provider types",
			pr1:      newProviderRef(myProvider{field: "A"}),
			pr2:      newProviderRef(otherProvider{field: "A"}),
			expected: false,
		},
		{
			name:     "non-comparable, equal instances",
			pr1:      newProviderRef(nonComparableProvider{data: map[string]string{"k": "v"}}),
			pr2:      newProviderRef(nonComparableProvider{data: map[string]string{"k": "v"}}),
			expected: true,
		},
		{
			name:     "non-comparable, different instances",
			pr1:      newProviderRef(nonComparableProvider{data: map[string]string{"k": "v1"}}),
			pr2:      newProviderRef(nonComparableProvider{data: map[string]string{"k": "v2"}}),
			expected: false,
		},
		{
			name:     "non-comparable, nil vs empty map",
			pr1:      newProviderRef(nonComparableProvider{data: nil}),
			pr2:      newProviderRef(nonComparableProvider{data: map[string]string{}}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr1.equals(tt.pr2); got != tt.expected {
				t.Errorf("providerReference.equals() = %v, want %v", got, tt.expected)
			}
		})
	}
}
