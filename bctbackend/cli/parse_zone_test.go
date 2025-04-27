package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseZones(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testcases := []struct {
			strings []string
			zones   []int
		}{
			{[]string{"1"}, []int{1}},
			{[]string{"1", "2"}, []int{1, 2}},
			{[]string{"1-3"}, []int{1, 2, 3}},
			{[]string{"1-5"}, []int{1, 2, 3, 4, 5}},
			{[]string{"1-3,5-8"}, []int{1, 2, 3, 5, 6, 7, 8}},
			{[]string{"1-3,1-3"}, []int{1, 2, 3}},
			{[]string{"5,2"}, []int{2, 5}},
			{[]string{"5, 2"}, []int{2, 5}},
		}

		for _, testcase := range testcases {
			testLabel := "Parsing " + fmt.Sprintf("%v", testcase.strings)
			t.Run(testLabel, func(t *testing.T) {
				zones, err := parseZones(testcase.strings)
				require.NoError(t, err)
				require.Equal(t, testcase.zones, zones)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		testcases := []struct {
			strings []string
		}{
			{[]string{""}},
			{[]string{"1-1"}},
			{[]string{"3-1"}},
			{[]string{"1-2-3"}},
			{[]string{","}},
			{[]string{"a"}},
			{[]string{"a,1"}},
			{[]string{"a-2"}},
		}

		for _, testcase := range testcases {
			testLabel := "Parsing " + fmt.Sprintf("%v", testcase.strings)
			t.Run(testLabel, func(t *testing.T) {
				_, err := parseZones(testcase.strings)
				require.Error(t, err)
			})
		}
	})
}
