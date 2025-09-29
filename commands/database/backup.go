package database

import (
	"bctbackend/commands/common"
	"database/sql"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type backupDatabaseCommand struct {
	common.Command
}

func NewDatabaseBackupCommand() *cobra.Command {
	var command *backupDatabaseCommand

	command = &backupDatabaseCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "backup <filename>",
				Short: "Backup the database",
				Args:  cobra.MaximumNArgs(1),
				Long: heredoc.Doc(`
					This command makes a copy of the current database.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *backupDatabaseCommand) execute(args []string) error {
	targetPath := c.determineDatabaseName(args)

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if _, err := db.Exec("VACUUM INTO ?", targetPath); err != nil {
			c.PrintErrorf("Failed to backup database\n")
			return fmt.Errorf("failed to backup database %s: %w", targetPath, err)
		}

		c.Printf("Database backup completed successfully to %s\n", targetPath)
		return nil
	})
}

func (c *backupDatabaseCommand) determineDatabaseName(args []string) string {
	if len(args) == 1 {
		return args[0]
	} else {
		now := time.Now()
		formattedDate := now.Format("2006-01-02-03-04-05")
		return fmt.Sprintf("backup-%s.db", formattedDate))
	}
}
