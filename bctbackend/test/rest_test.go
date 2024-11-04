package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"
	"bctbackend/database/queries"

	"github.com/stretchr/testify/assert"
)

func TestListItems(t *testing.T) {
	t.Run("No items", func(t *testing.T) {
		db, router := createRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		request, err := http.NewRequest("GET", "/api/v1/items", nil)

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			assert.Equal(t, 200, writer.Code)

			expected := []models.Item{}
			actual := fromJson[[]models.Item](writer.Body.String())
			assert.Equal(t, expected, *actual)
		}
	})

	t.Run("One item", func(t *testing.T) {
		db, router := createRestRouter()
		writer := httptest.NewRecorder()
		defer db.Close()

		sellerId := addTestSeller(db)
		item := models.NewItem(0, 100, "test item", 1000, models.Shoes, sellerId, false, false)
		itemId, err := queries.AddItem(db, item.Timestamp, item.Description, item.PriceInCents, item.CategoryId, item.SellerId, item.Donation, item.Charity)

		if !assert.NoError(t, err) {
			return
		}

		item.ItemId = itemId

		request, err := http.NewRequest("GET", "/api/v1/items", nil)

		if assert.NoError(t, err) {
			router.ServeHTTP(writer, request)

			assert.Equal(t, 200, writer.Code)

			expected := []models.Item{*item}
			actual := fromJson[[]models.Item](writer.Body.String())
			assert.Equal(t, expected, *actual)
		}
	})
}
