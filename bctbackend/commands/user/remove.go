package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewUserRemoveCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "remove <user-id>",
		Short: "Removes a user",
		Long: heredoc.Doc(`
				This command permanently removes a user from the database.
				This command is only provided for completeness but is
				not intended to be actually used.

				It is not possible to remove a user that has items associated with them.
			   `),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the user ID from the first argument
				userId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid user ID: %s\n", args[0])
					return err
				}

				if err = queries.RemoveUserWithId(db, userId); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Failed to remove user")
					return err
				}

				fmt.Fprintln(cmd.OutOrStdout(), "User removed successfully")
				return nil
			})
		},
	}

	return &command
}
