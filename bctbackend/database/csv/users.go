package csv

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"encoding/csv"
	"io"
)

func OutputUsers(db *sql.DB, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	err := csvWriter.Write([]string{"user_id", "role_id", "last_activity", "password"})
	if err != nil {
		return err
	}

	users, err := queries.ListUsers(db)
	if err != nil {
		return err
	}

	for _, user := range users {
		idString := models.IdToString(user.UserId)
		roleString, err := models.NameOfRole(user.RoleId)
		if err != nil {
			return err
		}
		var lastActivityString string

		if user.LastActivity != nil {
			lastActivityString = models.TimestampToString(*user.LastActivity)
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
			return err
		}
	}

	return nil
}
