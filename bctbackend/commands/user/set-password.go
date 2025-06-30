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

	return command.AsCobraCommand()
}

func (c *SetUserPasswordCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		// Parse the user ID from the first argument
		userId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid user ID: %s\n", args[0])
			return err
		}
		newPassword := args[1]

		err = queries.UpdateUserPassword(db, userId, newPassword)
		if err != nil {
			c.PrintErrorf("Failed to update user password")
			return fmt.Errorf("failed to update database: %w", err)
		}

		c.Printf("Password updated successfully")
		return nil
	})
}
