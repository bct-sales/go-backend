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

		if UserWithIdExists(db, userId) {
			return &UserIdAlreadyInUseError{UserId: userId}
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

	return result.LastInsertId()
}

func UserWithIdExists(
	db *sql.DB,
	userId models.Id) bool {

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

	return err == nil
}

// GetUserWithId retrieves a user from the database by their user ID.
// An UnknownUserError is returned if the user does not exist.
func GetUserWithId(db *sql.DB, userId models.Id) (models.User, error) {
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
		var dummyResult models.User

		if errors.Is(err, sql.ErrNoRows) {
			return dummyResult, &UnknownUserError{UserId: userId}
		}

		return dummyResult, err
	}

	return models.User{
		UserId:       userId,
		RoleId:       roleId,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		Password:     password,
	}, nil
}

func GetUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query(
		`
			SELECT user_id, role_id, created_at, password
			FROM users
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []models.User{}

	for rows.Next() {
		var userId models.Id
		var roleId models.Id
		var createdAt models.Timestamp
		var password string

		err = rows.Scan(&userId, &roleId, &createdAt, &password)

		if err != nil {
			return nil, err
		}

		user := models.User{
			UserId:    userId,
			RoleId:    roleId,
			CreatedAt: createdAt,
			Password:  password,
		}
		users = append(users, user)
	}

	return users, nil
}

func UpdateUserPassword(db *sql.DB, userId models.Id, password string) error {
	_, err := db.Exec(
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
// An UnknownUserError is returned if the user does not exist.
func CheckUserRole(db *sql.DB, userId models.Id, expectedRoleId models.Id) (bool, error) {
	user, err := GetUserWithId(db, userId)

	if err != nil {
		return false, err
	}

	if user.RoleId != expectedRoleId {
		return false, nil
	}

	return true, nil
}
