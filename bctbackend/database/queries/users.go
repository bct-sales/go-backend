package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

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

	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, &UnknownUserError{UserId: userId}
	}

	return models.User{
		UserId:       userId,
		RoleId:       roleId,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		Password:     password,
	}, nil
}

func ListUsers(db *sql.DB) ([]models.User, error) {
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

		users = append(users, models.User{
			UserId:    userId,
			RoleId:    roleId,
			CreatedAt: createdAt,
			Password:  password,
		})
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

func CheckUserRole(db *sql.DB, userId models.Id, expectedRoleId models.Id) error {
	user, err := GetUserWithId(db, userId)

	if err != nil {
		return err
	}

	if user.RoleId != expectedRoleId {
		expectedRoleName, err1 := models.NameOfRole(expectedRoleId)
		actualRoleName, err2 := models.NameOfRole(user.RoleId)

		if joinedError := errors.Join(err1, err2); joinedError != nil {
			return joinedError
		}

		return fmt.Errorf("user should have role %s but has role %s instead", expectedRoleName, actualRoleName)
	}

	return nil
}
