//go:build test

package helpers

import (
	database "bctbackend/database"
	"database/sql"
	"log"
)

func CreateInMemoryDatabase() *sql.DB {
	db, error := sql.Open("sqlite", "file:memdb?mode=memory&cache=shared")
	if error != nil {
		panic(error)
	}

	// Each connection has its own in memory DB.
	// We want only one DB, so we force a single connection.
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)

	db.Exec("PRAGMA foreign_keys = 1")

	return db
}

func CreateInitializedInMemoryDatabase() *sql.DB {
	db := CreateInMemoryDatabase()

	if err := database.InitializeDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	return db
}
