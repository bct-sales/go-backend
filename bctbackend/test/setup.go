//go:build test

package rest

import (
	"net/http/httptest"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type SetupObject struct {
	Db     *sql.DB
	Router *gin.Engine
	Writer *httptest.ResponseRecorder
}

func Setup() (*SetupObject, *sql.DB) {
	db := aux.OpenInitializedDatabase()

	setup := SetupObject{
		Db: db,
	}

	return &setup, db
}

func SetupRestTest() (*SetupObject, *gin.Engine, *httptest.ResponseRecorder) {
	db, router := aux.CreateRestRouter()
	writer := httptest.NewRecorder()

	setup := SetupObject{
		Db:     db,
		Router: router,
		Writer: writer,
	}

	return &setup, router, writer
}

func (s *SetupObject) Close() {
	if s.Db != nil {
		s.Db.Close()
	}

	s.Db = nil
	s.Router = nil
}

func (s *SetupObject) User(roleId models.Id, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, roleId, options...)
}

func (s *SetupObject) Admin(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.AdminRoleId, options...)
}

func (s *SetupObject) Cashier(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.CashierRoleId, options...)
}

func (s *SetupObject) Seller(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.SellerRoleId, options...)
}

func (s *SetupObject) Session(userId models.Id, options ...func(*aux.AddSessionData)) string {
	return aux.AddSessionToDatabase(s.Db, userId)
}

func (s *SetupObject) LoggedIn(user *models.User, options ...func(*aux.AddSessionData)) (*models.User, string) {
	session := aux.AddSessionToDatabase(s.Db, user.UserId, options...)
	return user, session
}

func (s *SetupObject) Item(seller models.Id, options ...func(*aux.AddItemData)) *models.Item {
	return aux.AddItemToDatabase(s.Db, seller, options...)
}

func (s *SetupObject) Sale(cashier models.Id, itemIds []models.Id, options ...func(*aux.AddSaleData)) models.Id {
	return aux.AddSaleToDatabase(s.Db, cashier, itemIds, options...)
}

func (s *SetupObject) RequireNoSuchUser(t *testing.T, userId models.Id) {
	exists, err := queries.UserWithIdExists(s.Db, userId)
	require.NoError(t, err)
	require.False(t, exists)
}

func (s *SetupObject) RequireNoSuchItem(t *testing.T, itemId models.Id) {
	exists, err := queries.ItemWithIdExists(s.Db, itemId)
	require.NoError(t, err)
	require.False(t, exists)
}
