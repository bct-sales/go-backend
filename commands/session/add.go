package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
	"database/sql"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type AddSessionCommand struct {
	common.Command
	userID     uint64        `exhaustruct:"optional"`
	expiration time.Duration `exhaustruct:"optional"`
}

func NewAddSessionCommand() *cobra.Command {
	var command *AddSessionCommand

	command = &AddSessionCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Adds sessions to database",
				Long: heredoc.Doc(`
					This command adds a new session to the database.
					Adding a session will not have any effect:
					a user still needs to have the session id set in their cookie.
					Only provided for debugging purposes, not very useful in other contexts.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(cmd.Context())
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().Uint64VarP(&command.userID, "user", "u", 0, "ID of the user")
	command.CobraCommand.Flags().DurationVarP(&command.expiration, "expiration", "e", 24*time.Hour, "Time before session expires (default: 1 day)")

	if err := command.CobraCommand.MarkFlagRequired("user"); err != nil {
		panic("failed to mark user as required")
	}

	return command.AsCobraCommand()
}

func (c *AddSessionCommand) execute(ctx context.Context) error {
	expirationTime := models.Now().AddDuration(c.expiration)

	return c.WithOpenedDatabase(ctx, func(ctx context.Context, db *sql.DB) error {
		sessionID, err := queries.AddSession(ctx, db, models.ID(c.userID), expirationTime)
		if err != nil {
			return err
		}

		c.Printf("%s", sessionID)

		return nil
	})
}
