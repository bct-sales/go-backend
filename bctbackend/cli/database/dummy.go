package database

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

func ResetDatabaseAndFillWithDummyData(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	slog.Info("Adding categories")
	{
		GenerateDefaultCategories(func(id models.Id, name string) error {
			return queries.AddCategory(db, id, name)
		})
	}

	slog.Info("Adding admin user")
	{
		id := models.Id(1)
		role := models.AdminRoleId
		createdAt := models.Now()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		if err = queries.AddUserWithId(db, id, role, createdAt, lastActivity, password); err != nil {
			return err
		}
	}

	slog.Info("Adding cashier user")
	{
		id := models.Id(2)
		role := models.CashierRoleId
		createdAt := models.Now()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		if err = queries.AddUserWithId(db, id, role, createdAt, lastActivity, password); err != nil {
			return err
		}
	}

	slog.Info("Adding sellers")
	addSellers := func(addUser func(userId models.Id, roleId models.Id, createdAt models.Timestamp, lastActivity *models.Timestamp, password string)) {
		for area := 1; area <= 12; area++ {
			for offset := 0; offset != 4; offset++ {
				userId := models.Id(area*100 + offset)
				roleId := models.SellerRoleId
				createdAt := models.Now()
				var lastActivity *models.Timestamp = nil
				password := fmt.Sprintf("%d", userId)

				addUser(userId, roleId, createdAt, lastActivity, password)
			}
		}
	}
	if err := queries.AddUsers(db, addSellers); err != nil {
		return fmt.Errorf("failed to add sellers: %v", err)
	}

	slog.Info("Adding some items")
	queries.AddItem(db, 1, "T-Shirt", 1000, CategoryId_Clothing140_152, 100, false, false, false, false)
	queries.AddItem(db, 1, "Jeans", 1000, CategoryId_Clothing140_152, 100, false, false, false, false)
	queries.AddItem(db, 1, "Nike sneakers", 2000, CategoryId_Shoes, 100, false, false, false, false)
	queries.AddItem(db, 1, "Adidas sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Puma sneakers", 2000, CategoryId_Shoes, 200, false, false, true, false)
	queries.AddItem(db, 1, "Reebok sneakers", 2000, CategoryId_Shoes, 200, false, true, false, false)
	queries.AddItem(db, 1, "Converse sneakers", 2000, CategoryId_Shoes, 200, true, false, false, false)
	queries.AddItem(db, 1, "Vans sneakers", 2000, CategoryId_Shoes, 200, true, true, false, false)
	queries.AddItem(db, 1, "New Balance sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Asics sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Hoka sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Saucony sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Brooks sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Mizuno sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "On sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false)
	queries.AddItem(db, 1, "Combat boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Hiking boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Winter boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Rain boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Snow boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Bean boots", 2000, CategoryId_Shoes, 300, false, false, false, false)
	queries.AddItem(db, 1, "Cowboy boots", 2000, CategoryId_Shoes, 300, false, false, false, false)

	return nil
}
