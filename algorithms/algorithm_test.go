//go:build test

package algorithms_test

import (
	"bctbackend/algorithms"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRange(t *testing.T) {
	tests := []struct {
		start int
		end   int
		want  []int
	}{
		{
			start: 0,
			end:   0,
			want:  []int{},
		},
		{
			start: 0,
			end:   1,
			want:  []int{0},
		},
		{
			start: 1,
			end:   1,
			want:  []int{},
		},
		{
			start: 0,
			end:   5,
			want:  []int{0, 1, 2, 3, 4},
		},
		{
			start: 1,
			end:   5,
			want:  []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("range(%d, %d)", tt.start, tt.end)

		t.Run(testName, func(t *testing.T) {
			got := algorithms.Range(tt.start, tt.end)

			require.Equal(t, tt.want, got)
		})
	}
}
