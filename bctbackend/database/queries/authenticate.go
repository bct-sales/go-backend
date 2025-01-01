package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
)

func AuthenticateUser(db *sql.DB, userId models.Id, password string) (models.Id, error) {
	row := db.QueryRow(
		`
			SELECT role_id, password
			FROM users
			where user_id = $1
		`,
		userId)

	var roleId models.Id
	var expectedPassword string
	err := row.Scan(&roleId, &expectedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, &UnknownUserError{}
		}

		return 0, err
	}

	if expectedPassword == password {
		return roleId, nil
	} else {
		return 0, &WrongPasswordError{}
	}
}
