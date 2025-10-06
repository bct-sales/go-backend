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

			actual := FromJson[rest.GetSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.GetSoldItemsSuccessResponse{
				SoldItems: []rest.GetSoldItemsEntry{},
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
			sale := setup.Sale(cashier.UserId, []models.Id{soldItem.ItemID})

			url := path.SoldItems()
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusOK, writer.Code)

			actual := FromJson[rest.GetSoldItemsSuccessResponse](t, writer.Body.String())
			expected := &rest.GetSoldItemsSuccessResponse{
				SoldItems: []rest.GetSoldItemsEntry{
					{
						SaleId:          sale.SaleID,
						CashierId:       sale.CashierID,
						TransactionTime: util.ConvertTimestampToDateTime(sale.TransactionTime),
						ItemId:          soldItem.ItemID,
						AddedAt:         util.ConvertTimestampToDateTime(soldItem.AddedAt),
						Description:     soldItem.Description,
						PriceInCents:    soldItem.PriceInCents,
						ItemCategory:    soldItem.CategoryID,
						SellerId:        soldItem.SellerID,
						Donation:        soldItem.Donation,
						Charity:         soldItem.Charity,
					},
				},
			}
			require.Equal(t, expected, actual)
		})
	})
}
