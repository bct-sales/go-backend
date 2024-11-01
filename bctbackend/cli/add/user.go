package add

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"time"

	"database/sql"

	_ "modernc.org/sqlite"
)

func AddUser(databasePath string, userId models.Id, role models.Id, password string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("Error while opening database: %v", err)
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	if err := queries.AddUserWithId(db, userId, role, timestamp, password); err != nil {
		return err
	}

	return nil
}
