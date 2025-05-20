//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/defs"
	"database/sql"
	"strconv"

	_ "modernc.org/sqlite"
)

type AddItemData struct {
	AddedAt      *models.Timestamp
	Description  *string
	PriceInCents *models.MoneyInCents
	ItemCategory *models.Id
	Donation     *bool
	Charity      *bool
	Frozen       *bool
	Hidden       *bool
}

func (data *AddItemData) FillWithDefaults() {
	if data.AddedAt == nil {
		addedAt := models.NewTimestamp(0)
		data.AddedAt = &addedAt
	}

	if data.Description == nil {
		description := "description"
		data.Description = &description
	}

	if data.PriceInCents == nil {
		priceInCents := models.NewMoneyInCents(100)
		data.PriceInCents = &priceInCents
	}

	if data.ItemCategory == nil {
		itemCategory := defs.Shoes
		data.ItemCategory = &itemCategory
	}

	if data.Donation == nil {
		donation := false
		data.Donation = &donation
	}

	if data.Charity == nil {
		charity := false
		data.Charity = &charity
	}

	if data.Frozen == nil {
		frozen := false
		data.Frozen = &frozen
	}

	if data.Hidden == nil {
		panic("Hidden is nil")
	}
}

func WithAddedAt(addedAt models.Timestamp) func(*AddItemData) {
	return func(data *AddItemData) {
		data.AddedAt = &addedAt
	}
}

func WithDescription(description string) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Description = &description
	}
}

func WithPriceInCents(priceInCents models.MoneyInCents) func(*AddItemData) {
	return func(data *AddItemData) {
		data.PriceInCents = &priceInCents
	}
}

func WithItemCategory(itemCategory models.Id) func(*AddItemData) {
	return func(data *AddItemData) {
		data.ItemCategory = &itemCategory
	}
}

func WithDonation(donation bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Donation = &donation
	}
}

func WithCharity(charity bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Charity = &charity
	}
}

func WithFrozen(frozen bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Frozen = &frozen
	}
}

func WithHidden(hidden bool) func(*AddItemData) {
	return func(data *AddItemData) {
		data.Hidden = &hidden
	}
}

func WithDummyData(k int) func(*AddItemData) {
	return func(data *AddItemData) {
		addedAt := models.NewTimestamp(0)
		description := "description " + strconv.Itoa(k)
		priceInCents := models.NewMoneyInCents(100 + int64(k))
		itemCategory := defs.Shoes
		donation := k%2 == 0
		charity := k%3 == 0
		frozen := k%2 == 0

		data.AddedAt = &addedAt
		data.Description = &description
		data.PriceInCents = &priceInCents
		if data.ItemCategory == nil {
			data.ItemCategory = &itemCategory
		}
		data.Donation = &donation
		data.Charity = &charity
		data.Frozen = &frozen
	}
}

func AddItemToDatabase(db *sql.DB, sellerId models.Id, options ...func(*AddItemData)) *models.Item {
	data := AddItemData{}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	itemId, err := queries.AddItem(db, *data.AddedAt, *data.Description, *data.PriceInCents, *data.ItemCategory, sellerId, *data.Donation, *data.Charity, *data.Frozen, *data.Hidden)
	if err != nil {
		panic(err)
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		panic(err)
	}

	return item
}
