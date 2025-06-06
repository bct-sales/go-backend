package database

import (
	"bctbackend/commands/common"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewDatabaseBackupCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "backup <filename>",
		Short: "Backup the database",
		Args:  cobra.ExactArgs(1),
		Long: heredoc.Doc(`
			This command makes a copy of the current database.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := args[0]

			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				if _, err := db.Exec("VACUUM INTO ?", targetPath); err != nil {
					return fmt.Errorf("failed to backup database %s: %w", targetPath, err)
				}

				fmt.Fprintln(cmd.OutOrStdout(), "Database backup completed successfully")
				return nil
			})
		},
	}

	return &command
}
