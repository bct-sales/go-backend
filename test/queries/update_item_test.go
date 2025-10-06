//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		oldAddedAt := models.Timestamp(1000)
		oldDescription := "description"
		oldPriceInCents := models.MoneyInCents(1000)
		oldCharity := false
		oldDonation := false
		oldCategory := aux.CategoryId_BabyChildEquipment

		newAddedAt := models.Timestamp(2000)
		newDescription := "new description"
		newPriceInCents := models.MoneyInCents(2000)
		newCharity := true
		newDonation := true
		newCategory := aux.CategoryId_Clothing140_152

		for _, updateAddedAt := range []bool{false, true} {
			for _, updateDescription := range []bool{false, true} {
				for _, updatePriceInCents := range []bool{false, true} {
					for _, updateCharity := range []bool{false, true} {
						for _, updateDonation := range []bool{false, true} {
							for _, updateCategory := range []bool{false, true} {
								testLabel := fmt.Sprintf("updateAddedAt=%v, updateDescription=%v, updatePriceInCents=%v, updateCharity=%v, updateDonation=%v, updateCategory=%v",
									updateAddedAt,
									updateDescription,
									updatePriceInCents,
									updateCharity,
									updateDonation,
									updateCategory,
								)
								t.Run(testLabel, func(t *testing.T) {
									setup, db := NewDatabaseFixture(WithDefaultCategories)
									defer setup.Close()

									seller := setup.Seller()

									item := setup.Item(
										seller.UserID,
										aux.WithAddedAt(oldAddedAt),
										aux.WithDescription(oldDescription),
										aux.WithPriceInCents(oldPriceInCents),
										aux.WithDonation(oldDonation),
										aux.WithCharity(oldCharity),
										aux.WithItemCategory(oldCategory),
										aux.WithFrozen(false),
										aux.WithHidden(false),
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

									setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
										err := queries.UpdateItem(
											transaction,
											item.ItemID,
											&itemUpdate,
										)

										require.NoError(t, err)
									})

									updatedItem, err := queries.GetItemWithID(db, item.ItemID)
									require.NoError(t, err)

									require.Equal(t, item.ItemID, updatedItem.ItemID)
									require.Equal(t, seller.UserID, updatedItem.SellerID)
									require.Equal(t, expectedAddedAt, updatedItem.AddedAt)
									require.Equal(t, expectedDescription, updatedItem.Description)
									require.Equal(t, expectedPriceInCents, updatedItem.PriceInCents)
									require.Equal(t, expectedCharity, updatedItem.Charity)
									require.Equal(t, expectedDonation, updatedItem.Donation)
									require.Equal(t, expectedCategory, updatedItem.CategoryID)
									require.Equal(t, false, updatedItem.Frozen)
								})
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			itemId := models.ID(1)
			itemUpdate := queries.ItemUpdate{}

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateItem(transaction, itemId, &itemUpdate)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
			})
		})

		t.Run("Frozen item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			item := setup.Item(
				seller.UserID,
				aux.WithFrozen(true),
				aux.WithHidden(false),
			)

			itemUpdate := queries.ItemUpdate{}

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateItem(db, item.ItemID, &itemUpdate)
				requireDatabaseWrappedError(t, err, dberr.ErrItemFrozen)
			})
		})

		t.Run("Hidden item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			item := setup.Item(
				seller.UserID,
				aux.WithFrozen(false),
				aux.WithHidden(true),
			)

			itemUpdate := queries.ItemUpdate{}

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateItem(db, item.ItemID, &itemUpdate)
				requireDatabaseWrappedError(t, err, dberr.ErrItemHidden)
			})
		})

		t.Run("Invalid price", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			item := setup.Item(
				seller.UserID,
				aux.WithFrozen(false),
				aux.WithHidden(false),
			)

			invalidPrice := models.MoneyInCents(-100)
			itemUpdate := queries.ItemUpdate{
				PriceInCents: &invalidPrice,
			}

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateItem(db, item.ItemID, &itemUpdate)
				requireDatabaseWrappedError(t, err, dberr.ErrInvalidPrice)
			})
		})
	})
}
