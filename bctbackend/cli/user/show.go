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
	_ "modernc.org/sqlite"
)

func ShowUser(databasePath string, userId models.Id) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		var noSuchUserError *queries.NoSuchUserError
		if errors.As(err, &noSuchUserError) {
			return fmt.Errorf("user with id %d does not exist", userId)
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
		return fmt.Errorf("unknown role id: %d", user.RoleId)
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

	sellerItems, err := queries.GetSellerItems(db, user.UserId)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Items")

	err = formatting.PrintItems(sellerItems)
	if err != nil {
		return err
	}

	return nil
}

func showCashier(db *sql.DB, user *models.User) error {
	pterm.DefaultSection.Println("User Data")
	formatting.PrintUser(user)

	soldItems, err := queries.GetItemsSoldBy(db, user.UserId)
	if err != nil {
		return err
	}

	pterm.DefaultSection.Println("Sold Items")

	formatting.PrintItems(soldItems)

	return nil
}
