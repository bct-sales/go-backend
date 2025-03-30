package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

func AddSale(databasePath string, cashierId models.Id, items []models.Id) (err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	timestamp := time.Now().Unix()

	saleId, err := queries.AddSale(db, cashierId, timestamp, items)

	if err != nil {
		err = fmt.Errorf("failed to add sale: %v", err)
		return err
	}

	fmt.Println("Sale added successfully")

	err = formatting.PrintSale(db, saleId)

	if err != nil {
		slog.Error("An error occurred while trying to format the output; sale is still added to the database.", "error", err)
		err = nil
		return err // Don't return an error here, as the sale is already added to the database.
	}

	err = nil
	return err
}
