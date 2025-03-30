package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListUsers(databasePath string) (err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	users := []*models.User{}
	if err := queries.GetUsers(db, queries.CollectTo(&users)); err != nil {
		err = fmt.Errorf("error while listing users: %v", err)
		return err
	}

	tableData := pterm.TableData{
		{"ID", "Role", "Created At", "Last Activity", "Password"},
	}

	userCount := 0

	for _, user := range users {
		idString := strconv.FormatInt(user.UserId, 10)

		roleString, err := models.NameOfRole(user.RoleId)

		if err != nil {
			err = fmt.Errorf("error while converting role to string: %v", err)
			return err
		}

		createdAtString := time.Unix(user.CreatedAt, 0).String()

		var lastActivityString string
		if user.LastActivity != nil {
			lastActivityString = time.Unix(*user.LastActivity, 0).String()
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
		err = fmt.Errorf("error while rendering table: %v", err)
		return err
	}

	fmt.Printf("Number of users listed: %d\n", userCount)

	err = nil
	return err
}
