//go:build test

package rest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertFailureType(t *testing.T, writer *httptest.ResponseRecorder, expectedStatusCode int, expectedFailureType string) bool {
	if !assert.Equal(t, expectedStatusCode, writer.Code) {
		return false
	}

	var response map[string]string

	if !assert.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response)) {
		return false
	}

	failureType, ok := response["type"]

	if !assert.True(t, ok) {
		return false
	}

	return assert.Equal(t, expectedFailureType, failureType)
}
