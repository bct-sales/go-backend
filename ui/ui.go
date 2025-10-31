package ui

import (
	"bctbackend/ui/pages/users"
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

func Start(database *sql.DB) error {
	program := tea.NewProgram(users.New(database), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}
