package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// AddUserWithId adds a user to the database with a specific user ID.
// An UserIdAlreadyInUseError is returned if the user ID is already in use.
func AddUserWithId(
	db *sql.DB,
	userId models.Id,
	roleId models.Id,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) error {

	_, err := db.Exec(
		`
			INSERT INTO users (user_id, role_id, created_at, last_activity, password)
			VALUES ($1, $2, $3, $4, $5)
		`,
		userId,
		roleId,
		createdAt,
		lastActivity,
		password,
	)

	if err != nil {
		if !models.IsValidRole(roleId) {
			return &NoSuchRoleError{RoleId: roleId}
		}

		userExists, err := UserWithIdExists(db, userId)
		if err != nil {
			return err
		}
		if userExists {
			return fmt.Errorf("trying to add user with id %d: %w", userId, ErrUserIdAlreadyInUse)
		}

		return fmt.Errorf("failed to add user with id %d: %w", userId, err)
	}

	return nil
}

func AddUser(
	db *sql.DB,
	roleId models.Id,
	createdAt models.Timestamp,
	lastActivity *models.Timestamp,
	password string) (models.Id, error) {

	result, err := db.Exec(
		`
			INSERT INTO users (role_id, created_at, last_activity, password)
			VALUES ($1, $2, $3, $4)
		`,
		roleId,
		createdAt,
		lastActivity,
		password,
	)

	if err != nil {
		if !models.IsValidRole(roleId) {
			return 0, &NoSuchRoleError{RoleId: roleId}
		}

		return 0, err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return models.Id(userId), nil
}

type AddUsersCallback func(addUser func(userId models.Id, roleId models.Id, createdAt models.Timestamp, lastActivity *models.Timestamp, password string))

func AddUsers(db *sql.DB, callback AddUsersCallback) error {
	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?)"

	add := func(userId models.Id, roleId models.Id, createdAt models.Timestamp, lastActivity *models.Timestamp, password string) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, userId, roleId, createdAt, lastActivity, password)
	}

	callback(add)

	if len(valuesString) == 0 {
		return nil
	}

	query := `INSERT INTO users (user_id, role_id, created_at, last_activity, password) VALUES ` + strings.Join(valuesString, ",")

	if _, err := db.Exec(query, arguments...); err != nil {
		return err
	}

	return nil
}

func UserWithIdExists(
	db *sql.DB,
	userId models.Id) (bool, error) {

	row := db.QueryRow(
		`
			SELECT 1
			FROM users
			WHERE user_id = $1
		`,
		userId,
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

// GetUserWithId retrieves a user from the database by their user ID.
// An NoSuchUserError is returned if the user does not exist.
func GetUserWithId(db *sql.DB, userId models.Id) (*models.User, error) {
	row := db.QueryRow(
		`
			SELECT role_id, created_at, last_activity, password
			FROM users
			WHERE user_id = $1
		`,
		userId,
	)

	user := models.User{UserId: userId}
	err := row.Scan(&user.RoleId, &user.CreatedAt, &user.LastActivity, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get user with id %d: %w", userId, ErrNoSuchUser)
		}

		return nil, err
	}

	return &user, nil
}

// GetUsers retrieves all users from the database.
func GetUsers(db *sql.DB, receiver func(*models.User) error) (r_err error) {
	rows, err := db.Query(
		`
			SELECT user_id, role_id, created_at, last_activity, password
			FROM users
		`,
	)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var userId models.Id
		var roleId models.Id
		var createdAt models.Timestamp
		var lastActivity *models.Timestamp
		var password string

		if err := rows.Scan(&userId, &roleId, &createdAt, &lastActivity, &password); err != nil {
			return err
		}

		user := models.NewUser(userId, roleId, createdAt, lastActivity, password)
		if err := receiver(user); err != nil {
			return err
		}

	}

	return nil
}

type UserWithItemCount struct {
	models.User
	ItemCount int
}

func GetUsersWithItemCount(db *sql.DB, itemSelection ItemSelection, receiver func(*UserWithItemCount) error) (r_err error) {
	query := fmt.Sprintf(
		`
			SELECT users.user_id, role_id, created_at, last_activity, password, COALESCE(COUNT(i.item_id), 0) AS item_count
			FROM users LEFT JOIN %s i ON users.user_id = i.seller_id
			GROUP BY users.user_id
			ORDER BY users.user_id
		`,
		ItemsTableFor(itemSelection))
	rows, err := db.Query(query)

	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var userWithItemCount UserWithItemCount

		if err := rows.Scan(&userWithItemCount.UserId, &userWithItemCount.RoleId, &userWithItemCount.CreatedAt, &userWithItemCount.LastActivity, &userWithItemCount.Password, &userWithItemCount.ItemCount); err != nil {
			return err
		}

		if err := receiver(&userWithItemCount); err != nil {
			return err
		}
	}

	return nil
}

// UpdateUserPassword updates the password of a user in the database by their user ID.
// An NoSuchUserError is returned if the user does not exist.
func UpdateUserPassword(db *sql.DB, userId models.Id, password string) error {
	userExists, err := UserWithIdExists(db, userId)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf("failed to update password of user %d: %w", userId, ErrNoSuchUser)
	}

	_, err = db.Exec(
		`
			UPDATE users
			SET password = $1
			WHERE user_id = $2
		`,
		password,
		userId,
	)

	return err
}

