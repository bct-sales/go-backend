//go:build test

package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bctbackend/defs"
	"bctbackend/rest/path"
	restapi "bctbackend/rest/seller"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"

	"github.com/stretchr/testify/assert"
)

func TestAddSellerItem(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.NewId(1), models.NewId(2), models.NewId(100)} {
			for _, price := range []models.MoneyInCents{1, 100, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for _, categoryId := range defs.ListCategories() {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								t.Run(fmt.Sprintf("sellerId=%d price=%d description=%s categoryId=%d donation=%t charity=%t", sellerId, price, description, categoryId, donation, charity), func(t *testing.T) {
									db, router := test.CreateRestRouter()
									writer := httptest.NewRecorder()
									defer db.Close()

									seller := test.AddSellerWithIdToDatabase(db, sellerId)
									sessionId := test.AddSessionToDatabase(db, seller.UserId)

									url := path.SellerItems().Id(seller.UserId)
									payload := restapi.AddSellerItemPayload{
										Price:       price,
										Description: description,
										CategoryId:  categoryId,
										Donation:    &donation,
										Charity:     &charity,
									}
									request := test.CreatePostRequest(url, &payload)
									request.AddCookie(test.CreateCookie(sessionId))
									router.ServeHTTP(writer, request)

									if assert.Equal(t, http.StatusCreated, writer.Code) {
										response := test.FromJson[restapi.AddSellerItemResponse](writer.Body.String())

										itemsInDatabase, err := queries.GetItems(db)
										if assert.NoError(t, err) {
											if assert.Equal(t, 1, len(itemsInDatabase)) {
												itemInDatabase := itemsInDatabase[0]
												assert.Equal(t, response.ItemId, itemInDatabase.ItemId)
												assert.Equal(t, seller.UserId, itemInDatabase.SellerId)
												assert.Equal(t, price, itemInDatabase.PriceInCents)
												assert.Equal(t, description, itemInDatabase.Description)
												assert.Equal(t, categoryId, itemInDatabase.CategoryId)
												assert.Equal(t, donation, itemInDatabase.Donation)
												assert.Equal(t, charity, itemInDatabase.Charity)
											}
										}
									}
								})
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failing", func(t *testing.T) {
		t.Run("Zero price", func(t *testing.T) {
			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := defs.Clothing50_56
			donation := false
			charity := false

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().Id(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusBadRequest, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Empty description", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := ""
			categoryId := defs.Shoes
			donation := false
			charity := false

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().Id(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusBadRequest, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Invalid category", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().Id(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusBadRequest, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			admin := test.AddAdminToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.SellerItems().Id(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusForbidden, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			cashier := test.AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)

			url := path.SellerItems().Id(seller.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusForbidden, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Invalid url", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().Raw("a")
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusBadRequest, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller1 := test.AddSellerToDatabase(db)
			seller2 := test.AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller2.UserId)

			url := path.SellerItems().Id(seller1.UserId)
			payload := restapi.AddSellerItemPayload{
				Price:       price,
				Description: description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := test.CreatePostRequest(url, &payload)
			request.AddCookie(test.CreateCookie(sessionId))
			router.ServeHTTP(writer, request)

			assert.Equal(t, http.StatusForbidden, writer.Code)
			itemsInDatabase, err := queries.GetItems(db)
			if assert.NoError(t, err) {
				assert.Equal(t, 0, len(itemsInDatabase))
			}
		})

		t.Run("Adding as nonexistent seller", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			assert.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := test.AddSellerToDatabase(db)
			nonexistentId := models.NewId(1000)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			if assert.False(t, queries.UserWithIdExists(db, nonexistentId)) {
				url := path.SellerItems().Id(nonexistentId)
				payload := restapi.AddSellerItemPayload{
					Price:       price,
					Description: description,
					CategoryId:  categoryId,
					Donation:    &donation,
					Charity:     &charity,
				}
				request := test.CreatePostRequest(url, &payload)
				request.AddCookie(test.CreateCookie(sessionId))
				router.ServeHTTP(writer, request)

				assert.Equal(t, http.StatusForbidden, writer.Code)
				itemsInDatabase, err := queries.GetItems(db)
				if assert.NoError(t, err) {
					assert.Equal(t, 0, len(itemsInDatabase))
				}
			}
		})
	})
}
