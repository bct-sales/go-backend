package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

func AddSale(databasePath string, cashierId models.Id, items []models.Id) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	timestamp := models.Now()

	saleId, err := queries.AddSale(db, cashierId, timestamp, items)

	if err != nil {
		return fmt.Errorf("failed to add sale: %w", err)
	}

	fmt.Println("Sale added successfully")

	err = formatting.PrintSale(db, saleId)

	if err != nil {
		slog.Error("An error occurred while trying to format the output; sale is still added to the database.", "error", err)
		return nil // Don't return an error here, as the sale is already added to the database.
	}

	return nil
}
