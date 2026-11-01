package algorithms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCount(t *testing.T) {
	values := []string{"apple", "banana", "apple", "cherry", "banana", "apple"}

	result := Tally(values, IdentityFunction)

	require.Equal(t, map[string]int{
		"apple":  3,
		"banana": 2,
		"cherry": 1,
	}, result)
}

func TestCountEmpty(t *testing.T) {
	result := Tally([]int(nil), IdentityFunction)

	require.Empty(t, result)
}
