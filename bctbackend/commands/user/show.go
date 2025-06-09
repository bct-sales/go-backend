package user

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
)

type showUserCommand struct {
	common.Command
}

func NewUserShowCommand() *cobra.Command {
	var command *showUserCommand

	command = &showUserCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "show <user-id>",
				Short: "Show user info",
				Long:  `This command shows detailed information about a specified user.`,
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *showUserCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		// Parse the user ID from the first argument
		userId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid user ID: %s\n", args[0])
			return err
		}

		// Fetch user information from the database
		user, err := queries.GetUserWithId(db, userId)
		if err != nil {
			if errors.Is(err, database.ErrNoSuchUser) {
				c.PrintErrorf("User with ID %d does not exist.\n", userId)
				return err
			}

			c.PrintErrorf("Failed to get user with ID %d: %s\n", userId, err.Error())
			return err
		}

		// Display user information based on their role
		switch user.RoleId {
		case models.AdminRoleId:
			return c.showAdmin(user)
		case models.SellerRoleId:
			return c.showSeller(db, user)
		case models.CashierRoleId:
			return c.showCashier(db, user)
		default:
			c.PrintErrorf("Bug encountered: user has unrecognized role %d\n", user.RoleId)
			return database.ErrNoSuchRole
		}
	})
}

func (c *showUserCommand) showAdmin(user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	return formatting.PrintUser(user)
}

func (c *showUserCommand) showSeller(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")

	err := formatting.PrintUser(user)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print user data: %s", err.Error()), 1)
	}

	sellerItems, err := queries.GetSellerItems(db, user.UserId, queries.AllItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get seller items: %s", err.Error()), 1)
	}

	categoryNameTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed get category names: %s", err.Error()), 1)
	}

	pterm.DefaultSection.Println("Items")

	err = formatting.PrintItems(categoryNameTable, sellerItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print items: %s", err.Error()), 1)
	}

	return nil
}

func (c *showUserCommand) showCashier(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	if err := formatting.PrintUser(user); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print user data: %s", err.Error()), 1)
	}

	soldItems, err := queries.GetItemsSoldBy(db, user.UserId)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get items sold by cashier: %s", err.Error()), 1)
	}

	categoryNameTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get category names: %s", err.Error()), 1)
	}

	pterm.DefaultSection.Println("Sold Items")

	if err := formatting.PrintItems(categoryNameTable, soldItems); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print items: %s", err.Error()), 1)
	}

	return nil
}
