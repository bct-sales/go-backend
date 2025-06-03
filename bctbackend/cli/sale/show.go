package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"errors"
	"fmt"
)

func ShowSale(databasePath string, saleId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	err = formatting.PrintSale(db, saleId)
	if err != nil {
		return fmt.Errorf("failed to print sale: %w", err)
	}

	return nil
}
