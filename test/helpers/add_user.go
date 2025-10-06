//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"
)

type AddUserData struct {
	UserID       *models.ID
	RoleID       models.RoleID
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

func WithUserID(userId models.ID) func(*AddUserData) {
	return func(data *AddUserData) {
		data.UserID = &userId
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

func WithNoLastActivity() func(*AddUserData) {
	return func(data *AddUserData) {
		data.LastActivity = nil
	}
}

func WithLastActivity(lastActivity models.Timestamp) func(*AddUserData) {
	return func(data *AddUserData) {
		data.LastActivity = &lastActivity
	}
}

func AddUserToDatabase(db *sql.DB, roleId models.RoleID, options ...func(*AddUserData)) *models.User {
	data := AddUserData{
		RoleID: roleId,
	}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	var userId models.ID
	if data.UserID == nil {
		var err error
		userId, err = queries.AddUser(db, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	} else {
		userId = *data.UserID
		var err error
		err = queries.AddUserWithID(db, userId, roleId, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	}

	user, err := queries.GetUserWithID(db, userId)
	if err != nil {
		panic(err)
	}

	return user
}
