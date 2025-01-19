//go:build test

package rest

import (
	"bctbackend/rest/path"
	"bctbackend/test"
	. "bctbackend/test/setup"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionExpiration(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := AddAdminToDatabase(db)
	sessionId := test.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := test.CreateGetRequest(url)
	request.AddCookie(test.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "session_not_found")
}

func TestMissingSessionId(t *testing.T) {
	db, router := test.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := AddAdminToDatabase(db)
	test.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := test.CreateGetRequest(url)

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
}