// CheckUserRole checks if a user has a specific role.
// An NoSuchUserError is returned if the user does not exist.
// A InvalidRoleError is returned if the user has a different role.
func CheckUserRole(db *sql.DB, userId models.Id, expectedRoleId models.Id) error {
	user, err := GetUserWithId(db, userId)

	if err != nil {
		return err
	}

	if user.RoleId != expectedRoleId {
		return fmt.Errorf("user %d expected to have role %d: %w", userId, expectedRoleId, ErrInvalidRole)
	}

	return nil
}

// RemoveUserWithId removes a user from the database by their user ID.
// An NoSuchUserError is returned if the user does not exist.
// An error is returned if the user cannot be removed, e.g., because items or sales are
// associated with the user.
func RemoveUserWithId(db *sql.DB, userId models.Id) error {
	{
		userExist, err := UserWithIdExists(db, userId)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to remove user with id %d: %w", userId, ErrNoSuchUser)
		}
	}

	_, err := db.Exec(
		`
			DELETE FROM users
			WHERE user_id = $1
		`,
		userId,
	)

	return err
}

func UpdateLastActivity(db *sql.DB, userId models.Id, lastActivity models.Timestamp) error {
	{
		userExist, err := UserWithIdExists(db, userId)
		if err != nil {
			return err
		}
		if !userExist {
			return fmt.Errorf("failed to update last activity of user %d: %w", userId, ErrNoSuchUser)
		}
	}

	_, err := db.Exec(
		`
			UPDATE users
			SET last_activity = $1
			WHERE user_id = $2
		`,
		lastActivity,
		userId,
	)

	return err
}

func GetSellerItemCount(db *sql.DB, sellerId models.Id) (int, error) {
	// Ensure the user exists and is a seller
	{
		seller, err := GetUserWithId(db, sellerId)
		if err != nil {
			return 0, fmt.Errorf("failed to check user in GetSellerItemCount: %w", err)
		}
		if seller.RoleId != models.SellerRoleId {
			return 0, fmt.Errorf("failed to get item count of non-seller %d: %w", sellerId, ErrInvalidRole)
		}
	}

	row := db.QueryRow(
		`
			SELECT COUNT(items.item_id)
			FROM items
			WHERE items.seller_id = $1
		`,
		sellerId,
	)

	var itemCount int
	err := row.Scan(&itemCount)

	if err != nil {
		return 0, fmt.Errorf("failed to get seller's %d item count: %w", sellerId, err)
	}

	return itemCount, nil
}

func GetSellerFrozenItemCount(db *sql.DB, sellerId models.Id) (int, error) {
	// Ensure the user exists and is a seller
	{
		seller, err := GetUserWithId(db, sellerId)
		if err != nil {
			return 0, err
		}
		if seller.RoleId != models.SellerRoleId {
			return 0, fmt.Errorf("failed to get frozen item count of non-seller %d: %w", sellerId, ErrInvalidRole)
		}
	}

	query :=
		`
			SELECT COUNT(item_id)
			FROM items
			WHERE seller_id = $1 AND frozen
		`
	row := db.QueryRow(query, sellerId)

	var itemCount int
	err := row.Scan(&itemCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get seller %d's frozen item count: %w", sellerId, err)
	}

	return itemCount, nil
}

func GetSellerHiddenItemCount(db *sql.DB, sellerId models.Id) (int, error) {
	// Ensure the user exists and is a seller
	{
		seller, err := GetUserWithId(db, sellerId)
		if err != nil {
			return 0, err
		}
		if seller.RoleId != models.SellerRoleId {
			return 0, fmt.Errorf("failed to get hidden item count of non-seller %d: %w", sellerId, ErrInvalidRole)
		}
	}

	query :=
		`
			SELECT COUNT(item_id)
			FROM items
			WHERE seller_id = $1 AND hidden
		`
	row := db.QueryRow(query, sellerId)

	var itemCount int
	err := row.Scan(&itemCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get seller's hidden item count: %w", err)
	}

	return itemCount, nil
}

func GetSellerTotalPriceOfAllItems(db *sql.DB, sellerId models.Id, itemSelection ItemSelection) (models.MoneyInCents, error) {
	// Ensure the user exists and is a seller
	{
		cashier, err := GetUserWithId(db, sellerId)
		if err != nil {
			return 0, err
		}
		if cashier.RoleId != models.SellerRoleId {
			return 0, fmt.Errorf("failed to get total price of all items of non-seller %d: %w", sellerId, ErrInvalidRole)
		}
	}

	itemTable := ItemsTableFor(itemSelection)
	query := fmt.Sprintf(
		`
			SELECT COALESCE(SUM(i.price_in_cents), 0)
			FROM %s i
			WHERE i.seller_id = $1
		`,
		itemTable,
	)
	row := db.QueryRow(query, sellerId)

	var totalPrice models.MoneyInCents
	err := row.Scan(&totalPrice)

	if err != nil {
		return 0, err
	}

	return totalPrice, nil
}
