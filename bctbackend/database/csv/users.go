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

	headers := []string{"user_id", "role_id", "last_activity", "password"}
	err := csvWriter.Write(headers)
	if err != nil {
		return err
	}

	writeRow := func(user *models.User) error {
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

		return nil
	}

	if err := queries.GetUsers(db, writeRow); err != nil {
		return err
	}

	return nil
}
