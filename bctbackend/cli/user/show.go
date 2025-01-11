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

func ShowUser(databasePath string, userId models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	user, err := queries.GetUserWithId(db, userId)

	if err != nil {
		var unknownUserError *queries.UnknownUserError
		if errors.As(err, &unknownUserError) {
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
		return showCashier(user)
	default:
		return fmt.Errorf("unknown role id: %d", user.RoleId)
	}
}

func showAdmin(user models.User) error {
	return formatting.PrintUser(user)
}

func showSeller(db *sql.DB, user models.User) error {
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

	err = formatting.PrintSellerItems(sellerItems)
	if err != nil {
		return err
	}

	return nil
}

func showCashier(user models.User) error {
	return formatting.PrintUser(user)
}
