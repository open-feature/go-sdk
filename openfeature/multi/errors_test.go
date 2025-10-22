package multi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AggregateError_Error(t *testing.T) {
	t.Run("empty error", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{})
		assert.Empty(t, err.Error())
	})

	t.Run("single error", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{
			{
				Err:          fmt.Errorf("test error"),
				ProviderName: "test-provider",
			},
		})

		assert.Equal(t, "Provider test-provider: test error", err.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := NewAggregateError([]ProviderError{
			{
				Err:          fmt.Errorf("test error"),
				ProviderName: "test-provider1",
			},
			{
				Err:          fmt.Errorf("test error"),
				ProviderName: "test-provider2",
			},
		})

		assert.Equal(t, "Provider test-provider1: test error\nProvider test-provider2: test error", err.Error())
	})
}
