//go:build test

package rest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireFailureType(t *testing.T, writer *httptest.ResponseRecorder, expectedStatusCode int, expectedFailureType string) {
	require.Equal(t, expectedStatusCode, writer.Code)

	var response map[string]string
	err := json.Unmarshal(writer.Body.Bytes(), &response)
	require.NoError(t, err)

	failureType, ok := response["type"]
	require.True(t, ok, "failure type not found in response: %s", writer.Body.String())
	require.Equal(t, expectedFailureType, failureType, "unexpected failure type: %s", writer.Body.String())
}
