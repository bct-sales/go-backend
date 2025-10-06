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

func WithUserID(userID models.ID) func(*AddUserData) {
	return func(data *AddUserData) {
		data.UserID = &userID
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

func AddUserToDatabase(db *sql.DB, roleID models.RoleID, options ...func(*AddUserData)) *models.User {
	data := AddUserData{
		RoleID: roleID,
	}

	for _, option := range options {
		option(&data)
	}

	data.FillWithDefaults()

	var userID models.ID
	if data.UserID == nil {
		var err error
		userID, err = queries.AddUser(db, roleID, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	} else {
		userID = *data.UserID
		var err error
		err = queries.AddUserWithID(db, userID, roleID, *data.CreatedAt, data.LastActivity, *data.Password)

		if err != nil {
			panic(err)
		}
	}

	user, err := queries.GetUserWithID(db, userID)
	if err != nil {
		panic(err)
	}

	return user
}
