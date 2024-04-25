package openfeature_test

import (
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
)

func TestNoopProvider_Metadata(t *testing.T) {
	tests := map[string]struct {
		want openfeature.Metadata
	}{
		"Given a NOOP provider, then Metadata() will return NoopProvider": {
			want: openfeature.Metadata{Name: "NoopProvider"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := openfeature.NoopProvider{}
			if got := e.Metadata(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
