package queries

import (
	models "bctrest/db/models"
	"bctrest/security"
	"database/sql"
	"errors"
)

func AddUser(
	db *sql.DB,
	userId models.Id,
	roleId models.Id,
	timestamp models.Timestamp,
	passwordHash string,
	passwordSalt string) error {

	_, err := db.Exec(
		`
			INSERT INTO users (user_id, role_id, timestamp, password_hash, password_salt)
			VALUES ($1, $2, $3, $4, $5)
		`,
		userId,
		roleId,
		timestamp,
		passwordHash,
		passwordSalt,
	)

	return err
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
			SELECT password_hash, password_salt
			FROM users
			where user_id = $1
		`,
		userId)

	var expectedHash string
	var salt string
	err := row.Scan(&expectedHash, &salt)

	if err != nil {
		return err
	}

	actualHash := security.HashPassword(password, salt)

	if expectedHash == actualHash {
		return nil
	} else {
		return errors.New("invalid password")
	}
}
