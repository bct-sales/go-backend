//go:build test

package setup

import (
	"maps"
	"net/http/httptest"
	"testing"

	"bctbackend/clock"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/logging"
	"bctbackend/server"
	aux "bctbackend/test/helpers"
	"database/sql"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type DatabaseFixture struct {
	Clock *clock.ManualClock
	Db    *sql.DB
}

func initializeDatabaseFixture(t *testing.T, fixture *DatabaseFixture, options ...func(*testing.T, *DatabaseFixture)) {
	fixture.Db = aux.CreateInitializedInMemoryDatabase()
	fixture.Clock = clock.NewManualClock(0)

	for _, option := range options {
		option(t, fixture)
	}
}

func NewDatabaseFixture(t *testing.T, options ...func(*testing.T, *DatabaseFixture)) (DatabaseFixture, *sql.DB) {
	var fixture DatabaseFixture
	initializeDatabaseFixture(t, &fixture, options...)

	return fixture, fixture.Db
}

func WithDefaultCategories(t *testing.T, fixture *DatabaseFixture) {
	fixture.DefaultCategories(t)
}

func (f *DatabaseFixture) Close() {
	if f.Db != nil {
		f.Db.Close()
	}
	f.Db = nil

	f.Clock.StopAllTickers()
	f.Clock = nil
}

type RestFixture struct {
	DatabaseFixture
	Server *server.Server
	Writer *httptest.ResponseRecorder
}

func initializeRestFixture(t *testing.T, fixture *RestFixture, logger logging.Logger, databaseOptions ...func(*testing.T, *DatabaseFixture)) {
	initializeDatabaseFixture(t, &fixture.DatabaseFixture, databaseOptions...)
	server := aux.CreateRestServer(fixture.DatabaseFixture.Db, logger, fixture.Clock)
	fixture.Server = server
	fixture.Writer = httptest.NewRecorder()
}

type TestLogger struct {
	t     *testing.T
	extra map[string]any
}

func (logger *TestLogger) Debug(message string) {
	logger.t.Logf("Debug: %s (%v)", message, logger.extra)
}

func (logger *TestLogger) Info(message string) {
	logger.t.Logf("Info: %s (%v)", message, logger.extra)
}

func (logger *TestLogger) Warn(message string) {
	logger.t.Logf("Warning: %s (%v)", message, logger.extra)
}

func (logger *TestLogger) Error(message string) {
	logger.t.Logf("Error: %s (%v)", message, logger.extra)
}

func (logger *TestLogger) With(key string, value any) logging.Logger {
	extraCopy := maps.Clone(logger.extra)
	extraCopy[key] = value

	return &TestLogger{
		t:     logger.t,
		extra: extraCopy,
	}
}

func createTestLogger(t *testing.T) logging.Logger {
	return &TestLogger{
		t:     t,
		extra: make(map[string]any),
	}
}

func NewRestFixture(t *testing.T, databaseOptions ...func(*testing.T, *DatabaseFixture)) (RestFixture, *server.Server, *httptest.ResponseRecorder) {
	logger := createTestLogger(t)

	var fixture RestFixture
	initializeRestFixture(t, &fixture, logger, databaseOptions...)
	return fixture, fixture.Server, fixture.Writer
}

func (f *RestFixture) NewResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func (f *RestFixture) Close() {
	f.Server.Shutdown()
	f.DatabaseFixture.Close()
	f.Server = nil
	f.Writer = nil
}

func (s DatabaseFixture) Category(t *testing.T, id models.ID, name string) {
	if err := queries.AddCategoryWithID(t.Context(), s.Db, id, name); err != nil {
		panic(err)
	}
}

func (s DatabaseFixture) DefaultCategories(t *testing.T) {
	categoryNameTable := aux.DefaultCategoryNameTable()
	for id, name := range categoryNameTable {
		s.Category(t, id, name)
	}
}

func (s DatabaseFixture) User(t *testing.T, roleID models.RoleID, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(t.Context(), s.Db, roleID, options...)
}

func (s DatabaseFixture) Admin(t *testing.T, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(t.Context(), s.Db, models.NewAdminRoleID(), options...)
}

func (s DatabaseFixture) Cashier(t *testing.T, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(t.Context(), s.Db, models.NewCashierRoleID(), options...)
}

func (s DatabaseFixture) Seller(t *testing.T, options ...func(*aux.AddUserData)) *models.User {
	return aux.AddUserToDatabase(t.Context(), s.Db, models.NewSellerRoleID(), options...)
}

func (s DatabaseFixture) Session(t *testing.T, userID models.ID, options ...func(*aux.AddSessionData)) models.SessionID {
	return aux.AddSessionToDatabase(t.Context(), s.Db, userID, s.Clock.Now(), options...)
}

func (s DatabaseFixture) LoggedIn(t *testing.T, user *models.User, options ...func(*aux.AddSessionData)) (*models.User, models.SessionID) {
	session := aux.AddSessionToDatabase(t.Context(), s.Db, user.UserID, s.Clock.Now(), options...)
	return user, session
}

func (s DatabaseFixture) Item(t *testing.T, seller models.ID, options ...func(*aux.AddItemData)) *models.Item {
	return aux.AddItemToDatabase(t.Context(), s.Db, seller, options...)
}

func (s DatabaseFixture) Items(t *testing.T, seller models.ID, count int, options ...func(*aux.AddItemData)) []*models.Item {
	items := []*models.Item{}

	for i := 0; i < count; i++ {
		updatedOptions := append([]func(*aux.AddItemData){aux.WithDummyData(i)}, options...)
		item := s.Item(t, seller, updatedOptions...)
		items = append(items, item)
	}

	return items
}

func (s DatabaseFixture) Sale(t *testing.T, cashier models.ID, itemIDs []models.ID, options ...func(*aux.AddSaleData)) *models.Sale {
	transaction, err := queries.NewTransactionalDatabaseQuerier(t.Context(), s.Db)
	if err != nil {
		panic(err)
	}

	sale := aux.AddSaleToDatabase(t.Context(), transaction, cashier, itemIDs, options...)

	if err := transaction.Commit(); err != nil {
		panic(err)
	}

	return sale
}

func (s DatabaseFixture) RequireNoSuchUsers(t *testing.T, userIDs ...models.ID) {
	for _, userID := range userIDs {
		exists, err := queries.UserWithIDExists(t.Context(), s.Db, userID)
		require.NoError(t, err)
		require.False(t, exists)
	}
}

func (s DatabaseFixture) GenerateNonexistentID(t *testing.T, existenceChecker func(id models.ID) (bool, error)) models.ID {
	id := models.ID(9999)
	attemptsLeft := 5

	for {
		exists, err := existenceChecker(id)
		require.NoError(t, err)

		if !exists {
			return id
		}

		id += 23
		attemptsLeft--
		if attemptsLeft == 0 {
			require.Fail(t, "Failed to generate nonexistent id")
		}
	}
}

func (s DatabaseFixture) GenerateNonexistentUserID(t *testing.T) models.ID {
	return s.GenerateNonexistentID(t, func(id models.ID) (bool, error) { return queries.UserWithIDExists(t.Context(), s.Db, id) })
}

func (s DatabaseFixture) RequireNoSuchItems(t *testing.T, itemIDs ...models.ID) {
	for _, itemID := range itemIDs {
		exists, err := queries.ItemWithIDExists(t.Context(), s.Db, itemID)
		require.NoError(t, err)
		require.False(t, exists)
	}
}

func (s DatabaseFixture) GenerateNonexistentItemID(t *testing.T) models.ID {
	return s.GenerateNonexistentID(t, func(id models.ID) (bool, error) { return queries.ItemWithIDExists(t.Context(), s.Db, id) })
}

func (s DatabaseFixture) RequireNoSuchSales(t *testing.T, saleIDs ...models.ID) {
	for _, itemID := range saleIDs {
		exists, err := queries.SaleWithIDExists(t.Context(), s.Db, itemID)
		require.NoError(t, err)
		require.False(t, exists)
	}
}

func (s DatabaseFixture) GenerateNonexistentSaleID(t *testing.T) models.ID {
	return s.GenerateNonexistentID(t, func(id models.ID) (bool, error) { return queries.SaleWithIDExists(t.Context(), s.Db, id) })
}

func (s DatabaseFixture) RequireFrozen(t *testing.T, saleID ...models.ID) {
	for _, id := range saleID {
		frozen, err := queries.IsItemFrozen(t.Context(), s.Db, id)
		require.NoError(t, err)
		require.True(t, frozen)
	}
}

func (s DatabaseFixture) RequireNotFrozen(t *testing.T, saleID ...models.ID) {
	for _, id := range saleID {
		frozen, err := queries.IsItemFrozen(t.Context(), s.Db, id)
		require.NoError(t, err)
		require.False(t, frozen)
	}
}

func (s DatabaseFixture) WithTransaction(t *testing.T, f func(transaction *queries.TransactionalDatabaseQuerier)) {
	transaction, transactionErr := queries.NewTransactionalDatabaseQuerier(t.Context(), s.Db)
	require.NoError(t, transactionErr)

	f(transaction)

	commitErr := transaction.Commit()
	require.NoError(t, commitErr)
}

func (s DatabaseFixture) WithTransactionErr(t *testing.T, f func(transaction *queries.TransactionalDatabaseQuerier) error) error {
	transaction, transactionErr := queries.NewTransactionalDatabaseQuerier(t.Context(), s.Db)
	require.NoError(t, transactionErr)
	defer transaction.RollbackIfNotCommitted()

	if err := f(transaction); err != nil {
		return err
	}

	commitErr := transaction.Commit()
	require.NoError(t, commitErr)

	return nil
}
