package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type AddUserCommand struct {
	common.Command
	userID           int    `exhaustruct:"optional"`
	role             string `exhaustruct:"optional"`
	password         string `exhaustruct:"optional"`
	generatePassword bool   `exhaustruct:"optional"`
}

func NewUserAddCommand() *cobra.Command {
	var command *AddUserCommand

	command = &AddUserCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Add a new user",
				Long: heredoc.Doc(`
					This command adds a new user to the database.
				`),
				Example: heredoc.Doc(`
					# Adds an admin
					bctbackend user add --id 1 --role admin --password azer1234

					# Adds a seller with an auto-generated password
					bctbackend user add --id 100 --role seller --generate-password
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().IntVar(&command.userID, "id", 0, "ID of the user to add")
	command.CobraCommand.Flags().StringVar(&command.role, "role", "", "Role of the user (admin, seller, cashier)")
	command.CobraCommand.Flags().StringVar(&command.password, "password", "", "Password for the user")
	command.CobraCommand.Flags().BoolVar(&command.generatePassword, "generate-password", false, "Generate password for the user")
	if err := command.CobraCommand.MarkFlagRequired("id"); err != nil {
		panic("failed to mark id as required")
	}
	if err := command.CobraCommand.MarkFlagRequired("role"); err != nil {
		panic("failed to mark role as required")
	}
	command.CobraCommand.MarkFlagsOneRequired("password", "generate-password")
	command.CobraCommand.MarkFlagsMutuallyExclusive("password", "generate-password")

	return command.AsCobraCommand()
}

func (c *AddUserCommand) execute() error {
	role := c.role
	userID := c.userID

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		roleID, err := models.ParseRole(role)
		if err != nil {
			c.PrintErrorf("Invalid role; should be admin, seller or cashier\n")
			return err
		}

		var password string
		if c.CobraCommand.Flags().Changed("password") {
			password = c.password
		} else if c.CobraCommand.Flags().Changed("generate-password") {
			var err error
			password, err = common.FindUnusedPassword(db)
			if err != nil {
				return err
			}
		}

		timestamp := models.Now()
		var lastActivity *models.Timestamp = nil

		if err := queries.AddUserWithID(db, models.ID(userID), roleID, timestamp, lastActivity, password); err != nil {
			c.PrintErrorf("Failed to add user\n")
			return err
		}

		c.Printf("User added successfully\n")
		return nil
	})
}
