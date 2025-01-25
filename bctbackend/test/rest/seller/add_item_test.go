//go:build test

package rest

import (
	. "bctbackend/test/setup"
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

	"github.com/stretchr/testify/require"
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

									seller := AddSellerToDatabase(db, WithUserId(sellerId))
									sessionId := test.AddSessionToDatabase(db, seller.UserId)

									url := path.SellerItems().WithSellerId(seller.UserId)
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

									require.Equal(t, http.StatusCreated, writer.Code)

									response := test.FromJson[restapi.AddSellerItemResponse](writer.Body.String())

									itemsInDatabase, err := queries.GetItems(db)

									require.NoError(t, err)
									require.Equal(t, 1, len(itemsInDatabase))

									itemInDatabase := itemsInDatabase[0]

									require.Equal(t, response.ItemId, itemInDatabase.ItemId)
									require.Equal(t, seller.UserId, itemInDatabase.SellerId)
									require.Equal(t, price, itemInDatabase.PriceInCents)
									require.Equal(t, description, itemInDatabase.Description)
									require.Equal(t, categoryId, itemInDatabase.CategoryId)
									require.Equal(t, donation, itemInDatabase.Donation)
									require.Equal(t, charity, itemInDatabase.Charity)
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

			seller := AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().WithSellerId(seller.UserId)
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
			require.Equal(t, http.StatusBadRequest, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
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

			seller := AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().WithSellerId(seller.UserId)
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
			require.Equal(t, http.StatusBadRequest, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid category", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().WithSellerId(seller.UserId)
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
			require.Equal(t, http.StatusBadRequest, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			admin := AddAdminToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, admin.UserId)

			url := path.SellerItems().WithSellerId(seller.UserId)
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
			require.Equal(t, http.StatusForbidden, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			cashier := AddCashierToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, cashier.UserId)
			url := path.SellerItems().WithSellerId(seller.UserId)
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
			require.Equal(t, http.StatusForbidden, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid url", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			url := path.SellerItems().WithRawSellerId("a")
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

			require.Equal(t, http.StatusBadRequest, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)

			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller1 := AddSellerToDatabase(db)
			seller2 := AddSellerToDatabase(db)
			sessionId := test.AddSessionToDatabase(db, seller2.UserId)

			url := path.SellerItems().WithSellerId(seller1.UserId)
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

			require.Equal(t, http.StatusForbidden, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)

			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as nonexistent seller", func(t *testing.T) {
			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.NewId(1000)
			donation := false
			charity := false

			require.NotContains(t, defs.ListCategories(), categoryId)

			db, router := test.CreateRestRouter()
			writer := httptest.NewRecorder()
			defer db.Close()

			seller := AddSellerToDatabase(db)
			nonexistentId := models.NewId(1000)
			sessionId := test.AddSessionToDatabase(db, seller.UserId)

			userExists, err := queries.UserWithIdExists(db, nonexistentId)
			require.NoError(t, err)
			require.False(t, userExists)

			url := path.SellerItems().WithSellerId(nonexistentId)
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

			require.Equal(t, http.StatusForbidden, writer.Code)

			itemsInDatabase, err := queries.GetItems(db)

			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})
	})
}
