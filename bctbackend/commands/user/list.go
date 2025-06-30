package user

import (
	"bctbackend/commands/common"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewUserListCommand() *cobra.Command {
	var format string

	command := cobra.Command{
		Use:   "list",
		Short: "List all users",
		Long:  `This command lists all users in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "table":
				return listUsersInTableFormat(cmd)
			case "csv":
				return listUsersInCSVFormat(cmd)
			default:
				fmt.Fprintf(cmd.ErrOrStderr(), "Invalid format: %s\n", format)
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	command.Flags().StringVar(&format, "format", "table", "Output format (table, csv)")

	return &command
}

func listUsersInTableFormat(cmd *cobra.Command) error {
	return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
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
			roleString := user.RoleId.Name()
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

		if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
			return fmt.Errorf("error while rendering table: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Number of users listed: %d\n", userCount)

		return nil
	})
}

func listUsersInCSVFormat(cmd *cobra.Command) error {
	return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
		err := dbcsv.OutputUsers(db, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to output users: %w", err)
		}

		return nil
	})
}
