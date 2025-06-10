package user

import (
	"bctbackend/commands/common"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

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
		return models.VisitRole(user.RoleId, &showUser{command: c, database: db, user: user})
	})
}

type showUser struct {
	command  *showUserCommand
	database *sql.DB
	user     *models.User
}

func (s *showUser) Admin() error {
	return s.command.showAdmin(s.user)
}

func (s *showUser) Seller() error {
	return s.command.showSeller(s.database, s.user)
}

func (s *showUser) Cashier() error {
	return s.command.showCashier(s.database, s.user)
}

func (c *showUserCommand) showAdmin(user *models.User) error {
	pterm.DefaultSection.Println("User Data")

	if err := c.printUserTable(user); err != nil {
		return err
	}

	return nil
}

func (c *showUserCommand) showSeller(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")

	if err := c.printUserTable(user); err != nil {
		return err
	}

	sellerItems, err := queries.GetSellerItems(db, user.UserId, queries.AllItems)
	if err != nil {
		c.PrintErrorf("Failed to get seller items\n")
		return err
	}

	categoryNameTable, err := c.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Items")

	err = c.printItems(categoryNameTable, sellerItems)
	if err != nil {
		return err
	}

	return nil
}

func (c *showUserCommand) showCashier(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	if err := c.printUserTable(user); err != nil {
		return err
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

	if err := c.printItems(categoryNameTable, soldItems); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to print items: %s", err.Error()), 1)
	}

	return nil
}

func (c *showUserCommand) printUserTable(user *models.User) error {
	var lastActivity string
	if user.LastActivity != nil {
		lastActivity = user.LastActivity.FormattedDateTime()
	} else {
		lastActivity = "N/A"
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"ID", user.UserId.String()},
		{"Role", user.RoleId.Name()},
		{"Created At", user.CreatedAt.FormattedDateTime()},
		{"Last Activity", lastActivity},
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		c.PrintErrorf("Failed to render user table\n")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func (c *showUserCommand) printItems(categoryNameTable map[models.Id]string, items []*models.Item) error {
	tableData := pterm.TableData{
		{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity", "Added At", "Frozen", "Hidden"},
	}

	for _, item := range items {
		categoryName, ok := categoryNameTable[item.CategoryID]
		if !ok {
			c.PrintErrorf("No category found for item with ID %d\n", item.ItemID)
			return database.ErrNoSuchCategory
		}

		tableData = append(tableData, []string{
			item.ItemID.String(),
			item.Description,
			item.PriceInCents.DecimalNotation(),
			categoryName,
			item.SellerID.String(),
			strconv.FormatBool(item.Donation),
			strconv.FormatBool(item.Charity),
			item.AddedAt.FormattedDateTime(),
			strconv.FormatBool(item.Frozen),
			strconv.FormatBool(item.Hidden),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		c.PrintErrorf("Failed to render items table\n")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
