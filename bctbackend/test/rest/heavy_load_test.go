//go:build test

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	restapi "bctbackend/rest"
	"bctbackend/rest/path"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

func TestHeavyLoad(t *testing.T) {
	setup, router, writer := NewRestFixture()
	defer setup.Close()

	seller, sessionId := setup.LoggedIn(setup.Seller())
	url := path.SellerItems().WithSellerId(seller.UserId)

	itemCount := 1000
	for i := 0; i < itemCount; i++ {
		price := models.MoneyInCents(100 * (i + 1))
		description := fmt.Sprintf("Test item %d", i)
		categoryId := defs.Clothing104_116
		donation := false
		charity := false

		payload := restapi.AddSellerItemPayload{
			Price:       &price,
			Description: &description,
			CategoryId:  categoryId,
			Donation:    &donation,
			Charity:     &charity,
		}

		request := CreatePostRequest(url, &payload, WithSessionCookie(sessionId))
		router.ServeHTTP(writer, request)
		require.Equal(t, http.StatusCreated, writer.Code)
	}

	itemsInDatabase := []*models.Item{}
	err := queries.GetItems(setup.Db, queries.CollectTo(&itemsInDatabase), queries.IncludeHidden)
	require.NoError(t, err)
	require.Equal(t, itemCount, len(itemsInDatabase))
}
