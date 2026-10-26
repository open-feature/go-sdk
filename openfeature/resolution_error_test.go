package openfeature

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewResolutionError(t *testing.T) {
	t.Run("preserves arbitrary code", func(t *testing.T) {
		err := NewResolutionError(ErrorCode("CUSTOM_CODE"), "custom")
		if err.code != ErrorCode("CUSTOM_CODE") {
			t.Errorf("expected code %q, got %q", "CUSTOM_CODE", err.code)
		}
		if err.message != "custom" {
			t.Errorf("expected message %q, got %q", "custom", err.message)
		}
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			t.Errorf("expected no wrapped error, got %v", unwrapped)
		}
	})

	t.Run("unwrap with no wrapped errors", func(t *testing.T) {
		err := NewResolutionError(GeneralCode, "no cause")
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			t.Errorf("expected nil, got %v", unwrapped)
		}
	})

	t.Run("unwrap with one wrapped error", func(t *testing.T) {
		err := NewResolutionError(GeneralCode, "one cause", io.EOF)
		if unwrapped := errors.Unwrap(err); unwrapped != io.EOF {
			t.Errorf("expected %v, got %v", io.EOF, unwrapped)
		}
		if !errors.Is(err, io.EOF) {
			t.Errorf("expected error to match %v", io.EOF)
		}
	})

	t.Run("unwrap with several wrapped errors", func(t *testing.T) {
		err := NewResolutionError(GeneralCode, "several causes", io.EOF, io.ErrNoProgress)
		for _, want := range []error{io.EOF, io.ErrNoProgress} {
			if !errors.Is(err, want) {
				t.Errorf("expected error to match %v", want)
			}
		}
		if unwrapped := errors.Unwrap(err); unwrapped == nil {
			t.Error("expected joined error, got nil")
		}
	})
}

func TestResolutionErrorWithOriginal(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"general", NewGeneralResolutionError("flag not found", io.ErrNoProgress)},
		{"parse", NewParseErrorResolutionError("flag not found", io.ErrNoProgress)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.err
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
		})
	}
}
