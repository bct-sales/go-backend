package queries

import (
	database "bctrest/db"
	models "bctrest/db/models"
	"bctrest/security"
	"database/sql"

	_ "modernc.org/sqlite"
)

func openDatabase() *sql.DB {
	db, error := sql.Open("sqlite", ":memory:")

	if error != nil {
		panic(error)
	}

	db.Exec("PRAGMA foreign_keys = 1")

	return db
}

func openInitializedDatabase() *sql.DB {
	db := openDatabase()

	database.InitializeDatabase(db)

	return db
}

func addSeller(db *sql.DB, id models.Id) {
	password := "test"
	salt := "xxx"
	hash := security.HashPassword(password, salt)

	AddUser(db, id, models.SellerRoleId, 0, hash, salt)
}
