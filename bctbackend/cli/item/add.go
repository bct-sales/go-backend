package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	if _, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity); err != nil {
		return err
	}

	categoryName, err := models.NameOfCategory(categoryId)

	if err != nil {
		return fmt.Errorf("error while converting category to string: %v", err)
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
		return fmt.Errorf("error while rendering table: %v", err)
	}

	return nil
}
