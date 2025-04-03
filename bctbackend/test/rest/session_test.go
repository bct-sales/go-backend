//go:build test

package rest

import (
	"bctbackend/rest/path"
	"bctbackend/test/setup"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionExpiration(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := setup.AddAdminToDatabase(db)
	sessionId := setup.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := setup.CreateGetRequest(url)
	request.AddCookie(setup.CreateCookie(sessionId))

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "session_not_found")
}

func TestMissingSessionId(t *testing.T) {
	db, router := setup.CreateRestRouter()
	writer := httptest.NewRecorder()
	defer db.Close()

	admin := setup.AddAdminToDatabase(db)
	setup.AddSessionToDatabaseWithExpiration(db, admin.UserId, -1)

	url := path.Items().String()
	request := setup.CreateGetRequest(url)

	router.ServeHTTP(writer, request)
	RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
}
