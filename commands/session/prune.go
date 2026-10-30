package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type PruneSessionsCommand struct {
	common.Command
}

func NewPruneSessionsCommand() *cobra.Command {
	var command *PruneSessionsCommand

	command = &PruneSessionsCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "prune",
				Short: "Prunes expired sessions",
				Long: heredoc.Doc(`
					This command removes all expired sessions from the database.
					Executing this command should not have any observable effect
					other than slightly reducing the database's size.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(cmd.Context())
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *PruneSessionsCommand) execute(ctx context.Context) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		cutOff := models.Now()

		if err := queries.DeleteExpiredSessions(db, cutOff); err != nil {
			return err
		}

		return nil
	})
}
