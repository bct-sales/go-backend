package session

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type AddSessionCommand struct {
	common.Command
	userId     uint64
	expiration string
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
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().Uint64VarP(&command.userId, "user", "u", 0, "ID of the user")
	command.CobraCommand.Flags().StringVarP(&command.expiration, "expiration", "e", "1", "Time before session expires")

	if err := command.CobraCommand.MarkFlagRequired("user"); err != nil {
		panic("failed to mark user as required")
	}

	if err := command.CobraCommand.MarkFlagRequired("expiration"); err != nil {
		panic("failed to mark expiration as required")
	}

	return command.AsCobraCommand()
}

func (c *AddSessionCommand) execute() error {
	expiration, err := c.parseExpiration()
	if err != nil {
		return err
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		sessionId, err := queries.AddSession(db, models.ID(c.userId), expiration)
		if err != nil {
			return err
		}

		c.Printf("%s", sessionId)

		return nil
	})
}

func (c *AddSessionCommand) parseExpiration() (models.Timestamp, error) {
	expiration := c.expiration

	numberString := expiration[0 : len(expiration)-1]
	unit := expiration[len(expiration)-1]

	number, err := strconv.ParseUint(numberString, 10, 64)
	if err != nil {
		return 0, err
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

	return 0, fmt.Errorf("invalid unit")
}
