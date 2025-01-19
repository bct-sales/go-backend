//go:build test

package setup

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"bctbackend/rest"
	"database/sql"

	gin "github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func CreateRestRouter() (*sql.DB, *gin.Engine) {
	db := OpenInitializedDatabase()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rest.DefineEndpoints(db, router)

	return db, router
}

type AddUserData struct {
	UserId       *models.Id
	RoleId       models.Id
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
		createdAt := models.NewTimestamp(0)
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

func AddUserToDatabase(db *sql.DB, roleId models.Id, options ...func(*AddUserData)) models.User {
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

func AddSellerToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.SellerRoleId, options...)
}

func AddCashierToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.CashierRoleId, options...)
}

func AddAdminToDatabase(db *sql.DB, options ...func(*AddUserData)) models.User {
	return AddUserToDatabase(db, models.AdminRoleId, options...)
}
