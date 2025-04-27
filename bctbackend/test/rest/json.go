package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func ToJson(x any) string {
	jsonData, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func FromJson[T any](t *testing.T, jsonString string) *T {
	var x T
	err := json.Unmarshal([]byte(jsonString), &x)
	require.NoError(t, err, "Failed to unmarshal JSON: %s", jsonString)
	return &x
}
