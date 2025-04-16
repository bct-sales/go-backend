//go:build test

package rest

import (
	"bctbackend/rest"
	"bctbackend/rest/path"
	. "bctbackend/test"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogout(t *testing.T) {
	setup, router, writer := SetupRestTest()
	defer setup.Close()

	_, sessionId := setup.LoggedIn(setup.Admin())

	url := path.Logout().String()
	request := CreatePostRequest(url, &rest.LogoutPayload{}, WithCookie(sessionId))
	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)
}
