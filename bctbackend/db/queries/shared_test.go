package queries

import (
	database "bctbackend/db"
	models "bctbackend/db/models"
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

	AddUser(db, id, models.SellerRoleId, 0, password)
}
