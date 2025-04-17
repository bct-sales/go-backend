//go:build test

package rest

import (
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
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
	RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
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
