//go:build test

package algorithms

import (
	"bctbackend/algorithms"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDuplicates(t *testing.T) {
	t.Run("Integer slices", func(t *testing.T) {
		test := func(input []int, expected []int) {
			result := algorithms.RemoveDuplicates(input)
			require.ElementsMatch(t, expected, result)
		}

		test([]int{}, []int{})
		test([]int{1}, []int{1})
		test([]int{1, 2}, []int{1, 2})
		test([]int{1, 1}, []int{1})
		test([]int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5})
	})

	t.Run("String slices", func(t *testing.T) {
		test := func(input []string, expected []string) {
			result := algorithms.RemoveDuplicates(input)
			require.ElementsMatch(t, expected, result)
		}

		test([]string{}, []string{})
		test([]string{"a"}, []string{"a"})
		test([]string{"a", "a"}, []string{"a"})
	})
}
