//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"
)

type AddUserData struct {
	UserId       *models.Id
	RoleId       models.RoleId
	Password     *string
	CreatedAt    *models.Timestamp
	LastActivity *models.Timestamp
}

func (data *AddUserData) FillWithDefaults() {
	if data.Password == nil {
		password := "test"
		data.Password = &password
	}

	if data.CreatedAt == nil {
		createdAt := models.Timestamp(0)
		data.CreatedAt = &createdAt
	}
}

func WithUserId(userId models.Id) func(*AddUserData) {
	return func(data *AddUserData) {
		data.UserId = &userId
	}
}

func WithPassword(password string) func(*AddUserData) {
	return func(data *AddUserData) {
		data.Password = &password
	}
}

func WithCreatedAt(createdAt models.Timestamp) func(*AddUserData) {
	return func(data *AddUserData) {
		data.CreatedAt = &createdAt
	}
}

func WithLastActivity(lastActivity models.Timestamp) func(*AddUserData) {
	return func(data *AddUserData) {
		data.LastActivity = &lastActivity
	}
}

func AddUserToDatabase(db *sql.DB, roleId models.RoleId, options ...func(*AddUserData)) *models.User {
	data := AddUserData{
		RoleId: roleId,
	}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	var userId models.Id
	if data.UserId == nil {
		var err error
		userId, err = queries.AddUser(db, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	} else {
		userId = *data.UserId
		var err error
		err = queries.AddUserWithId(db, userId, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	}

	user, err := queries.GetUserWithId(db, userId)
	if err != nil {
		panic(err)
	}

	return user
}
