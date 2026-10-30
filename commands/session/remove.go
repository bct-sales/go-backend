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

type RemoveSessionCommand struct {
	common.Command
	sessionID string `exhaustruct:"optional"`
	userID    uint64 `exhaustruct:"optional"`
	all       bool   `exhaustruct:"optional"`
}

func NewRemoveSessionCommand() *cobra.Command {
	var command *RemoveSessionCommand

	command = &RemoveSessionCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "remove",
				Short: "Removes session(s) from database",
				Long: heredoc.Doc(`
					This command can be used to remove sessions from the database.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(cmd.Context())
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().StringVarP(&command.sessionID, "session", "s", "", "ID of the session to be removed")
	command.CobraCommand.Flags().Uint64VarP(&command.userID, "user", "u", 0, "ID of the user")
	command.CobraCommand.Flags().BoolVar(&command.all, "all", true, "remove all sessions")

	command.CobraCommand.MarkFlagsOneRequired("session", "user", "all")
	command.CobraCommand.MarkFlagsMutuallyExclusive("session", "user", "all")

	return command.AsCobraCommand()
}

func (c *RemoveSessionCommand) execute(ctx context.Context) error {
	if c.CobraCommand.Flags().Changed("session") {
		return c.removeBySessionID(ctx, models.SessionID(c.sessionID))
	} else if c.CobraCommand.Flags().Changed("user") {
		return c.removeByUserID(ctx, models.ID(c.userID))
	} else if c.CobraCommand.Flags().Changed("all") {
		return c.removeAll(ctx)
	} else {
		panic("should not happen; cobra is supposed to keep this from happening")
	}
}

func (c *RemoveSessionCommand) removeBySessionID(ctx context.Context, sessionID models.SessionID) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		return queries.DeleteSession(db, sessionID)
	})
}

func (c *RemoveSessionCommand) removeByUserID(ctx context.Context, userID models.ID) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		return queries.DeleteSessionWithUser(db, userID)
	})
}

func (c *RemoveSessionCommand) removeAll(ctx context.Context) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		return queries.DeleteAllSessions(db)
	})
}
