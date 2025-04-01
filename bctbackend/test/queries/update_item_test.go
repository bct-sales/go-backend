//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestUpdateItem(t *testing.T) {
	oldAddedAt := models.Timestamp(1000)
	oldDescription := "description"
	oldPriceInCents := models.MoneyInCents(1000)
	oldCharity := false
	oldDonation := false
	oldCategory := defs.BabyChildEquipment

	newAddedAt := models.Timestamp(2000)
	newDescription := "new description"
	newPriceInCents := models.MoneyInCents(2000)
	newCharity := true
	newDonation := true
	newCategory := defs.Clothing140_152

	for _, updateAddedAt := range []bool{false, true} {
		for _, updateDescription := range []bool{false, true} {
			for _, updatePriceInCents := range []bool{false, true} {
				for _, updateCharity := range []bool{false, true} {
					for _, updateDonation := range []bool{false, true} {
						for _, updateCategory := range []bool{false, true} {
							db := OpenInitializedDatabase()
							defer db.Close()

							seller := AddSellerToDatabase(db)

							item := AddItemToDatabase(
								db,
								seller.UserId,
								WithAddedAt(oldAddedAt),
								WithDescription(oldDescription),
								WithPriceInCents(oldPriceInCents),
								WithDonation(oldDonation),
								WithCharity(oldCharity),
								WithItemCategory(oldCategory),
								WithFrozen(false),
							)

							var itemUpdate queries.ItemUpdate
							expectedAddedAt := oldAddedAt
							expectedDescription := oldDescription
							expectedPriceInCents := oldPriceInCents
							expectedCharity := oldCharity
							expectedDonation := oldDonation
							expectedCategory := oldCategory

							if updateAddedAt {
								itemUpdate.AddedAt = &newAddedAt
								expectedAddedAt = newAddedAt
							}

							if updateDescription {
								itemUpdate.Description = &newDescription
								expectedDescription = newDescription
							}

							if updatePriceInCents {
								itemUpdate.PriceInCents = &newPriceInCents
								expectedPriceInCents = newPriceInCents
							}

							if updateCharity {
								itemUpdate.Charity = &newCharity
								expectedCharity = newCharity
							}

							if updateDonation {
								itemUpdate.Donation = &newDonation
								expectedDonation = newDonation
							}

							if updateCategory {
								itemUpdate.CategoryId = &newCategory
								expectedCategory = newCategory
							}

							err := queries.UpdateItem(
								db,
								item.ItemId,
								&itemUpdate,
							)

							require.NoError(t, err)

							updatedItem, err := queries.GetItemWithId(db, item.ItemId)
							require.NoError(t, err)

							require.Equal(t, item.ItemId, updatedItem.ItemId)
							require.Equal(t, seller.UserId, updatedItem.SellerId)
							require.Equal(t, expectedAddedAt, updatedItem.AddedAt)
							require.Equal(t, expectedDescription, updatedItem.Description)
							require.Equal(t, expectedPriceInCents, updatedItem.PriceInCents)
							require.Equal(t, expectedCharity, updatedItem.Charity)
							require.Equal(t, expectedDonation, updatedItem.Donation)
							require.Equal(t, expectedCategory, updatedItem.CategoryId)
							require.Equal(t, false, updatedItem.Frozen)
						}
					}
				}
			}
		}
	}

}
