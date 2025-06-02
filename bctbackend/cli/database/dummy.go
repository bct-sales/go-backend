package database

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

func ResetDatabaseAndFillWithDummyData(databasePath string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	slog.Info("Resetting database")
	if err = database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %w", err)
	}

	slog.Info("Adding categories")
	{
		if err := GenerateDefaultCategories(func(id models.Id, name string) error { return queries.AddCategoryWithId(db, id, name) }); err != nil {
			return fmt.Errorf("failed to add categories: %w", err)
		}
	}

	slog.Info("Adding admin user")
	{
		id := models.Id(1)
		role := models.AdminRoleId
		createdAt := models.Now()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		if err = queries.AddUserWithId(db, id, role, createdAt, lastActivity, password); err != nil {
			return fmt.Errorf("failed to add admin: %w", err)
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
			return fmt.Errorf("failed to add cashier: %w", err)
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
		return fmt.Errorf("failed to add sellers: %w", err)
	}

	{
		now := models.Now()

		slog.Info("Adding some items")
		err := errors.Join(
			addItem(db, now, "T-Shirt", 1000, CategoryId_Clothing140_152, 100, false, false, false, false),
			addItem(db, now, "Jeans", 1000, CategoryId_Clothing140_152, 100, false, false, false, false),
			addItem(db, now, "Nike sneakers", 2000, CategoryId_Shoes, 100, false, false, false, false),
			addItem(db, now, "Adidas sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Puma sneakers", 2000, CategoryId_Shoes, 200, false, false, true, false),
			addItem(db, now, "Reebok sneakers", 2000, CategoryId_Shoes, 200, false, true, false, false),
			addItem(db, now, "Converse sneakers", 2000, CategoryId_Shoes, 200, true, false, false, false),
			addItem(db, now, "Vans sneakers", 2000, CategoryId_Shoes, 200, true, true, false, false),
			addItem(db, now, "New Balance sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Asics sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Hoka sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Saucony sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Brooks sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Mizuno sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "On sneakers", 2000, CategoryId_Shoes, 200, false, false, false, false),
			addItem(db, now, "Combat boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Hiking boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Winter boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Rain boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Snow boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Bean boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
			addItem(db, now, "Cowboy boots", 2000, CategoryId_Shoes, 300, false, false, false, false),
		)

		if err != nil {
			return fmt.Errorf("failed to add items: %w", err)
		}
	}

	return nil
}

func addItem(db *sql.DB, addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryId models.Id, sellerId models.Id, donation bool, charity bool, frozen bool, hidden bool) error {
	_, err := queries.AddItem(db, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
	return fmt.Errorf("failed to add item: %w", err)
}
