package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

func NewUserSetPasswordCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "set-password <user-id> <new-password>",
		Short: "Sets user password",
		Long:  `This command updates a user's password.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the user ID from the first argument
				userId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid user ID: %s\n", args[0])
					return err
				}
				newPassword := args[1]

				err = queries.UpdateUserPassword(db, userId, newPassword)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Failed to update user password")
					return fmt.Errorf("failed to update database: %w", err)
				}

				fmt.Fprintln(cmd.OutOrStdout(), "Password updated successfully")
				return nil
			})
		},
	}

	return &command
}
