package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type AddSessionCommand struct {
	common.Command
	userID     uint64 `exhaustruct:"optional"`
	expiration string `exhaustruct:"optional"`
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
	command.CobraCommand.Flags().StringVarP(&command.expiration, "expiration", "e", "1h", "Time before session expires (number followed by s, m, h, or d)")

	if err := command.CobraCommand.MarkFlagRequired("user"); err != nil {
		panic("failed to mark user as required")
	}

	if err := command.CobraCommand.MarkFlagRequired("expiration"); err != nil {
		panic("failed to mark expiration as required")
	}

	return command.AsCobraCommand()
}

func (c *AddSessionCommand) execute(ctx context.Context) error {
	expiration, err := c.parseExpiration()
	if err != nil {
		return err
	}

	return c.WithOpenedDatabase(ctx, func(ctx context.Context, db *sql.DB) error {
		sessionID, err := queries.AddSession(ctx, db, models.ID(c.userID), expiration)
		if err != nil {
			return err
		}

		c.Printf("%s", sessionID)

		return nil
	})
}

func (c *AddSessionCommand) parseExpiration() (models.Timestamp, error) {
	expiration := c.expiration
	if len(expiration) < 2 {
		return 0, fmt.Errorf("invalid expiration %q: expected a number followed by s, m, h, or d", expiration)
	}

	numberString := expiration[0 : len(expiration)-1]
	unit := expiration[len(expiration)-1]

	number, err := strconv.ParseUint(numberString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid expiration %q: %w", expiration, err)
	}

	multiplier, err := c.parseTimeUnit(unit)
	if err != nil {
		return 0, err
	}

	result := models.Now() + models.Timestamp(number*multiplier)
	return result, nil
}

func (c *AddSessionCommand) parseTimeUnit(unit byte) (uint64, error) {
	switch unit {
	case 's':
		return 1, nil

	case 'm':
		return 60, nil

	case 'h':
		return 60 * 60, nil

	case 'd':
		return 24 * 60 * 60, nil
	}

	return 0, fmt.Errorf("invalid expiration unit %q: expected s, m, h, or d", unit)
}
