package users

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

func fetchUsers(database *sql.DB) tea.Cmd {
	return func() tea.Msg {
		var users []*models.User

		if err := queries.GetUsers(database, queries.CollectTo(&users)); err != nil {
			return &databaseErrorMessage{
				err:     err,
				message: "failed to fetch users from database",
			}
		}

		return usersFetchedMessage{
			users: users,
		}
	}
}

func updateUserPassword(database *sql.DB, userID models.ID, password string) tea.Cmd {
	return func() tea.Msg {
		if err := queries.UpdateUserPassword(database, userID, password); err != nil {
			return &databaseErrorMessage{
				err:     err,
				message: "failed to fetch users from database",
			}
		}

		var users []*models.User
		if err := queries.GetUsers(database, queries.CollectTo(&users)); err != nil {
			return &databaseErrorMessage{
				err:     err,
				message: "failed to fetch users from database",
			}
		}

		return usersFetchedMessage{
			users: users,
		}
	}
}
