package csv

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
)

func OutputUsers(db *sql.DB, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"user_id", "role_id", "last_activity", "password"}
	err := csvWriter.Write(headers)
	if err != nil {
		return fmt.Errorf("fail;ed to write csv headers: %w", err)
	}

	writeRow := func(user *models.User) error {
		idString := user.UserId.String()
		roleString, err := models.NameOfRole(user.RoleId)
		if err != nil {
			panic(fmt.Sprintf("failed to get name of role %d: %v", user.RoleId, err))
		}
		var lastActivityString string

		if user.LastActivity != nil {
			lastActivityString = user.LastActivity.FormattedDateTime()
		} else {
			lastActivityString = "N/A"
		}

		err = csvWriter.Write([]string{
			idString,
			roleString,
			lastActivityString,
			user.Password,
		})
		if err != nil {
			return fmt.Errorf("failed to write row to csv: %w", err)
		}

		return nil
	}

	if err := queries.GetUsers(db, writeRow); err != nil {
		return fmt.Errorf("failed to write users to file: %w", err)
	}

	return nil
}
