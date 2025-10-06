//go:build test

package rest

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	path "bctbackend/server/paths"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestHeavyLoad(t *testing.T) {
	t.Run("Updating many items at once", func(t *testing.T) {
		setup, router, _ := NewRestFixture(WithDefaultCategories)
		defer func() {
			setup.Close()
		}()

		seller, sessionID := setup.LoggedIn(setup.Seller())
		itemCount := 100
		items := setup.Items(seller.UserID, itemCount, aux.WithHidden(false), aux.WithFrozen(false))

		waitGroup := sync.WaitGroup{}

		for _, item := range items {
			waitGroup.Add(1)

			go func() {
				defer waitGroup.Done()

				time.Sleep(time.Second)
				url := path.Item(item.ItemID)
				payload := struct {
					PriceInCents int  `json:"priceInCents"`
					Donation     bool `json:"donation"`
					Charity      bool `json:"charity"`
				}{
					PriceInCents: int(item.PriceInCents) * 2,
					Donation:     !item.Donation,
					Charity:      !item.Charity,
				}

				request := CreatePutRequest(url, &payload, WithSessionCookie(sessionID))
				writer := setup.NewResponseRecorder()
				router.ServeHTTP(writer, request)
				require.Equal(t, http.StatusNoContent, writer.Code, "body", writer.Body.String())
			}()
		}

		waitGroup.Wait()

		newItems := []*models.Item{}
		err := queries.NewGetItemsQuery().Execute(setup.Db, queries.CollectTo(&newItems))
		require.NoError(t, err)
		require.Len(t, newItems, itemCount)

		for i, item := range items {
			newItem := newItems[i]

			require.Equal(t, item.PriceInCents*2, newItem.PriceInCents)
			require.Equal(t, !item.Donation, newItem.Donation)
			require.Equal(t, !item.Charity, newItem.Charity)
			require.Equal(t, item.Description, newItem.Description)
		}
	})
}
