//go:build test

package rest

import (
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, router, writer := NewRestFixture(WithDefaultCategories)
		defer setup.Close()

		_, sessionId := setup.LoggedIn(setup.Admin())

		url := path.Logout()
		request := CreatePostRequest(url, &rest.LogoutPayload{}, WithSessionCookie(sessionId))
		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusOK, writer.Code)
	})

	t.Run("Expired session", func(t *testing.T) {
		setup, router, writer := NewRestFixture(WithDefaultCategories)
		defer setup.Close()

		_, sessionId := setup.LoggedIn(setup.Admin(), aux.WithExpiration(100))

		setup.Clock.Advance(200)

		url := path.Logout()
		request := CreatePostRequest(url, &rest.LogoutPayload{}, WithSessionCookie(sessionId))
		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusUnauthorized, writer.Code, "body", writer.Body.String())
	})
}
