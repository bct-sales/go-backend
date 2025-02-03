package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func AddSale(databasePath string, cashierId models.Id, items []models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	saleId, err := queries.AddSale(db, cashierId, timestamp, items)

	if err != nil {
		return fmt.Errorf("failed to add sale: %v", err)
	}

	fmt.Println("Sale added successfully")

	err = formatting.PrintSale(db, saleId)

	if err != nil {
		fmt.Printf("An error occurred while trying to format the output: %v\n", err)
		fmt.Printf("Sale is still added to the database.\n")
	}

	return nil
}
