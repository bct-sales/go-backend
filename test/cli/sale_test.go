//go:build test

package cli_tests

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"

	"github.com/stretchr/testify/require"
)

// setupSaleTest creates a workspace with an initialized database, a cashier
// with ID 1, a seller with ID 100 and two items, returning a run helper, the
// database path and the IDs of the two items.
func setupSaleTest(t *testing.T) (func(...string) CommandResult, string, []string) {
	t.Helper()

	run, databasePath := setupDatabaseTest(t)

	requireSuccess(t, run("user", "add", "--id", "1", "--role", "cashier", "--password", "pw1", "--db", databasePath))
	requireSuccess(t, run("user", "add", "--id", "100", "--role", "seller", "--password", "pw100", "--db", databasePath))

	itemIDs := []string{
		addItem(t, run, databasePath, "red t-shirt"),
		addItem(t, run, databasePath, "blue jeans"),
	}

	return run, databasePath, itemIDs
}

// collectSaleIDs returns all sale IDs in the database.
// Note: the query holds the single database connection while iterating,
// so no other queries may be nested inside the receiver callback.
func collectSaleIDs(t *testing.T, db *sql.DB) []models.ID {
	t.Helper()

	var saleIDs []models.ID

	err := queries.NewGetSalesQuery().Execute(t.Context(), db, func(sale *models.SaleSummary) error {
		saleIDs = append(saleIDs, sale.SaleID)
		return nil
	})
	require.NoError(t, err)

	return saleIDs
}

// addSale adds a sale with the given item IDs and returns the ID of the sale,
// read back from the database.
func addSale(t *testing.T, run func(...string) CommandResult, databasePath string, itemIDs ...string) models.ID {
	t.Helper()

	result := run("sale", "add", "--cashier", "1", "--items", strings.Join(itemIDs, ","), "--db", databasePath)
	requireSuccess(t, result)

	var saleID models.ID

	VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
		saleIDs := collectSaleIDs(t, db)
		require.Len(t, saleIDs, 1)
		saleID = saleIDs[0]
	})

	require.NotZero(t, saleID)

	return saleID
}

func TestSaleCommands(t *testing.T) {
	t.Run("Add creates a sale", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)

		result := run("sale", "add", "--cashier", "1", "--items", strings.Join(itemIDs, ","), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Sale added successfully")
		require.Contains(t, result.Stdout, "Item 1")
		require.Contains(t, result.Stdout, "Item 2")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			saleIDs := collectSaleIDs(t, db)
			require.Len(t, saleIDs, 1)

			saleItems, err := queries.GetSaleItems(t.Context(), db, saleIDs[0])
			require.NoError(t, err)
			require.Len(t, saleItems, 2)
		})
	})

	t.Run("Add with an unknown cashier fails", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)

		result := run("sale", "add", "--cashier", "999", "--items", strings.Join(itemIDs, ","), "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "failed to add sale")
	})

	t.Run("Add with duplicate items fails", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)

		result := run("sale", "add", "--cashier", "1", "--items", itemIDs[0]+","+itemIDs[0], "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "failed to add sale")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			require.Empty(t, collectSaleIDs(t, db))
		})
	})

	t.Run("Add with an unknown item fails", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "add", "--cashier", "1", "--items", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "failed to add sale")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			require.Empty(t, collectSaleIDs(t, db))
		})
	})

	t.Run("Add requires the items flag", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "add", "--cashier", "1", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "required flag(s)")
	})

	t.Run("List on empty database", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No sales found")
	})

	t.Run("List shows sales", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of sales listed: 1")
		require.Contains(t, result.Stdout, "20.00")
	})

	t.Run("Show displays sale details", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		saleID := addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "show", saleID.String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Total Cost")
		require.Contains(t, result.Stdout, "20.00")
		require.Contains(t, result.Stdout, "Number of Items")
		require.Contains(t, result.Stdout, "red t-shirt")
	})

	t.Run("Show of a nonexistent sale fails", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "show", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "No sale found with ID 999")
	})

	t.Run("Show with an invalid ID fails", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "show", "abc", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Invalid sale ID")
	})

	t.Run("Remove deletes the sale", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		saleID := addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "remove", saleID.String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, fmt.Sprintf("Removing sale with id %s", saleID.String()))

		result = run("sale", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No sales found")
	})

	t.Run("Remove of a nonexistent sale rolls back", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "remove", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "NONE of the sales was actually removed")
	})

	t.Run("Remove-all requires force", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "remove-all", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Please use --force")

		result = run("sale", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of sales listed: 1")

		result = run("sale", "remove-all", "--force", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "All sales removed successfully.")

		result = run("sale", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No sales found")
	})

	t.Run("Items on empty database", func(t *testing.T) {
		run, databasePath, _ := setupSaleTest(t)

		result := run("sale", "items", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No sold items found")
	})

	t.Run("Items lists sold items", func(t *testing.T) {
		run, databasePath, itemIDs := setupSaleTest(t)
		addSale(t, run, databasePath, itemIDs...)

		result := run("sale", "items", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "red t-shirt")
		require.Contains(t, result.Stdout, "blue jeans")
		require.Contains(t, result.Stdout, "Number of sales listed: 2")

		result = run("sale", "items", "--format", "csv", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "red t-shirt")
	})
}
