//go:build setup

package rest

import (
	"bctbackend/rest"
	"bctbackend/rest/path"
	. "bctbackend/setup/setup"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/setupify/require"
)

func setupLogout(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := AddAdminToDatabase(db)
	sessionId := setup.AddSessionToDatabase(db, admin.UserId)

	url := path.Logout().String()
	request := setup.CreatePostRequest(url, &rest.LogoutPayload{})
	request.AddCookie(setup.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)
}
