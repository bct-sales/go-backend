//go:build test

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/rest/path"
	"bctbackend/test"

	"github.com/stretchr/testify/require"
)

func TestListSellerItems(t *testing.T) {
	for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
		for _, itemCount := range []int{0, 1, 5, 100} {
			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db, test.WithUserId(sellerId))
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			expectedItems := []models.Item{}
			for i := 0; i < itemCount; i++ {
				expectedItems = append(expectedItems, *test.AddItemToDatabase(db, seller.UserId, i))
			}

			url := path.SellerItems().WithSellerId(seller.UserId)
			request := test.CreateGetRequest(url)

			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := test.FromJson[[]models.Item](writer.Body.String())
			require.Equal(t, expectedItems, *actual)
		}
	}
}
