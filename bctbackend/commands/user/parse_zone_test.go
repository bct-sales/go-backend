package user

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseZones(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testcases := []struct {
			input string
			zones []int
		}{
			{"1", []int{1}},
			{"1,2", []int{1, 2}},
			{"1-3", []int{1, 2, 3}},
			{"1-5", []int{1, 2, 3, 4, 5}},
			{"1-3,5-8", []int{1, 2, 3, 5, 6, 7, 8}},
			{"1-3,1-3", []int{1, 2, 3}},
			{"5,2", []int{2, 5}},
			{"5, 2", []int{2, 5}},
		}

		for _, testcase := range testcases {
			testLabel := "Parsing " + fmt.Sprintf("%v", testcase.input)
			t.Run(testLabel, func(t *testing.T) {
				t.Parallel()
				zones, err := parseZones(testcase.input)
				require.NoError(t, err)
				require.Equal(t, testcase.zones, zones)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		testcases := []struct {
			input string
		}{
			{""},
			{"1-1"},
			{"3-1"},
			{"1-2-3"},
			{"1,2-4-5,7"},
			{"1,2-,3"},
			{","},
			{"a"},
			{"a,1"},
			{"a-2"},
		}

		for _, testcase := range testcases {
			testLabel := "Parsing " + fmt.Sprintf("%v", testcase.input)
			t.Run(testLabel, func(t *testing.T) {
				t.Parallel()
				_, err := parseZones(testcase.input)
				require.Error(t, err)
			})
		}
	})
}
