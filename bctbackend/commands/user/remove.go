package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type RemoveUserCommand struct {
	common.Command
}

func NewUserRemoveCommand() *cobra.Command {
	var command *RemoveUserCommand

	command = &RemoveUserCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
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
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *RemoveUserCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		userId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid user ID: %s\n", args[0])
			return err
		}

		if err = queries.RemoveUserWithId(db, userId); err != nil {
			c.PrintErrorf("Failed to remove user")
			return err
		}

		c.Printf("User removed successfully")

		return nil
	})
}
