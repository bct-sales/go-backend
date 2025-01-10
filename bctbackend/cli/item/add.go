package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"fmt"
	"time"

	"github.com/pterm/pterm"
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

	reportErrorFormattingOutput := func(err error) error {
		fmt.Printf("An error occurred while trying to format the output: %v\n", err)
		fmt.Printf("Item is still added to the database.\n")
		return nil
	}

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	if _, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity); err != nil {
		return err
	}

	fmt.Println("Item added successfully!")

	categoryName, err := defs.NameOfCategory(categoryId)

	if err != nil {
		return reportErrorFormattingOutput(err)
	}

	tableData := pterm.TableData{
		{"Description", description},
		{"Price", fmt.Sprintf("%.2f", float64(priceInCents)/100.0)},
		{"Category", categoryName},
		{"Seller", fmt.Sprintf("%d", sellerId)},
		{"Donation", fmt.Sprintf("%t", donation)},
		{"Charity", fmt.Sprintf("%t", charity)},
	}

	err = pterm.DefaultTable.WithData(tableData).Render()

	if err != nil {
		return reportErrorFormattingOutput(err)
	}

	return nil
}
