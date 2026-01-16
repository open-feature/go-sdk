package multi

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ProviderError_Error(t *testing.T) {
	t.Run("with nil err", func(t *testing.T) {
		err := &ProviderError{
			ProviderName: "TestError",
		}
		assert.EqualError(t, err, "Provider TestError: <nil>")
		assert.Equal(t, "TestError", err.ProviderName)
	})

	t.Run("with custom error", func(t *testing.T) {
		originalErr := errors.New("custom error message")
		err := &ProviderError{
			ProviderName: "TestError",
			err:          originalErr,
		}
		assert.EqualError(t, err, "Provider TestError: custom error message")
		assert.Equal(t, "TestError", err.ProviderName)
		assert.ErrorIs(t, err, originalErr)
	})
}

func Test_AggregateError_Error(t *testing.T) {
	t.Run("empty error", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{})
		assert.Empty(t, err.Error())
	})

	t.Run("single error", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{
			{
				err:          fmt.Errorf("test error"),
				ProviderName: "test-provider",
			},
		})

		assert.Equal(t, "Provider test-provider: test error", err.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{
			{
				err:          fmt.Errorf("test error"),
				ProviderName: "test-provider1",
			},
			{
				err:          fmt.Errorf("test error"),
				ProviderName: "test-provider2",
			},
		})

		assert.Equal(t, "Provider test-provider1: test error\nProvider test-provider2: test error", err.Error())
	})
}

func Test_multiErrGroup(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*multiErrGroup)
		wantErrs []string
	}{
		{
			name: "no errors - all goroutines succeed",
			setup: func(meg *multiErrGroup) {
				for range 3 {
					meg.Go(func() error {
						time.Sleep(10 * time.Millisecond)
						return nil
					})
				}
			},
			wantErrs: nil,
		},
		{
			name: "single error among successful goroutines",
			setup: func(meg *multiErrGroup) {
				meg.Go(func() error {
					return errors.New("error 0")
				})
				meg.Go(func() error {
					return nil
				})
			},
			wantErrs: []string{"error 0"},
		},
		{
			name: "multiple errors collected",
			setup: func(meg *multiErrGroup) {
				meg.Go(func() error {
					return errors.New("error 1")
				})
				meg.Go(func() error {
					return nil
				})
				meg.Go(func() error {
					return errors.New("error 2")
				})
				meg.Go(func() error {
					return errors.New("error 3")
				})
			},
			wantErrs: []string{"error 1", "error 2", "error 3"},
		},
		{
			name:     "empty group returns no error",
			setup:    func(meg *multiErrGroup) {},
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var meg multiErrGroup
			tt.setup(&meg)

			err := meg.Wait()

			if tt.wantErrs != nil {
				require.Error(t, err)
				for _, errMsg := range tt.wantErrs {
					assert.ErrorContains(t, err, errMsg)
				}
			}
		})
	}
}

func Test_multiErrGroup_MultipleWaits(t *testing.T) {
	var meg multiErrGroup

	meg.Go(func() error {
		return errors.New("test error")
	})

	err1 := meg.Wait()
	require.Error(t, err1)

	err2 := meg.Wait()
	require.Error(t, err2)
	assert.Equal(t, err1.Error(), err2.Error())
}
