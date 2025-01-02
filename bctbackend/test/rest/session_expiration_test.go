//go:build test

package rest

import (
	"bctbackend/rest/path"
	"bctbackend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionExpiration(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := test.AddAdminToDatabase(db)
	sessionId := test.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := test.CreateGetRequest(url)
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	assertFailureType(t, writer, http.StatusUnauthorized, "session_not_found")
}
