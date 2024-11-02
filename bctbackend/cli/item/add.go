package item

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"time"

	"database/sql"

	_ "modernc.org/sqlite"
)

func AddItem(
	databasePath string,
	description string,
	priceInCents models.MoneyInCents,
	categoryId models.Id,
	sellerId models.Id,
	donation bool,
	charity bool) error {

	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("error while opening database: %v", err)
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	if _, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity); err != nil {
		return err
	}

	return nil
}
