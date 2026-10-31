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
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				admin, sessionID := setup.LoggedIn(t, setup.Admin(t))

				url := path.User(admin.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code)

				response := FromJSON[restapi.GetAdminInformationByAdminSuccessResponse](t, writer.Body.String())
				require.Equal(t, "admin", response.Role)
				require.Equal(t, admin.Password, response.Password)
				require.Equal(t, rest.ConvertTimestampToDateTime(admin.CreatedAt), response.CreatedAt)
				require.NotNil(t, response.LastActivity)
			})

			t.Run("Information about seller", func(t *testing.T) {
				for _, item_count := range []int{0, 1, 2, 5, 10} {
					testLabel := fmt.Sprintf("Item count: %d", item_count)

					t.Run(testLabel, func(t *testing.T) {
						setup, router, writer := NewRestFixture(t, WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller(t)
						_, sessionID := setup.LoggedIn(t, setup.Admin(t))

						setup.Items(t, seller.UserID, item_count, aux.WithHidden(false))
						url := path.User(seller.UserID)
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

						response := FromJSON[restapi.GetSellerInformationByAdminSuccessResponse](t, writer.Body.String())
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
					setup, router, writer := NewRestFixture(t, WithDefaultCategories)
					defer setup.Close()

					cashier := setup.Cashier(t)
					_, sessionID := setup.LoggedIn(t, setup.Admin(t))

					url := path.User(cashier.UserID)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

					response := FromJSON[restapi.GetCashierInformationByAdminSuccessResponse](t, writer.Body.String())
					require.Equal(t, "cashier", response.Role)
					require.Equal(t, cashier.Password, response.Password)
					require.Equal(t, rest.ConvertTimestampToDateTime(cashier.CreatedAt), response.CreatedAt)
					require.Empty(t, response.Sales)
				})

				t.Run("One sale", func(t *testing.T) {
					setup, router, writer := NewRestFixture(t, WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller(t)
					cashier := setup.Cashier(t)
					_, sessionID := setup.LoggedIn(t, setup.Admin(t))

					item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
					sale := setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})

					url := path.User(cashier.UserID)
					request := CreateGetRequest(url, WithSessionCookie(sessionID))
					router.ServeHTTP(writer, request)
					require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

					response := FromJSON[restapi.GetCashierInformationByAdminSuccessResponse](t, writer.Body.String())
					require.Equal(t, "cashier", response.Role)
					require.Equal(t, cashier.Password, response.Password)
					require.Equal(t, rest.ConvertTimestampToDateTime(cashier.CreatedAt), response.CreatedAt)
					require.Len(t, *response.Sales, 1)
					require.Equal(t, sale.SaleID, (*response.Sales)[0].SaleID)
				})

				t.Run("Multiple sales", func(t *testing.T) {
					for _, saleCount := range []int{2, 5, 10} {
						testLabel := fmt.Sprintf("Sale count: %d", saleCount)

						t.Run(testLabel, func(t *testing.T) {
							setup, router, writer := NewRestFixture(t, WithDefaultCategories)
							defer setup.Close()

							seller := setup.Seller(t)
							cashier := setup.Cashier(t)
							_, sessionID := setup.LoggedIn(t, setup.Admin(t))

							algorithms.RepeatWithError(saleCount, func() error {
								item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
								setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})
								return nil
							})

							url := path.User(cashier.UserID)
							request := CreateGetRequest(url, WithSessionCookie(sessionID))
							router.ServeHTTP(writer, request)
							require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

							response := FromJSON[restapi.GetCashierInformationByAdminSuccessResponse](t, writer.Body.String())
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
						setup, router, writer := NewRestFixture(t, WithDefaultCategories)
						defer setup.Close()

						seller, sessionID := setup.LoggedIn(t, setup.Seller(t))
						expectedTotal := models.MoneyInCents(0)

						for i := 0; i != unfrozenItemCount; i++ {
							price := models.MoneyInCents((i + 1) * 50)
							setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithPriceInCents(price), aux.WithHidden(false))
							expectedTotal += price
						}

						for i := 0; i != frozenItemCount; i++ {
							price := models.MoneyInCents((i + 1) * 50)
							setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithFrozen(true), aux.WithPriceInCents(price), aux.WithHidden(false))
							expectedTotal += price
						}

						url := path.User(seller.UserID)
						request := CreateGetRequest(url, WithSessionCookie(sessionID))
						router.ServeHTTP(writer, request)
						require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

						response := FromJSON[restapi.GetSellerInformationBySellerSuccessResponse](t, writer.Body.String())
						require.Equal(t, unfrozenItemCount+frozenItemCount, response.ItemCount)
						require.Equal(t, frozenItemCount, response.FrozenItemCount)
					})
				}
			}
		})

		t.Run("Logged in as cashier", func(t *testing.T) {
			t.Run("Information about self", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)
				cashier, sessionID := setup.LoggedIn(t, setup.Cashier(t))

				items := setup.Items(t, seller.UserID, 5, aux.WithDummyData(1), aux.WithHidden(false))
				sale1TransactionTime := models.Timestamp(1000)
				sale2TransactionTime := models.Timestamp(4000)
				sale1 := setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID, items[1].ItemID}, aux.WithTransactionTime(sale1TransactionTime))
				sale2 := setup.Sale(t, cashier.UserID, []models.ID{items[2].ItemID, items[3].ItemID}, aux.WithTransactionTime(sale2TransactionTime))

				url := path.User(cashier.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusOK, writer.Code, writer.Body.String())

				response := FromJSON[restapi.GetCashierInformationByCashierSuccessResponse](t, writer.Body.String())

				require.Len(t, *response.Sales, 2)

				require.Equal(t, sale1.SaleID, (*response.Sales)[0].SaleID)
				require.Equal(t, rest.ConvertTimestampToDateTime(sale1TransactionTime), (*response.Sales)[0].TransactionTime)

				require.Equal(t, sale2.SaleID, (*response.Sales)[1].SaleID)
				require.Equal(t, rest.ConvertTimestampToDateTime(sale2TransactionTime), (*response.Sales)[1].TransactionTime)
			})
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No cookie", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)

			url := path.User(cashier.UserID)
			request := CreateGetRequest(url)
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)

			url := path.User(cashier.UserID)
			request := CreateGetRequest(url, WithSessionCookie("xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "no_such_session")
		})

		t.Run("Cookie without session id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)

			url := path.User(cashier.UserID)
			request := CreateGetRequest(url, WithCookie("whatever", "xxx"))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusUnauthorized, "missing_session_id")
		})

		t.Run("Invalid user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(t, setup.Admin(t))

			url := path.UserStr("invalid")
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusBadRequest, "invalid_user_id")
		})

		t.Run("Nonexistent user id", func(t *testing.T) {
			setup, router, writer := NewRestFixture(t, WithDefaultCategories)
			defer setup.Close()

			_, sessionID := setup.LoggedIn(t, setup.Admin(t))
			nonexistentUserID := setup.GenerateNonexistentUserID(t)

			url := path.User(nonexistentUserID)
			request := CreateGetRequest(url, WithSessionCookie(sessionID))
			router.ServeHTTP(writer, request)
			RequireFailureType(t, writer, http.StatusNotFound, "no_such_user")
		})

		t.Run("Unauthorized access", func(t *testing.T) {
			t.Run("Seller querying information about other seller", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Seller(t))
				otherSeller := setup.Seller(t)

				url := path.User(otherSeller.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Seller querying information about cashier", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Seller(t))
				cashier := setup.Cashier(t)

				url := path.User(cashier.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Seller querying information about admin", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Seller(t))
				admin := setup.Admin(t)

				url := path.User(admin.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Cashier querying information about seller", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Cashier(t))
				seller := setup.Seller(t)

				url := path.User(seller.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Cashier querying information about other cashier", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Cashier(t))
				otherCashier := setup.Cashier(t)

				url := path.User(otherCashier.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})

			t.Run("Cashier querying information about admin", func(t *testing.T) {
				setup, router, writer := NewRestFixture(t, WithDefaultCategories)
				defer setup.Close()

				_, sessionID := setup.LoggedIn(t, setup.Cashier(t))
				admin := setup.Admin(t)

				url := path.User(admin.UserID)
				request := CreateGetRequest(url, WithSessionCookie(sessionID))
				router.ServeHTTP(writer, request)
				RequireFailureType(t, writer, http.StatusForbidden, "wrong_role")
			})
		})
	})
}
