package user

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
	_ "modernc.org/sqlite"
)

func ShowUser(databasePath string, userId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		if errors.Is(err, queries.ErrNoSuchUser) {
			return cli.Exit("User with the given id does not exist", 1)
		}

		return err
	}

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
}

func showAdmin(user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	return formatting.PrintUser(user)
}

func showSeller(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")

	err := formatting.PrintUser(user)
	if err != nil {
		return err
	}

	sellerItems, err := queries.GetSellerItems(db, user.UserId, queries.AllItems)
	if err != nil {
		return err
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Items")

	err = formatting.PrintItems(categoryTable, sellerItems)
	if err != nil {
		return err
	}

	return nil
}

func showCashier(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	if err := formatting.PrintUser(user); err != nil {
		return err
	}

	soldItems, err := queries.GetItemsSoldBy(db, user.UserId)
	if err != nil {
		return err
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Sold Items")

	if err := formatting.PrintItems(categoryTable, soldItems); err != nil {
		return err
	}

	return nil
}
