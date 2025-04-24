//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/algorithms"
	"bctbackend/database/models"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	rest "bctbackend/rest/shared"
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
					require.Equal(t, rest.FromTimestamp(seller.CreatedAt), response.CreatedAt)
					require.Nil(t, response.LastActivity)
					require.Len(t, *response.Items, item_count)
				})
			}
		})

		t.Run("Cashier", func(t *testing.T) {
			t.Run("Zero sales", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				cashier := setup.Cashier()
				_, sessionId := setup.LoggedIn(setup.Admin())

				url := path.Users().WithUserId(cashier.UserId)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

				response := FromJson[restapi.GetCashierInformationSuccessResponse](writer.Body.String())
				require.Equal(t, "cashier", response.Role)
				require.Equal(t, cashier.Password, response.Password)
				require.Equal(t, rest.FromTimestamp(cashier.CreatedAt), response.CreatedAt)
				require.Empty(t, response.Sales)
			})

			t.Run("One sale", func(t *testing.T) {
				setup, router, writer := NewRestFixture()
				defer setup.Close()

				seller := setup.Seller()
				cashier := setup.Cashier()
				_, sessionId := setup.LoggedIn(setup.Admin())

				item := setup.Item(seller.UserId, aux.WithDummyData(1))
				saleId := setup.Sale(cashier.UserId, []models.Id{item.ItemId})

				url := path.Users().WithUserId(cashier.UserId)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

				response := FromJson[restapi.GetCashierInformationSuccessResponse](writer.Body.String())
				require.Equal(t, "cashier", response.Role)
				require.Equal(t, cashier.Password, response.Password)
				require.Equal(t, rest.FromTimestamp(cashier.CreatedAt), response.CreatedAt)
				require.Len(t, *response.Sales, 1)
				require.Equal(t, saleId, (*response.Sales)[0].SaleId)
			})

			t.Run("Multiple sales", func(t *testing.T) {
				for _, saleCount := range []int{2, 5, 10} {
					testLabel := fmt.Sprintf("Sale count: %d", saleCount)

					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture()
						defer setup.Close()

						seller := setup.Seller()
						cashier := setup.Cashier()
						_, sessionId := setup.LoggedIn(setup.Admin())

						algorithms.Repeat(saleCount, func() error {
							item := setup.Item(seller.UserId, aux.WithDummyData(1))
							setup.Sale(cashier.UserId, []models.Id{item.ItemId})
							return nil
						})

						url := path.Users().WithUserId(cashier.UserId)
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

						response := FromJson[restapi.GetCashierInformationSuccessResponse](writer.Body.String())
						require.Equal(t, "cashier", response.Role)
						require.Equal(t, cashier.Password, response.Password)
						require.Equal(t, rest.FromTimestamp(cashier.CreatedAt), response.CreatedAt)
						require.NotNil(t, response.Sales)
						require.Len(t, *response.Sales, saleCount)
					})
				}
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.Users().WithUserId(cashier.UserId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.Users().WithUserId(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie("xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.Users().WithUserId(cashier.UserId)
			request := CreateGetRequest(url, WithCookie("whatever", "xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.Users().WithRawUserId("invalid")
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")
		})

		t.Run("Nonexistent user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture()
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			nonexistentUserId := models.Id(99999999)
			setup.RequireNoSuchUser(t, nonexistentUserId)

			url := path.Users().WithUserId(nonexistentUserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})
	})
}
