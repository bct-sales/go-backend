package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/meta"
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
)

// AddUserWithID adds a user to the database with a specific user ID.
// An ErrUserIDAlreadyInUse is returned if the user ID is already in use.
// An ErrNoSuchRole is returned if the role ID is invalid.
func AddUserWithID(
	database DatabaseQuerier,
	userID models.ID,
	roleID models.RoleID,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) (r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !roleID.IsValid() {
		return fmt.Errorf("invalid role id %d: %w", roleID.ID, dberr.ErrNoSuchRole)
	}

	_, err := database.Exec(
		`
			INSERT INTO users (user_id, role_id, created_at, last_activity, password)
			VALUES ($1, $2, $3, $4, $5)
		`,
		userID,
		roleID.Int64(),
		createdAt,
		lastActivity,
		password,
	)

	if err != nil {
		userExists, err := UserWithIDExists(database, userID)
		if err != nil {
			return err
		}
		if userExists {
			return fmt.Errorf("trying to add user with id %d: %w", userID, dberr.ErrIDAlreadyInUse)
		}

		return fmt.Errorf("failed to add user with id %d: %w", userID, err)
	}

	return nil
}

func AddUser(
	db DatabaseQuerier,
	roleID models.RoleID,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) (r_result models.ID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !roleID.IsValid() {
		return 0, fmt.Errorf("invalid role id %d: %w", roleID.ID, dberr.ErrNoSuchRole)
	}

	result, err := db.Exec(
		`
			INSERT INTO users (role_id, created_at, last_activity, password)
			VALUES ($1, $2, $3, $4)
		`,
		roleID.Int64(),
		createdAt,
		lastActivity,
		password,
	)

	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return models.ID(userID), nil
}

type AddUsersCallback func(addUser func(userID models.ID, roleID models.RoleID, createdAt models.Timestamp, lastActivity *models.Timestamp, password string))

func AddUsers(database DatabaseQuerier, callback AddUsersCallback) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?)"

	add := func(userID models.ID, roleID models.RoleID, createdAt models.Timestamp, lastActivity *models.Timestamp, password string) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, userID, roleID.Int64(), createdAt, lastActivity, password)
	}

	callback(add)

	if len(valuesString) == 0 {
		return nil
	}

	query := `INSERT INTO users (user_id, role_id, created_at, last_activity, password) VALUES ` + strings.Join(valuesString, ",")

	if _, err := database.Exec(query, arguments...); err != nil {
		return err
	}

	return nil
}

func UserWithIDExists(db DatabaseQuerier, userID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT
				1
			FROM
				users
			WHERE
				user_id = $1
		`,
		userID,
	)

	var value int
	err := row.Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetUserWithID retrieves a user from the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func GetUserWithID(db DatabaseQuerier, userID models.ID) (r_result *models.User, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT
				role_id,
				created_at,
				last_activity,
				password
			FROM
				users
			WHERE
				user_id = $1
		`,
		userID,
	)

	var roleID models.RoleID
	var createdAt models.Timestamp
	var lastActivity *models.Timestamp
	var password string
	err := row.Scan(&roleID.ID, &createdAt, &lastActivity, &password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get user with id %d: %w", userID, dberr.ErrNoSuchUser)
		}

		return nil, err
	}

	user := models.User{UserID: userID, RoleID: roleID, CreatedAt: createdAt, LastActivity: lastActivity, Password: password}
	return &user, nil
}

