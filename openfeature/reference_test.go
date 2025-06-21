package openfeature

import (
	"testing"
)

func TestProviderReferenceEquals(t *testing.T) {
	type myProvider struct {
		NoopProvider
		field string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr1.equals(tt.pr2); got != tt.expected {
				t.Errorf("providerReference.equals() = %v, want %v", got, tt.expected)
			}
		})
	}
}
