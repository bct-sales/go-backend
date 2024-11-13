//go:build test

package rest

import (
	"bctbackend/rest"
	"bctbackend/rest/path"
	"bctbackend/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogout(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := test.AddAdminToDatabase(db)
	sessionId := test.AddSessionToDatabase(db, admin.UserId)

	url := path.Logout().String()
	request := test.CreatePostRequest(url, &rest.LogoutPayload{})
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
}
