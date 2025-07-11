//go:build test

package setup

import (
	"net/http/httptest"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/server"
	aux "bctbackend/test/helpers"
	"database/sql"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
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
	Server *server.Server
	Writer *httptest.ResponseRecorder
}

func initializeRestFixture(fixture *RestFixture, databaseOptions ...func(*DatabaseFixture)) {
	initializeDatabaseFixture(&fixture.DatabaseFixture, databaseOptions...)
	server := aux.CreateRestServer(fixture.DatabaseFixture.Db)
	fixture.Server = server
	fixture.Writer = httptest.NewRecorder()
}

func NewRestFixture(databaseOptions ...func(*DatabaseFixture)) (RestFixture, *server.Server, *httptest.ResponseRecorder) {
	var fixture RestFixture
	initializeRestFixture(&fixture, databaseOptions...)
	return fixture, fixture.Server, fixture.Writer
}

func (f *RestFixture) Close() {
	f.DatabaseFixture.Close()
	f.Server = nil
	f.Writer = nil
}

func (s DatabaseFixture) Category(id models.Id, name string) {
	if err := queries.AddCategoryWithId(s.Db, id, name); err != nil {
		panic(err)
	}
}

func (s DatabaseFixture) DefaultCategories() {
	categoryNameTable := aux.DefaultCategoryNameTable()
	for id, name := range categoryNameTable {
		s.Category(id, name)
	}
}

func (s DatabaseFixture) User(roleId models.RoleId, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, roleId, options...)
}

func (s DatabaseFixture) Admin(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.NewAdminRoleId(), options...)
}

func (s DatabaseFixture) Cashier(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.NewCashierRoleId(), options...)
}

func (s DatabaseFixture) Seller(options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(s.Db, models.NewSellerRoleId(), options...)
}

func (s DatabaseFixture) Session(userId models.Id, options ...func(*aux.AddSessionData)) models.SessionId {
	return aux.AddSessionToDatabase(s.Db, userId)
}

func (s DatabaseFixture) LoggedIn(user *models.User, options ...func(*aux.AddSessionData)) (*models.User, models.SessionId) {
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

func (s DatabaseFixture) Sale(cashier models.Id, itemIds []models.Id, options ...func(*aux.AddSaleData)) *models.Sale {
	return aux.AddSaleToDatabase(s.Db, cashier, itemIds, options...)
}

func (s DatabaseFixture) RequireNoSuchUsers(t *testing.T, userIds ...models.Id) {
	for _, userId := range userIds {
		exists, err := queries.UserWithIdExists(s.Db, userId)
		require.NoError(t, err)
		require.False(t, exists)
	}
}

func (s DatabaseFixture) RequireNoSuchItems(t *testing.T, itemIds ...models.Id) {
	for _, itemId := range itemIds {
		exists, err := queries.ItemWithIdExists(s.Db, itemId)
		require.NoError(t, err)
		require.False(t, exists)
	}
}

func (s DatabaseFixture) RequireNoSuchSales(t *testing.T, saleIds ...models.Id) {
	for _, itemId := range saleIds {
		exists, err := queries.SaleWithIdExists(s.Db, itemId)
		require.NoError(t, err)
		require.False(t, exists)
	}
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
