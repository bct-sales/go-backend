package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"strconv"
	"time"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListUsers(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	users, err := queries.ListUsers(db)

	if err != nil {
		return fmt.Errorf("error while listing users: %v", err)
	}

	tableData := pterm.TableData{
		{"ID", "Role", "Created At", "Password"},
	}

	userCount := 0

	for _, user := range users {
		roleString, err := models.NameOfRole(user.RoleId)

		if err != nil {
			return fmt.Errorf("error while converting role to string: %v", err)
		}

		tableData = append(tableData, []string{
			strconv.FormatInt(user.UserId, 10),
			roleString,
			time.Unix(user.Timestamp, 0).String(),
			user.Password,
		})

		userCount++
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	if err != nil {
		return fmt.Errorf("error while rendering table: %v", err)
	}

	fmt.Printf("Number of users listed: %d\n", userCount)

	return nil
}
