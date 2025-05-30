//go:build test

package algorithms

import (
	"bctbackend/algorithms"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepeat(t *testing.T) {
	for _, repeatCount := range []int{0, 1, 2, 3, 10} {
		testLabel := fmt.Sprintf("Repeat %d times", repeatCount)
		t.Run(testLabel, func(t *testing.T) {
			t.Parallel()

			value := 0

			algorithms.Repeat(repeatCount, func() {
				value++
			})

			require.Equal(t, repeatCount, value, "Expected value to be incremented %d times", repeatCount)
		})
	}
}
