//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/server/path"
	"bctbackend/server/rest"
	. "bctbackend/test/setup"

	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"

	"github.com/stretchr/testify/require"
)

func TestAddSellerItem(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()

	t.Run("Successful", func(t *testing.T) {
		for _, sellerId := range []models.Id{models.Id(1), models.Id(2), models.Id(100)} {
			for _, price := range []models.MoneyInCents{1, 100, 10000} {
				for _, description := range []string{"Xyz", "Test Description"} {
					for categoryId, _ := range defaultCategoryNameTable {
						for _, donation := range []bool{true, false} {
							for _, charity := range []bool{true, false} {
								t.Run(fmt.Sprintf("sellerId=%d price=%d description=%s categoryId=%d donation=%t charity=%t", sellerId, price, description, categoryId, donation, charity), func(t *testing.T) {
									setup, router, writer := NewRestFixture(WithDefaultCategories)
									defer setup.Close()

									seller, sessionId := setup.LoggedIn(setup.Seller(aux.WithUserId(sellerId)))

									url := path.SellerItems(seller.UserId)
									payload := rest.AddSellerItemPayload{
										Price:       &price,
										Description: &description,
										CategoryId:  categoryId,
										Donation:    &donation,
										Charity:     &charity,
									}
									request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
									router.ServeHTTP(writer, request)

									require.Equal(t, http.StatusCreated, writer.Code)
									response := FromJson[rest.AddSellerItemResponse](t, writer.Body.String())

									itemsInDatabase := []*models.Item{}
									err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
									require.NoError(t, err)
									require.Equal(t, 1, len(itemsInDatabase))

									itemInDatabase := itemsInDatabase[0]
									require.Equal(t, response.ItemId, itemInDatabase.ItemID)
									require.Equal(t, seller.UserId, itemInDatabase.SellerID)
									require.Equal(t, price, itemInDatabase.PriceInCents)
									require.Equal(t, description, itemInDatabase.Description)
									require.Equal(t, categoryId, itemInDatabase.CategoryID)
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
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := aux.CategoryId_Clothing50_56
			donation := false
			charity := false

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_price")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Empty description", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := ""
			categoryId := aux.CategoryId_Shoes
			donation := false
			charity := false

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "invalid_item_description")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid category", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := models.Id(1000)
			donation := false
			charity := false

			require.NotContains(t, defaultCategoryNameTable, categoryId)

			seller, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_category")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as admin", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := aux.CategoryId_BabyChildEquipment
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding seller item as cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := aux.CategoryId_Clothing104_116
			donation := false
			charity := false

			seller := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Cashier())
			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid url", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := aux.CategoryId_BabyChildEquipment
			donation := false
			charity := false

			_, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItemsStr("a")
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding as different seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := aux.CategoryId_BabyChildEquipment
			donation := false
			charity := false

			seller1 := setup.Seller()
			_, sessionId := setup.LoggedIn(setup.Seller())

			url := path.SellerItems(seller1.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusForbidden, "wrong_seller")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Adding item to nonexistent seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(100)
			description := "Test Description"
			categoryId := aux.CategoryId_BabyChildEquipment
			donation := false
			charity := false

			_, sessionId := setup.LoggedIn(setup.Seller())
			nonexistentUserId := models.Id(1000)
			setup.RequireNoSuchUsers(t, nonexistentUserId)

			url := path.SellerItems(nonexistentUserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("No session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := aux.CategoryId_Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})

		t.Run("Invalid session ID in cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			price := models.MoneyInCents(0)
			description := "Test Description"
			categoryId := aux.CategoryId_Clothing50_56
			donation := false
			charity := false

			seller := setup.Seller()
			invalidSessionId := models.SessionId("xxx")

			url := path.SellerItems(seller.UserId)
			payload := rest.AddSellerItemPayload{
				Price:       &price,
				Description: &description,
				CategoryId:  categoryId,
				Donation:    &donation,
				Charity:     &charity,
			}
			request := CreatePostRequest(url, &payload, WithSessionCookie(invalidSessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")

			itemsInDatabase := []*models.Item{}
			err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 0, len(itemsInDatabase))
		})
	})
}
