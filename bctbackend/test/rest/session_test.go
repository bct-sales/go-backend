//go:build test

package rest

import (
	"bctbackend/rest/path"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"
	"net/http"
	"testing"
)

func TestSessionExpiration(t *testing.T) {
	setup, router, writer := SetupRestTest()
	defer setup.Close()

	_, sessionId := setup.LoggedIn(setup.Admin(), aux.WithExpiration(-1))

	url := path.Items().String()
	request := CreateGetRequest(url, WithCookie(sessionId))
	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "session_not_found")
}

func TestMissingSessionId(t *testing.T) {
	setup, router, writer := SetupRestTest()
	defer setup.Close()

	setup.LoggedIn(setup.Admin(), aux.WithExpiration(-1))

	url := path.Items().String()
	request := CreateGetRequest(url)

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
}
