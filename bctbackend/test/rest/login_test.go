//go:build test

package rest

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/rest/path"
	"bctbackend/security"
	. "bctbackend/test/setup"
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Seller", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			seller := setup.Seller()

			form := url.Values{}
			form.Add("username", models.IdToString(seller.UserId))
			form.Add("password", seller.Password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			var response map[string]string
			require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
			require.Equal(t, "seller", response["role"])

			cookies := writer.Result().Cookies()
			require.NotEmpty(t, cookies, "Expected cookies to be set")

			found := false
			sessionId := ""
			for _, cookie := range cookies {
				if cookie.Name == security.SessionCookieName {
					sessionId = cookie.Value
					found = true
					break
				}
			}
			require.True(t, found, "Expected session_id cookie to be set")

			sessionData, err := queries.GetSessionById(setup.Db, sessionId)
			require.NoError(t, err)
			require.Equal(t, seller.UserId, sessionData.UserId)
		})

		t.Run("Admin", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			admin := setup.Admin()

			form := url.Values{}
			form.Add("username", models.IdToString(admin.UserId))
			form.Add("password", admin.Password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			var response map[string]string
			require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
			require.Equal(t, "admin", response["role"])

			cookies := writer.Result().Cookies()
			require.NotEmpty(t, cookies, "Expected cookies to be set")

			found := false
			sessionId := ""
			for _, cookie := range cookies {
				if cookie.Name == security.SessionCookieName {
					sessionId = cookie.Value
					found = true
					break
				}
			}
			require.True(t, found, "Expected session_id cookie to be set")

			sessionData, err := queries.GetSessionById(setup.Db, sessionId)
			require.NoError(t, err)
			require.Equal(t, admin.UserId, sessionData.UserId)
		})

		t.Run("Cashier", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			cashier := setup.Cashier()

			form := url.Values{}
			form.Add("username", models.IdToString(cashier.UserId))
			form.Add("password", cashier.Password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			var response map[string]string
			require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
			require.Equal(t, "cashier", response["role"])

			cookies := writer.Result().Cookies()
			require.NotEmpty(t, cookies, "Expected cookies to be set")

			found := false
			sessionId := ""
			for _, cookie := range cookies {
				if cookie.Name == security.SessionCookieName {
					sessionId = cookie.Value
					found = true
					break
				}
			}
			require.True(t, found, "Expected session_id cookie to be set")

			sessionData, err := queries.GetSessionById(setup.Db, sessionId)
			require.NoError(t, err)
			require.Equal(t, cashier.UserId, sessionData.UserId)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown login", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			userId := models.Id(0)
			password := "xyz"

			form := url.Values{}
			form.Add("username", models.IdToString(userId))
			form.Add("password", password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})

		t.Run("Wrong password", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			seller := setup.Seller()
			userId := seller.UserId
			password := "wrong password"

			require.NotEqual(t, password, seller.Password, "Bug in tests if this assertion fails")

			form := url.Values{}
			form.Add("username", models.IdToString(userId))
			form.Add("password", password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "wrong_password")
		})

		t.Run("Missing username", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			cashier := setup.Cashier()

			form := url.Values{}
			form.Add("password", cashier.Password)

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_request")
		})

		t.Run("Missing password", func(t *testing.T) {
			setup, router, writer := SetupRestTest()
			defer setup.Close()

			cashier := setup.Cashier()

			form := url.Values{}
			form.Add("username", models.IdToString(cashier.UserId))

			url := path.Login().String()
			request, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_request")
		})
	})
}
