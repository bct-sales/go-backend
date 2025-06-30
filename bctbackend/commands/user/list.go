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

type ListUsersCommand struct {
	common.Command
	format string
}

func NewUserListCommand() *cobra.Command {
	var command *ListUsersCommand

	command = &ListUsersCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "List all users",
				Long:  `This command lists all users in the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	command.CobraCommand.Flags().StringVar(&command.format, "format", "table", "Output format (table, csv)")

	return command.AsCobraCommand()
}

func (c *ListUsersCommand) execute(args []string) error {
	switch c.format {
	case "table":
		return c.listUsersInTableFormat()
	case "csv":
		return c.listUsersInCSVFormat()
	default:
		c.PrintErrorf("Invalid format: %s\n", c.format)
		return fmt.Errorf("unknown format: %s", c.format)
	}
}

func (c *ListUsersCommand) listUsersInTableFormat() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
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
			c.PrintErrorf("Failed to render table: %v\n", err)
			return fmt.Errorf("error while rendering table: %w", err)
		}

		c.Printf("Number of users listed: %d\n", userCount)

		return nil
	})
}

func (c *ListUsersCommand) listUsersInCSVFormat() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		err := dbcsv.OutputUsers(db, os.Stdout)
		if err != nil {
			c.PrintErrorf("Failed to output users: %v\n", err)
			return fmt.Errorf("failed to output users: %w", err)
		}

		return nil
	})
}
