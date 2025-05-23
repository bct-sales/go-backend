package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListUsers(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	users := []*models.User{}
	if err := queries.GetUsers(db, queries.CollectTo(&users)); err != nil {
		return fmt.Errorf("error while listing users: %w", err)
	}

	tableData := pterm.TableData{
		{"ID", "Role", "Created At", "Last Activity", "Password"},
	}

	userCount := 0

	for _, user := range users {
		idString := user.UserId.String()

		roleString, err := models.NameOfRole(user.RoleId)

		if err != nil {
			return fmt.Errorf("error while converting role to string: %w", err)
		}

		createdAtString := user.CreatedAt.FormattedDateTime()

		var lastActivityString string
		if user.LastActivity != nil {
			lastActivityString = user.LastActivity.FormattedDateTime()
		} else {
			lastActivityString = "never"
		}

		passwordString := user.Password

		tableData = append(tableData, []string{
			idString,
			roleString,
			createdAtString,
			lastActivityString,
			passwordString,
		})

		userCount++
	}

	if err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		return fmt.Errorf("error while rendering table: %w", err)
	}

	fmt.Printf("Number of users listed: %d\n", userCount)

	return nil
}
