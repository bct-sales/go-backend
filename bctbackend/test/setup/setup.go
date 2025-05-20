//go:build test

package setup

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

type DatabaseFixture struct {
	Db *sql.DB
}

func initializeDatabaseFixture(fixture *DatabaseFixture, options ...func(*DatabaseFixture)) {
	fixture.Db = aux.OpenInitializedDatabase()

	for _, option := range options {
		option(fixture)
	}
}

func NewDatabaseFixture(options ...func(*DatabaseFixture)) (DatabaseFixture, *sql.DB) {
	var fixture DatabaseFixture
	initializeDatabaseFixture(&fixture, options...)

	return fixture, fixture.Db
}

func WithDefaultCategories(fixture *DatabaseFixture) {
	fixture.DefaultCategories()
}

func (f *DatabaseFixture) Close() {
	if f.Db != nil {
		f.Db.Close()
	}
	f.Db = nil
}

type RestFixture struct {
	DatabaseFixture
	Router *gin.Engine
	Writer *httptest.ResponseRecorder
}

func initializeRestFixture(fixture *RestFixture, databaseOptions ...func(*DatabaseFixture)) {
	initializeDatabaseFixture(&fixture.DatabaseFixture, databaseOptions...)
	router := aux.CreateRestRouter(fixture.DatabaseFixture.Db)
	fixture.Router = router
	fixture.Writer = httptest.NewRecorder()
}

func NewRestFixture(databaseOptions ...func(*DatabaseFixture)) (RestFixture, *gin.Engine, *httptest.ResponseRecorder) {
	var fixture RestFixture
	initializeRestFixture(&fixture, databaseOptions...)
	return fixture, fixture.Router, fixture.Writer
}

func (f *RestFixture) Close() {
	f.DatabaseFixture.Close()
	f.Router = nil
	f.Writer = nil
}

func (s DatabaseFixture) DefaultCategories() map[models.Id]string {
	table := map[models.Id]string{}

	categoryTable := DefaultCategoryTable()
	for id, name := range categoryTable {
		if err := queries.AddCategory(s.Db, id, name); err != nil {
			panic(err)
		}
	}

	return table
}

func (s DatabaseFixture) User(roleId models.Id, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, roleId, options...)
}

func (s DatabaseFixture) Admin(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.AdminRoleId, options...)
}

func (s DatabaseFixture) Cashier(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.CashierRoleId, options...)
}

func (s DatabaseFixture) Seller(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.SellerRoleId, options...)
}

func (s DatabaseFixture) Session(userId models.Id, options ...func(*aux.AddSessionData)) string {
	return aux.AddSessionToDatabase(s.Db, userId)
}

func (s DatabaseFixture) LoggedIn(user *models.User, options ...func(*aux.AddSessionData)) (*models.User, string) {
	session := aux.AddSessionToDatabase(s.Db, user.UserId, options...)
	return user, session
}

func (s DatabaseFixture) Item(seller models.Id, options ...func(*aux.AddItemData)) *models.Item {
	return aux.AddItemToDatabase(s.Db, seller, options...)
}

func (s DatabaseFixture) Items(seller models.Id, count int, options ...func(*aux.AddItemData)) []*models.Item {
	items := []*models.Item{}

	for i := 0; i < count; i++ {
		updatedOptions := append([]func(*aux.AddItemData){aux.WithDummyData(i)}, options...)
		item := s.Item(seller, updatedOptions...)
		items = append(items, item)
	}

	return items
}

func (s DatabaseFixture) Sale(cashier models.Id, itemIds []models.Id, options ...func(*aux.AddSaleData)) models.Id {
	return aux.AddSaleToDatabase(s.Db, cashier, itemIds, options...)
}

func (s DatabaseFixture) RequireNoSuchUser(t *testing.T, userId models.Id) {
	exists, err := queries.UserWithIdExists(s.Db, userId)
	require.NoError(t, err)
	require.False(t, exists)
}

func (s DatabaseFixture) RequireNoSuchItem(t *testing.T, itemId models.Id) {
	exists, err := queries.ItemWithIdExists(s.Db, itemId)
	require.NoError(t, err)
	require.False(t, exists)
}

func (s DatabaseFixture) RequireFrozen(t *testing.T, saleId ...models.Id) {
	for _, id := range saleId {
		frozen, err := queries.IsItemFrozen(s.Db, id)
		require.NoError(t, err)
		require.True(t, frozen)
	}
}

func (s DatabaseFixture) RequireNotFrozen(t *testing.T, saleId ...models.Id) {
	for _, id := range saleId {
		frozen, err := queries.IsItemFrozen(s.Db, id)
		require.NoError(t, err)
		require.False(t, frozen)
	}
}
