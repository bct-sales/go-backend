//go:build test

package helpers

import (
	database "bctbackend/database"
	"database/sql"
	"log"
)

func OpenDatabase() *sql.DB {
	// In-memory database needs to be opened with a shared cache,
	// otherwise each connection will see a different database.
	// Different connections are automatically used
	// when multiple tests access the same database in parallel (i.e., using goroutines).
	db, error := sql.Open("sqlite", "file:memdb?mode=memory&cache=shared")
	if error != nil {
		panic(error)
	}

	// This can also be used to prevent having multiple in-memory databases.
	// It also prevents deadlocks, so it's better not to have them for tests.
	db.SetMaxOpenConns(1)

	db.Exec("PRAGMA foreign_keys = 1")

	return db
}

func OpenInitializedDatabase() *sql.DB {
	db := OpenDatabase()

	if err := database.InitializeDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	return db
}
