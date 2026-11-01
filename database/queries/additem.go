package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"fmt"
)

// AddItem adds an item to the database.
// The ID of the newly added item is returned.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
// An ErrNoSuchCategory is returned if the itemCategoryID is invalid.
// An ErrInvalidPrice is returned if the priceInCents is invalid.
// An ErrInvalidItemDescription is returned if the description is invalid.
func AddItem(
	ctx context.Context,
	db DatabaseQuerier,
	addedAt models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryID models.ID,
	sellerID models.ID,
	donation bool,
	charity bool,
	frozen bool,
	hidden bool,
	large bool) (r_result models.ID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Validate inputs
	if !models.IsValidPrice(priceInCents) {
		return 0, fmt.Errorf("failed to add item with price %d: %w", priceInCents, dberr.ErrInvalidPrice)
	}
	if !models.IsValidItemDescription(description) {
		return 0, fmt.Errorf("failed to add item with description %s: %w", description, dberr.ErrInvalidItemDescription)
	}
	// No transaction is necessary here, since users don't change
	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("could not ensure user %d exists and is seller: %w", sellerID, err)
	}
	if frozen && hidden {
		return 0, fmt.Errorf("failed to add item: %w", dberr.ErrHiddenFrozenItem)
	}

	// Insert the item into the database
	itemID, err := isql.New(db).AddItem(ctx, isql.AddItemParams{
		AddedAt:        int64(addedAt),
		Description:    description,
		PriceInCents:   int64(priceInCents),
		ItemCategoryID: itemCategoryID.Int64(),
		SellerID:       sellerID.Int64(),
		Donation:       donation,
		Charity:        charity,
		Frozen:         frozen,
		Hidden:         hidden,
		Large:          large,
	})
	if err != nil {
		categoryExists, err2 := CategoryWithIDExists(ctx, db, itemCategoryID)
		if err2 != nil {
			return 0, fmt.Errorf("failed to determine whether category with given id exists: %w", err)
		}

		if !categoryExists {
			return 0, fmt.Errorf("failed to add item with category %d: %w", itemCategoryID, dberr.ErrNoSuchCategory)
		}

		return 0, fmt.Errorf("failed to insert item: %w", err)
	}

	return itemID, nil
}
