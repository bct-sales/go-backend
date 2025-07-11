//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/algorithms"
	"bctbackend/database/models"
	path "bctbackend/server/paths"
	restapi "bctbackend/server/rest"
	rest "bctbackend/server/shared"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestGetUserInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Logged in as admin", func(t *testing.T) {
			t.Run("Information about admin", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				admin, sessionId := setup.LoggedIn(setup.Admin())

				url := path.User(admin.UserId)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJson[restapi.GetAdminInformationSuccessResponse](t, writer.Body.String())
				require.Equal(t, "admin", response.Role)
				require.Equal(t, admin.Password, response.Password)
				require.Equal(t, rest.ConvertTimestampToDateTime(admin.CreatedAt), response.CreatedAt)
				require.NotNil(t, response.LastActivity)
			})

			t.Run("Information about seller", func(t *testing.T) {
				for _, item_count := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("Item count: %d", item_count)

					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						_, sessionId := setup.LoggedIn(setup.Admin())

						setup.Items(seller.UserId, item_count, aux.WithHidden(false))
						url := path.User(seller.UserId)
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

						response := FromJson[restapi.GetSellerInformationSuccessResponse](t, writer.Body.String())
						require.Equal(t, "seller", response.Role)
						require.Equal(t, seller.Password, response.Password)
						require.Equal(t, rest.ConvertTimestampToDateTime(seller.CreatedAt), response.CreatedAt)
						require.Nil(t, response.LastActivity)
						require.Len(t, *response.Items, item_count)
					})
				}
			})

			t.Run("Information about cashier", func(t *testing.T) {
				t.Run("Zero sales", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					cashier := setup.Cashier()
					_, sessionId := setup.LoggedIn(setup.Admin())

					url := path.User(cashier.UserId)
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

					response := FromJson[restapi.GetCashierInformationSuccessResponse](t, writer.Body.String())
					require.Equal(t, "cashier", response.Role)
					require.Equal(t, cashier.Password, response.Password)
					require.Equal(t, rest.ConvertTimestampToDateTime(cashier.CreatedAt), response.CreatedAt)
					require.Empty(t, response.Sales)
				})

				t.Run("One sale", func(t *testing.T) {
					setup, router, writer := NewRestFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					cashier := setup.Cashier()
					_, sessionId := setup.LoggedIn(setup.Admin())

					item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
					sale := setup.Sale(cashier.UserId, []models.Id{item.ItemID})

					url := path.User(cashier.UserId)
					request := CreateGetRequest(url, WithSessionCookie(sessionId))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

					response := FromJson[restapi.GetCashierInformationSuccessResponse](t, writer.Body.String())
					require.Equal(t, "cashier", response.Role)
					require.Equal(t, cashier.Password, response.Password)
					require.Equal(t, rest.ConvertTimestampToDateTime(cashier.CreatedAt), response.CreatedAt)
					require.Len(t, *response.Sales, 1)
					require.Equal(t, sale.SaleID, (*response.Sales)[0].SaleId)
				})

				t.Run("Multiple sales", func(t *testing.T) {
					for _, saleCount := range []int{2, 5, 10} {
						testLabel := fmt.Sprintf("Sale count: %d", saleCount)

						t.Run(testLabel, func(t *testing.T) {
							setup, router, writer := NewRestFixture(WithDefaultCategories)
							defer setup.Close()

							seller := setup.Seller()
							cashier := setup.Cashier()
							_, sessionId := setup.LoggedIn(setup.Admin())

							algorithms.RepeatWithError(saleCount, func() error {
								item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
								setup.Sale(cashier.UserId, []models.Id{item.ItemID})
								return nil
							})

							url := path.User(cashier.UserId)
							request := CreateGetRequest(url, WithSessionCookie(sessionId))
							router.ServeHTTP(writer, request)
							require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

							response := FromJson[restapi.GetCashierInformationSuccessResponse](t, writer.Body.String())
							require.Equal(t, "cashier", response.Role)
							require.Equal(t, cashier.Password, response.Password)
							require.Equal(t, rest.ConvertTimestampToDateTime(cashier.CreatedAt), response.CreatedAt)
							require.NotNil(t, response.Sales)
							require.Len(t, *response.Sales, saleCount)
						})
					}
				})
			})
		})

		t.Run("Logged in as seller", func(t *testing.T) {
			for _, unfrozenItemCount := range []int{0, 1, 2, 5, 10} {
				for _, frozenItemCount := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("Unfrozen item count: %d, Frozen item count: %d", unfrozenItemCount, frozenItemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(WithDefaultCategories)
						defer setup.Close()

						seller, sessionId := setup.LoggedIn(setup.Seller())
						expectedTotal := models.MoneyInCents(0)

						for i := 0; i != unfrozenItemCount; i++ {
							price := models.MoneyInCents((i + 1) * 50)
							setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithPriceInCents(price), aux.WithHidden(false))
							expectedTotal += price
						}

						for i := 0; i != frozenItemCount; i++ {
							price := models.MoneyInCents((i + 1) * 50)
							setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(true), aux.WithPriceInCents(price), aux.WithHidden(false))
							expectedTotal += price
						}

						url := path.User(seller.UserId)
						request := CreateGetRequest(url, WithSessionCookie(sessionId))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

						response := FromJson[restapi.GetSellerSummarySuccessResponse](t, writer.Body.String())
						require.Equal(t, unfrozenItemCount+frozenItemCount, response.ItemCount)
						require.Equal(t, frozenItemCount, response.FrozenItemCount)
					})
				}
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.User(cashier.UserId)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.User(cashier.UserId)
			request := CreateGetRequest(url, WithSessionCookie("xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			url := path.User(cashier.UserId)
			request := CreateGetRequest(url, WithCookie("whatever", "xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())

			url := path.UserStr("invalid")
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")
		})

		t.Run("Nonexistent user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(WithDefaultCategories)
			defer setup.Close()

			_, sessionId := setup.LoggedIn(setup.Admin())
			nonexistentUserId := models.Id(99999999)
			setup.RequireNoSuchUsers(t, nonexistentUserId)

			url := path.User(nonexistentUserId)
			request := CreateGetRequest(url, WithSessionCookie(sessionId))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})

		t.Run("Unauthorized access", func(t *testing.T) {
			t.Run("Logged in as seller", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Seller())
				otherSeller := setup.Seller()

				url := path.User(otherSeller.UserId)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Logged in as cashier", func(t *testing.T) {
				setup, router, writer := NewRestFixture(WithDefaultCategories)
				defer setup.Close()

				_, sessionId := setup.LoggedIn(setup.Cashier())
				seller := setup.Seller()

				url := path.User(seller.UserId)
				request := CreateGetRequest(url, WithSessionCookie(sessionId))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})
		})
	})
}
