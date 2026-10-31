package ui

import (
	"bctbackend/ui/pages/users"
	"context"
	"database/sql"
	"fmt"
	"os"

	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

func Start(ctx context.Context, database *sql.DB) error {
	logFile, err := os.Create("ui.log")
	if err != nil {
		fmt.Println("Failed to create log")
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	slog.Debug("Starting application")

	initialPage := users.New(ctx, database)

	program := tea.NewProgram(initialPage, tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}
