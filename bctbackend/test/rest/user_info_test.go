//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/database/models"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetUserInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			admin, sessionId := setup.LoggedIn(setup.Admin())

			url := path.Users().WithUserId(admin.UserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			response := FromJson[restapi.GetAdminInformationSuccessResponse](writer.Body.String())
			require.Equal(t, "admin", response.Role)
			require.Equal(t, admin.Password, response.Password)
			require.Equal(t, admin.CreatedAt, response.CreatedAt)
			require.NotNil(t, response.LastActivity)
		})

		t.Run("Seller", func(t *testing.T) {
			for _, item_count := range []int{0, 1, 2, 5, 10} {
				testLabel := fmt.Sprintf("Item count: %d", item_count)

				t.Run(testLabel, func(t *testing.T) {
					setup, router, writer := NewRestFixture()
					defer setup.Close()

					seller := setup.Seller()
					_, sessionId := setup.LoggedIn(setup.Admin())

					items := make([]*models.Item, item_count)
					for i := 0; i < item_count; i++ {
						items[i] = setup.Item(seller.UserId, aux.WithDummyData(i))
					}

					url := path.Users().WithUserId(seller.UserId)
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

					response := FromJson[restapi.GetSellerInformationSuccessResponse](writer.Body.String())
					require.Equal(t, "seller", response.Role)
					require.Equal(t, seller.Password, response.Password)
					require.Equal(t, seller.CreatedAt, response.CreatedAt)
					require.Nil(t, response.LastActivity)
					require.Len(t, *response.Items, item_count)
				})
			}
		})
	})
}
