//go:build test

package rest

import (
	"net/http"
	"testing"

	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestListSoldItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero zales", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			setup.Items(seller.UserId, 5, aux.WithHidden(false))

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.GetSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.GetSoldItemsSuccessResponse{
				SoldItems: []rest.GetSoldItemsEntry{},
			}
			require.Equal(t, expected, actual)
		})
	})
}
