package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
)

func AddUserWithId(
	db *sql.DB,
	userId models.Id,
	roleId models.Id,
	timestamp models.Timestamp,
	password string) error {

	_, err := db.Exec(
		`
			INSERT INTO users (user_id, role_id, timestamp, password)
			VALUES ($1, $2, $3, $4)
		`,
		userId,
		roleId,
		timestamp,
		password,
	)

	return err
}

func AddUser(
	db *sql.DB,
	roleId models.Id,
	timestamp models.Timestamp,
	password string) (models.Id, error) {

	result, err := db.Exec(
		`
			INSERT INTO users (role_id, timestamp, password)
			VALUES ($1, $2, $3)
		`,
		roleId,
		timestamp,
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

func AuthenticateUser(db *sql.DB, userId models.Id, password string) error {
	row := db.QueryRow(
		`
			SELECT password
			FROM users
			where user_id = $1
		`,
		userId)

	var expectedPassword string
	err := row.Scan(&expectedPassword)

	if err != nil {
		return err
	}

	if expectedPassword == password {
		return nil
	} else {
		return errors.New("invalid password")
	}
}

func GetUserWithId(db *sql.DB, userId models.Id) (models.User, error) {
	row := db.QueryRow(
		`
			SELECT role_id, timestamp, password
			FROM users
			WHERE user_id = $1
		`,
		userId,
	)

	var roleId models.Id
	var timestamp models.Timestamp
	var password string
	err := row.Scan(&roleId, &timestamp, &password)

	return models.User{
		UserId:    userId,
		RoleId:    roleId,
		Timestamp: timestamp,
		Password:  password,
	}, err
}
