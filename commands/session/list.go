package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type ListSessionsCommand struct {
	common.Command
}

func NewListSessionsCommand() *cobra.Command {
	var command *ListSessionsCommand

	command = &ListSessionsCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "Lists sessions in database",
				Long: heredoc.Doc(`
					This command lists all sessions in the database.
					This might include some expired sessions.
					Use the prune command first to remove all expired sessions.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *ListSessionsCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		sessions, err := queries.GetSessions(db)
		if err != nil {
			return fmt.Errorf("failed to fetch sessions from database: %w", err)
		}

		tableData := pterm.TableData{
			{"Session ID", "User ID", "Expires At"},
		}

		for _, session := range sessions {
			tableData = append(tableData, []string{
				session.SessionID.String(),
				session.UserID.String(),
				session.ExpirationTime.FormattedDateTime(),
			})
		}

		return nil
	})
}
