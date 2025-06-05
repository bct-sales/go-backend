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

func NewUserShowCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "show ID",
		Short: "Show user info",
		Long:  `This command shows detailed information about a specified user.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the user ID from the first argument
				userId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid user ID: %s\n", args[0])
					return err
				}

				// Fetch user information from the database
				user, err := queries.GetUserWithId(db, userId)
				if err != nil {
					if errors.Is(err, database.ErrNoSuchUser) {
						return cli.Exit("User with the given id does not exist", 1)
					}

					return cli.Exit(fmt.Sprintf("Failed to get user: %s", err.Error()), 1)
				}

				// Display user information based on their role
				switch user.RoleId {
				case models.AdminRoleId:
					return showAdmin(user)
				case models.SellerRoleId:
					return showSeller(db, user)
				case models.CashierRoleId:
					return showCashier(db, user)
				default:
					return cli.Exit(fmt.Sprintf("Bug encountered: user has unrecognized role %d", user.RoleId), 1)
				}
			})
		},
	}

	return &command
}

func showAdmin(user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	return formatting.PrintUser(user)
}

func showSeller(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")

	err := formatting.PrintUser(user)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print user data: %s", err.Error()), 1)
	}

	sellerItems, err := queries.GetSellerItems(db, user.UserId, queries.AllItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get seller items: %s", err.Error()), 1)
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed get category names: %s", err.Error()), 1)
	}

	pterm.DefaultSection.Println("Items")

	err = formatting.PrintItems(categoryTable, sellerItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print items: %s", err.Error()), 1)
	}

	return nil
}

func showCashier(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	if err := formatting.PrintUser(user); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print user data: %s", err.Error()), 1)
	}

	soldItems, err := queries.GetItemsSoldBy(db, user.UserId)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get items sold by cashier: %s", err.Error()), 1)
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get category names: %s", err.Error()), 1)
	}

	pterm.DefaultSection.Println("Sold Items")

	if err := formatting.PrintItems(categoryTable, soldItems); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print items: %s", err.Error()), 1)
	}

	return nil
}