// GetUsers retrieves all users from the database.
func GetUsers(db DatabaseQuerier, receiver func(*models.User) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT
				user_id,
				role_id,
				created_at,
				last_activity,
				password
			FROM
				users
		`,
	)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var userID models.ID
		var roleID models.RoleID
		var createdAt models.Timestamp
		var lastActivity *models.Timestamp
		var password string

		if err := rows.Scan(&userID, &roleID.ID, &createdAt, &lastActivity, &password); err != nil {
			return err
		}

		user := models.User{
			UserID:       userID,
			RoleID:       roleID,
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
			Password:     password,
		}
		if err := receiver(&user); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}

type UserWithItemCount struct {
	models.User
	ItemCount int
}

func GetUsersWithItemCount(db DatabaseQuerier, itemSelection ItemSelection, receiver func(*UserWithItemCount) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := fmt.Sprintf(
		`
			SELECT
				users.user_id,
				role_id,
				created_at,
				last_activity,
				password,
				COALESCE(COUNT(i.item_id), 0) AS item_count
			FROM
				users
			LEFT JOIN %s i ON users.user_id = i.seller_id
			GROUP BY
				users.user_id
			ORDER BY
				users.user_id
		`,
		ItemsTableFor(itemSelection))
	rows, err := db.Query(query)

	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var userID models.ID
		var roleID models.RoleID
		var createdAt models.Timestamp
		var lastActivity *models.Timestamp
		var password string
		var itemCount int
		if err := rows.Scan(&userID, &roleID.ID, &createdAt, &lastActivity, &password, &itemCount); err != nil {
			return err
		}

		userWithItemCount := UserWithItemCount{
			User: models.User{
				UserID:       userID,
				RoleID:       roleID,
				CreatedAt:    createdAt,
				LastActivity: lastActivity,
				Password:     password,
			},
			ItemCount: itemCount,
		}
		if err := receiver(&userWithItemCount); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}

// UpdateUserPassword updates the password of a user in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func UpdateUserPassword(database DatabaseQuerier, userID models.ID, password string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	userExists, err := UserWithIDExists(database, userID)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf("failed to update password of user %d: %w", userID, dberr.ErrNoSuchUser)
	}

	_, err = database.Exec(
		`
			UPDATE users
			SET password = $1
			WHERE user_id = $2
		`,
		password,
		userID,
	)

	return err
}

// EnsureUserExists checks if a user exists in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func EnsureUserExists(db DatabaseQuerier, userID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	userExists, err := UserWithIDExists(db, userID)
	if err != nil {
		return fmt.Errorf("failed to ensure user %d exists: %w", userID, err)
	}
	if !userExists {
		return fmt.Errorf("failed to ensure user %d exists: %w", userID, dberr.ErrNoSuchUser)
	}
	return nil
}

// EnsureUserExistsAndHasRole checks if a user has a specific role.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user has a different role.
func EnsureUserExistsAndHasRole(db DatabaseQuerier, userID models.ID, expectedRoleID models.RoleID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	user, err := GetUserWithID(db, userID)

	if err != nil {
		return err
	}

	if user.RoleID != expectedRoleID {
		return fmt.Errorf("user %d expected to have role %s but is %s instead: %w", userID, expectedRoleID.Name(), user.RoleID.Name(), dberr.ErrWrongRole)
	}

	return nil
}

// RemoveUserWithID removes a user from the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
// An error is returned if the user cannot be removed, e.g., because items or sales are
// associated with the user.
func RemoveUserWithID(db DatabaseQuerier, userID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	{
		userExist, err := UserWithIDExists(db, userID)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to remove user with id %d: %w", userID, dberr.ErrNoSuchUser)
		}
	}

	_, err := db.Exec(
		`
			DELETE FROM users
			WHERE user_id = $1
		`,
		userID,
	)

	return err
}

// UpdateLastActivity updates the last activity timestamp of a user in the database by their user ID.
// An ErrNoSuchUser is returned if the user does not exist.
func UpdateLastActivity(db DatabaseQuerier, userID models.ID, lastActivity models.Timestamp) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	{
		userExist, err := UserWithIDExists(db, userID)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to update last activity of user %d: %w", userID, dberr.ErrNoSuchUser)
		}
	}

	_, err := db.Exec(
		`
			UPDATE users
			SET last_activity = $1
			WHERE user_id = $2
		`,
		lastActivity,
		userID,
	)

	return err
}

type GetSellerItemCountFlag int

const (
	IncludeAll GetSellerItemCountFlag = iota
	Exclude
	IncludeOnly
)

// CountSellerItems returns the number of sold items owned by a given seller.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user with the given id is not a seller.
// The frozen and hidden parameters allow to specify whether to include or exclude frozen and hidden items.
// For example, if frozen equals IncludeAll, both frozen and unfrozen items are included.
// If frozen equals Exclude, only unfrozen items are included.
// If frozen equals IncludeOnly, only frozen items are included.
func CountSellerItems(db DatabaseQuerier, sellerID models.ID, frozen GetSellerItemCountFlag, hidden GetSellerItemCountFlag) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("failed to get hidden item count of user %d: %w", sellerID, err)
	}

	query := sq.Select("COUNT(item_id)").From("items").Where(sq.Eq{"seller_id": sellerID})

	switch frozen {
	case IncludeAll:
		// No additional condition needed, all items are included
	case Exclude:
		query = query.Where(sq.Eq{"frozen": false})
	case IncludeOnly:
		query = query.Where(sq.Eq{"frozen": true})
	}

	switch hidden {
	case IncludeAll:
		// No additional condition needed, all items are included
	case Exclude:
		query = query.Where(sq.Eq{"hidden": false})
	case IncludeOnly:
		query = query.Where(sq.Eq{"hidden": true})
	}

	queryString, queryArguments, sqlErr := query.ToSql()
	if sqlErr != nil {
		return 0, fmt.Errorf("failed to build SQL query")
	}
	row := db.QueryRow(queryString, queryArguments...)

	var itemCount int
	err := row.Scan(&itemCount)

	if err != nil {
		return 0, fmt.Errorf("failed to get seller's %d item count: %w", sellerID, err)
	}

	return itemCount, nil
}

// GetSellerTotalValueOfAllItems returns the total value of all items owned by a given seller.
// Note that whether an item has been sold does not matter.
// An ErrNoSuchUser is returned if the user does not exist.
// An ErrWrongRole is returned if the user with the given id is not a seller.
func GetSellerTotalValueOfAllItems(db DatabaseQuerier, sellerID models.ID, itemSelection ItemSelection) (r_result models.MoneyInCents, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Ensure the user exists and is a seller
	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("failed to get total price of all items of user %d: %w", sellerID, err)
	}

	itemTable := ItemsTableFor(itemSelection)
	query := fmt.Sprintf(
		`
			SELECT
				COALESCE(SUM(i.price_in_cents), 0)
			FROM
				%s i
			WHERE
				i.seller_id = $1
		`,
		itemTable,
	)
	row := db.QueryRow(query, sellerID)

	var totalPrice models.MoneyInCents
	err := row.Scan(&totalPrice)

	if err != nil {
		return 0, err
	}

	return totalPrice, nil
}

func CollectPasswords(db DatabaseQuerier) (r_result []string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := squirrel.Select(meta.User.Password).From(meta.User.Table)
	rows, err := query.RunWith(db).Query()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch passwords from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var password string

		if err := rows.Scan(&password); err != nil {
			return nil, fmt.Errorf("failed to scan row while fetching password: %w", err)
		}

		r_result = append(r_result, password)
	}

	return r_result, nil
}
