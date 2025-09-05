package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type SetUserPasswordCommand struct {
	common.Command
	noInvalidateSessions bool
}

func NewUserSetPasswordCommand() *cobra.Command {
	var command *SetUserPasswordCommand

	command = &SetUserPasswordCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "set-password <user-id> <new-password>",
				Short: "Sets user password",
				Long:  `This command updates a user's password.`,
				Args:  cobra.ExactArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.noInvalidateSessions, "no-invalidate", false, "do not invalidate sessions associated with user")

	return command.AsCobraCommand()
}

func (c *SetUserPasswordCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		userId, err := c.ParseUserId(args[0])
		if err != nil {
			return err
		}
		newPassword := args[1]

		if err := c.updatePassword(db, userId, newPassword); err != nil {
			return err
		}

		if !c.noInvalidateSessions {
			if err := c.invalidateSessions(db, userId); err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *SetUserPasswordCommand) updatePassword(db *sql.DB, userId models.Id, newPassword string) error {
	err := queries.UpdateUserPassword(db, userId, newPassword)
	if err != nil {
		c.PrintErrorf("Failed to update user password\n")
		return fmt.Errorf("failed to update database: %w", err)
	}

	c.Printf("Password updated\n")
	return nil
}

func (c *SetUserPasswordCommand) invalidateSessions(db *sql.DB, userId models.Id) error {
	err := queries.DeleteSessionWithUser(db, userId)

	if err != nil {
		c.PrintErrorf("Failed to invalidate sessions")
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	c.Printf("Invalidated sessions associated with user\n")
	return nil
}
