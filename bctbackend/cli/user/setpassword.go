package user

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"

	"database/sql"

	_ "modernc.org/sqlite"
)

func SetPassword(databasePath string, userId models.Id, password string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("error while opening database: %v", err)
	}

	defer db.Close()

	return queries.UpdateUserPassword(db, userId, password)
}
