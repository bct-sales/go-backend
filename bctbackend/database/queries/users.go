package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
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
			return &UserIdAlreadyInUseError{UserId: userId}
		}

		err = fmt.Errorf("failed to add user with id %d: %w", userId, err)
		return err
	}

	err = nil
	return err
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

	return result.LastInsertId()
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

	var roleId models.Id
	var createdAt models.Timestamp
	var lastActivity *models.Timestamp
	var password string
	err := row.Scan(&roleId, &createdAt, &lastActivity, &password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &NoSuchUserError{UserId: userId}
		}

		return nil, err
	}

	return &models.User{
		UserId:       userId,
		RoleId:       roleId,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		Password:     password,
	}, nil
}

func GetUsers(db *sql.DB, receiver func(*models.User) error) error {
	rows, err := db.Query(
		`
			SELECT user_id, role_id, created_at, password
			FROM users
		`,
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var userId models.Id
		var roleId models.Id
		var createdAt models.Timestamp
		var password string

		if err := rows.Scan(&userId, &roleId, &createdAt, &password); err != nil {
			return err
		}

		user := models.User{
			UserId:    userId,
			RoleId:    roleId,
			CreatedAt: createdAt,
			Password:  password,
		}

		if err := receiver(&user); err != nil {
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
		return &NoSuchUserError{UserId: userId}
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
		return &InvalidRoleError{UserId: userId, ExpectedRoleId: expectedRoleId}
	}

	return nil
}

// RemoveUserWithId removes a user from the database by their user ID.
// An NoSuchUserError is returned if the user does not exist.
// An error is returned if the user cannot be removed, e.g., because items or sales are
// associated with the user.
func RemoveUserWithId(db *sql.DB, userId models.Id) error {
	userExist, err := UserWithIdExists(db, userId)

	if err != nil {
		return err
	}

	if !userExist {
		return &NoSuchUserError{UserId: userId}
	}

	_, err = db.Exec(
		`
			DELETE FROM users
			WHERE user_id = $1
		`,
		userId,
	)

	return err
}
