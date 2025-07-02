package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/spf13/cobra"
)

type AddUserCommand struct {
	common.Command
	userId   int
	role     string
	password string
}

func NewUserAddCommand() *cobra.Command {
	var command *AddUserCommand

	command = &AddUserCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Add a new user",
				Long:  `This command adds a new user to the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().IntVar(&command.userId, "id", 0, "ID of the user to add")
	command.CobraCommand.Flags().StringVar(&command.role, "role", "", "Role of the user (admin, seller, cashier)")
	command.CobraCommand.Flags().StringVar(&command.password, "password", "", "Password for the user")
	command.CobraCommand.MarkFlagRequired("id")
	command.CobraCommand.MarkFlagRequired("role")
	command.CobraCommand.MarkFlagRequired("password")

	return command.AsCobraCommand()
}

func (c *AddUserCommand) execute() error {
	role := c.role
	userId := c.userId
	password := c.password

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		roleId, err := models.ParseRole(role)
		if err != nil {
			c.PrintErrorf("Invalid role; should be admin, seller or cashier\n")
			return err
		}

		timestamp := models.Now()
		var lastActivity *models.Timestamp = nil

		if err := queries.AddUserWithId(db, models.Id(userId), roleId, timestamp, lastActivity, password); err != nil {
			c.PrintErrorf("Failed to add user\n")
			return err
		}

		c.Printf("User added successfully\n")
		return nil
	})
}
