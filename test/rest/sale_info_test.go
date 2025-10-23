//go:build test

package rest

import (
	"net/http"
	"testing"

	"bctbackend/database/models"
	path "bctbackend/server/paths"
	restapi "bctbackend/server/rest"
	rest "bctbackend/server/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetSaleInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("As admin", func(t *testing.T) {
			t.Run("Single item in sale", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				cashier := setup.Cashier()

				transactionTime := models.Timestamp(100)
				item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
				sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID}, aux.WithTransactionTime(transactionTime))

				url := path.Sale(sale.SaleID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJSON[restapi.GetSaleInformationSuccessResponse](t, writer.Body.String())
				require.Equal(t, sale.SaleID, response.SaleID)
				require.Equal(t, cashier.UserID, response.CashierID)
				require.Equal(t, rest.ConvertTimestampToDateTime(transactionTime), response.TransactionTime)
				require.Equal(t, 1, len(response.Items))
				require.Equal(t, item.ItemID, response.Items[0].ItemID)
				require.Equal(t, item.SellerID, response.Items[0].SellerID)
				require.Equal(t, item.CategoryID, response.Items[0].CategoryID)
				require.Equal(t, item.Description, response.Items[0].Description)
				require.Equal(t, item.PriceInCents, response.Items[0].PriceInCents)
				require.Equal(t, item.Charity, *response.Items[0].Charity)
				require.Equal(t, item.Donation, *response.Items[0].Donation)
				require.Equal(t, item.Large, *response.Items[0].Large)
			})

			t.Run("Five item in sale", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(setup.Admin())
				seller := setup.Seller()
				cashier := setup.Cashier()

				transactionTime := models.Timestamp(100)
				itemCount := 5
				items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))
				itemIDs := models.CollectItemIDs(items)
				sale := setup.Sale(cashier.UserID, itemIDs, aux.WithTransactionTime(transactionTime))

				url := path.Sale(sale.SaleID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJSON[restapi.GetSaleInformationSuccessResponse](t, writer.Body.String())
				require.Equal(t, cashier.UserID, response.CashierID)
				require.Equal(t, rest.ConvertTimestampToDateTime(transactionTime), response.TransactionTime)
				require.Equal(t, itemCount, len(response.Items))

				for i, item := range items {
					require.Equal(t, item.ItemID, response.Items[i].ItemID)
					require.Equal(t, item.SellerID, response.Items[i].SellerID)
					require.Equal(t, item.CategoryID, response.Items[i].CategoryID)
					require.Equal(t, item.Description, response.Items[i].Description)
					require.Equal(t, item.PriceInCents, response.Items[i].PriceInCents)
					require.Equal(t, item.Charity, *response.Items[i].Charity)
					require.Equal(t, item.Donation, *response.Items[i].Donation)
					require.Equal(t, item.Large, *response.Items[i].Large)
				}
			})
		})

		t.Run("As owning cashier", func(t *testing.T) {
			t.Run("Five item in sale", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier, sessionID := setup.LoggedIn(setup.Cashier())

				transactionTime := models.Timestamp(100)
				itemCount := 5
				items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false))
				itemIDs := models.CollectItemIDs(items)
				sale := setup.Sale(cashier.UserID, itemIDs, aux.WithTransactionTime(transactionTime))

				url := path.Sale(sale.SaleID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJSON[restapi.GetSaleInformationSuccessResponse](t, writer.Body.String())
				require.Equal(t, cashier.UserID, response.CashierID)
				require.Equal(t, rest.ConvertTimestampToDateTime(transactionTime), response.TransactionTime)
				require.Equal(t, itemCount, len(response.Items))

				for i, item := range items {
					require.Equal(t, item.ItemID, response.Items[i].ItemID)
					require.Equal(t, item.SellerID, response.Items[i].SellerID)
					require.Equal(t, item.CategoryID, response.Items[i].CategoryID)
					require.Equal(t, item.Description, response.Items[i].Description)
					require.Equal(t, item.PriceInCents, response.Items[i].PriceInCents)
					require.Equal(t, item.Charity, *response.Items[i].Charity)
					require.Equal(t, item.Donation, *response.Items[i].Donation)
					require.Equal(t, item.Large, *response.Items[i].Large)
				}
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown sale", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(setup.Admin())
			nonexistentSaleID := setup.GenerateNonexistentSaleID(t)

			url := path.Sale(nonexistentSaleID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusNotFound, writer.Code)
		})

		t.Run("As seller", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller, sessionID := setup.LoggedIn(setup.Seller())
			cashier := setup.Cashier()
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID})

			url := path.Sale(sale.SaleID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})

		t.Run("As other cashier", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			_, sessionID := setup.LoggedIn(setup.Cashier())
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID})

			url := path.Sale(sale.SaleID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			require.Equal(t, http.StatusForbidden, writer.Code)
		})
	})
}
