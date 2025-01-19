//go:build test

package rest

import (
	"bctbackend/rest"
	"bctbackend/rest/path"
	"bctbackend/test"
	. "bctbackend/test/setup"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogout(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := AddAdminToDatabase(db)
	sessionId := test.AddSessionToDatabase(db, admin.UserId)

	url := path.Logout().String()
	request := test.CreatePostRequest(url, &rest.LogoutPayload{})
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	require.Equal(t, http.StatusOK, writer.Code)
}
