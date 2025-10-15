package openfeature

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestResolutionErrorWithOriginal(t *testing.T) {
	err := NewGeneralResolutionError("flag not found", io.ErrNoProgress)

	t.Run("wraps original error", func(t *testing.T) {
		if !errors.Is(err, io.ErrNoProgress) {
			t.Errorf("expected error to wrap %v", io.ErrNoProgress)
		}
	})

	t.Run("does not match unrelated error", func(t *testing.T) {
		if errors.Is(err, io.EOF) {
			t.Errorf("expected error to not match %v", io.EOF)
		}
	})

	t.Run("contains expected message", func(t *testing.T) {
		if !strings.Contains(err.Error(), "flag not found") {
			t.Errorf("expected message to contain %q, got %q", "flag not found", err.Error())
		}
	})

	t.Run("unwrap returns original", func(t *testing.T) {
		if unwrapped := errors.Unwrap(err); unwrapped != io.ErrNoProgress {
			t.Errorf("expected unwrap to return %v, got %v", io.ErrNoProgress, unwrapped)
		}
	})
}
