package cli

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"strconv"
	"time"

	"database/sql"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func listUsers(databasePath string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("error while opening database: %v", err)
	}

	defer db.Close()

	users, err := queries.ListUsers(db)

	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"ID", "Role", "Created At", "Password"},
	}

	for _, user := range users {
		roleString, err := models.RoleToString(user.RoleId)

		if err != nil {
			return err
		}

		tableData = append(tableData, []string{
			strconv.FormatInt(user.UserId, 10),
			roleString,
			time.Unix(user.Timestamp, 0).String(),
			user.Password,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	return nil
}
