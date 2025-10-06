//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	path "bctbackend/server/paths"
	"bctbackend/server/rest"
	util "bctbackend/server/shared"
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

			actual := FromJson[rest.ListSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.ListSoldItemsSuccessResponse{
				SoldItems: []rest.ListSoldItemsEntry{},
			}
			require.Equal(t, expected, actual)
		})

		t.Run("Single sale with single item", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			soldItem := items[0]
			sale := setup.Sale(cashier.UserId, []models.ID{soldItem.ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.ListSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.ListSoldItemsSuccessResponse{
				SoldItems: []rest.ListSoldItemsEntry{
					{
						SaleId:          sale.SaleID,
						CashierId:       sale.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale.TransactionTime),
						ItemId:          soldItem.ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(soldItem.AddedAt),
						Description:     soldItem.Description,
						PriceInCents:    soldItem.PriceInCents,
						ItemCategoryID:  soldItem.CategoryID,
						SellerId:        soldItem.SellerID,
						Donation:        soldItem.Donation,
						Charity:         soldItem.Charity,
					},
				},
			}
			require.Equal(t, expected, actual)
		})

		t.Run("Two sales with shared item", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))

			sale1 := setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})
			sale2 := setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[2].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.ListSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.ListSoldItemsSuccessResponse{
				SoldItems: []rest.ListSoldItemsEntry{
					{
						SaleId:          sale1.SaleID,
						CashierId:       sale1.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale1.TransactionTime),
						ItemId:          items[0].ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(items[0].AddedAt),
						Description:     items[0].Description,
						PriceInCents:    items[0].PriceInCents,
						ItemCategoryID:  items[0].CategoryID,
						SellerId:        items[0].SellerID,
						Donation:        items[0].Donation,
						Charity:         items[0].Charity,
					},
					{
						SaleId:          sale1.SaleID,
						CashierId:       sale1.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale1.TransactionTime),
						ItemId:          items[1].ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(items[1].AddedAt),
						Description:     items[1].Description,
						PriceInCents:    items[1].PriceInCents,
						ItemCategoryID:  items[1].CategoryID,
						SellerId:        items[1].SellerID,
						Donation:        items[1].Donation,
						Charity:         items[1].Charity,
					},
					{
						SaleId:          sale2.SaleID,
						CashierId:       sale2.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale2.TransactionTime),
						ItemId:          items[0].ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(items[0].AddedAt),
						Description:     items[0].Description,
						PriceInCents:    items[0].PriceInCents,
						ItemCategoryID:  items[0].CategoryID,
						SellerId:        items[0].SellerID,
						Donation:        items[0].Donation,
						Charity:         items[0].Charity,
					},
					{
						SaleId:          sale2.SaleID,
						CashierId:       sale2.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale2.TransactionTime),
						ItemId:          items[2].ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(items[2].AddedAt),
						Description:     items[2].Description,
						PriceInCents:    items[2].PriceInCents,
						ItemCategoryID:  items[2].CategoryID,
						SellerId:        items[2].SellerID,
						Donation:        items[2].Donation,
						Charity:         items[2].Charity,
					},
				},
			}
			require.Equal(t, expected, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("As seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller, sessionId := setup.LoggedIn(setup.Seller())
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("As cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier, sessionId := setup.LoggedIn(setup.Cashier())
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
		})

		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Cookie with fake session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie("fake_session_id"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[1].ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithCookie("whatever", "whatever"))
			router.ServeHTTP(writer, request)

			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})
	})
}
