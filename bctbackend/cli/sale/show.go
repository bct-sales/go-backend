package user

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"fmt"

	_ "modernc.org/sqlite"
)

func ShowSale(databasePath string, saleId models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer db.Close()

	err = formatting.PrintSale(db, saleId)

	if err != nil {
		return err
	}

	return nil
}
