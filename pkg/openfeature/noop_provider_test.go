package openfeature

import (
	"testing"
)

func TestNoopProvider_Name(t *testing.T) {
	tests := map[string]struct {
		want string
	}{
		"Given a NOOP provider, then Name() will return NoopProvider": {
			want: "NoopProvider",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := NoopProvider{}
			if got := e.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
