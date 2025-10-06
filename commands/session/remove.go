package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type RemoveSessionCommand struct {
	common.Command
	sessionID string
	userID    uint64
	all       bool
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
					return command.execute()
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

func (c *RemoveSessionCommand) execute() error {
	if c.CobraCommand.Flags().Changed("session") {
		return c.removeBySessionID(models.SessionID(c.sessionID))
	} else if c.CobraCommand.Flags().Changed("user") {
		return c.removeByUserID(models.ID(c.userID))
	} else if c.CobraCommand.Flags().Changed("all") {
		return c.removeAll()
	} else {
		panic("should not happen; cobra is supposed to keep this from happening")
	}
}

func (c *RemoveSessionCommand) removeBySessionID(sessionID models.SessionID) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		return queries.DeleteSession(db, sessionID)
	})
}

func (c *RemoveSessionCommand) removeByUserID(userID models.ID) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		return queries.DeleteSessionWithUser(db, userID)
	})
}

func (c *RemoveSessionCommand) removeAll() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		return queries.DeleteAllSessions(db)
	})
}
