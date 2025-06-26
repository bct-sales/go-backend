//go:build test

package rest

import (
	path "bctbackend/server/paths"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"net/http"
	"testing"
)

func TestSessionExpiration(t *testing.T) {
	setup, router, writer := NewRestFixture(WithDefaultCategories)
	defer setup.Close()

	_, sessionId := setup.LoggedIn(setup.Admin(), aux.WithExpiration(-1))

	url := path.Items()
	request := CreateGetRequest(url, WithSessionCookie(sessionId))
	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
}

func TestMissingSessionId(t *testing.T) {
	setup, router, writer := NewRestFixture(WithDefaultCategories)
	defer setup.Close()

	setup.LoggedIn(setup.Admin(), aux.WithExpiration(-1))

	url := path.Items()
	request := CreateGetRequest(url)

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
}
