package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func ShowSale(databasePath string, saleId models.Id) (err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	err = formatting.PrintSale(db, saleId)

	if err != nil {
		return err
	}

	err = nil
	return err
}
