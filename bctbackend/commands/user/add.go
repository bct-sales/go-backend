package user

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

func NewUserAddCommand() *cobra.Command {
	var userId int
	var role string
	var password string

	command := cobra.Command{
		Use:   "add",
		Short: "Add a new user",
		Long:  `This command adds a new user to the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				roleId, err := models.ParseRole(role)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Invalid role; should be admin, seller or cashier")
					return err
				}

				timestamp := models.Now()
				var lastActivity *models.Timestamp = nil

				if err := queries.AddUserWithId(db, models.Id(userId), roleId, timestamp, lastActivity, password); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Failed to add user")
					return err
				}

				fmt.Fprintln(cmd.OutOrStdout(), "User added successfully")
				return nil
			})
		},
	}

	command.Flags().IntVar(&userId, "id", 0, "ID of the user to add")
	command.Flags().StringVar(&role, "role", "", "Role of the user (admin, seller, cashier)")
	command.Flags().StringVar(&password, "password", "", "Password for the user")
	command.MarkFlagRequired("id")
	command.MarkFlagRequired("role")
	command.MarkFlagRequired("password")

	return &command
}
