//go:build test

package cli_tests

import (
	"database/sql"
	"regexp"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"

	"github.com/stretchr/testify/require"
)

// setupItemTest creates a workspace with an initialized database
// and a seller with ID 100, returning the workspace, a run helper
// and the database path.
func setupItemTest(t *testing.T) (*Workspace, func(...string) CommandResult, string) {
	t.Helper()

	workspace := NewWorkspace(t)
	databasePath := workspace.DatabasePath()

	run := func(arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	requireSuccess(t, run("db", "init", "--db", databasePath))
	requireSuccess(t, run("user", "add", "--id", "100", "--role", "seller", "--password", "pw100", "--db", databasePath))

	return workspace, run, databasePath
}

// addItem adds an item to the database and returns its ID as a string.
func addItem(t *testing.T, run func(...string) CommandResult, databasePath string, description string) string {
	t.Helper()

	result := run("item", "add", "--description", description, "--price", "1000", "--category", "1", "--seller", "100", "--db", databasePath)
	requireSuccess(t, result)

	match := regexp.MustCompile(`Item (\d+) added successfully`).FindStringSubmatch(result.Stdout)
	require.NotNil(t, match, "no item ID in output: %s", result.Stdout)

	return match[1]
}

func TestItemCopyArgumentValidation(t *testing.T) {
	workspace := NewWorkspace(t)

	run := func(arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	t.Run("Copy with no arguments fails cleanly", func(t *testing.T) {
		result := run("item", "copy")

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "accepts 1 arg(s), received 0")
	})

	t.Run("Copy with two arguments fails cleanly", func(t *testing.T) {
		result := run("item", "copy", "1001", "1002")

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "accepts 1 arg(s), received 2")
	})
}

// seedItemSpec describes the flags of an item to seed directly into the database.
type seedItemSpec struct {
	frozen      bool
	hidden      bool
	categoryID  models.ID
	description string
}

// seedItems seeds a seller with ID 100 and items with the given specs
// directly into the database, returning the IDs of the seeded items.
func seedItems(t *testing.T, databasePath string, specs ...seedItemSpec) []models.ID {
	return seedItemsForSeller(t, databasePath, models.ID(100), specs...)
}

// seedItemsForSeller seeds a seller with the given ID and items with the
// given specs directly into the database, returning the IDs of the seeded items.
func seedItemsForSeller(t *testing.T, databasePath string, sellerID models.ID, specs ...seedItemSpec) []models.ID {
	t.Helper()

	var itemIDs []models.ID

	SeedDatabase(t, databasePath, func(db *sql.DB) {
		require.NoError(t, queries.AddUserWithID(t.Context(), db, sellerID, models.NewSellerRoleID(), models.Now(), nil, "pw"+sellerID.String()))

		for _, spec := range specs {
			categoryID := spec.categoryID
			if categoryID == 0 {
				categoryID = models.ID(1)
			}

			description := spec.description
			if description == "" {
				description = "red t-shirt"
			}

			itemID, err := queries.AddItem(
				t.Context(),
				db,
				models.Now(),
				description,
				models.MoneyInCents(1000),
				categoryID,
				sellerID,
				false,
				false,
				spec.frozen,
				spec.hidden,
				false,
			)
			require.NoError(t, err)
			itemIDs = append(itemIDs, itemID)
		}
	})

	return itemIDs
}

// seedUser seeds a user with the given ID and role directly into the database.
func seedUser(t *testing.T, databasePath string, userID models.ID, roleID models.RoleID) {
	t.Helper()

	SeedDatabase(t, databasePath, func(db *sql.DB) {
		require.NoError(t, queries.AddUserWithID(t.Context(), db, userID, roleID, models.Now(), nil, "pw"+userID.String()))
	})
}

// setupDatabaseTest creates a workspace with an initialized database,
// returning a run helper and the database path.
func setupDatabaseTest(t *testing.T) (func(...string) CommandResult, string) {
	t.Helper()

	workspace := NewWorkspace(t)
	databasePath := workspace.DatabasePath()

	run := func(arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	requireSuccess(t, run("db", "init", "--db", databasePath))

	return run, databasePath
}

func TestItemLifecycle(t *testing.T) {
	t.Run("Freeze marks items as frozen", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "freeze", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Items frozen successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.True(t, item.Frozen)
		})
	})

	t.Run("Freeze marks multiple items as frozen", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{}, seedItemSpec{})

		result := run("item", "freeze", itemIDs[0].String(), itemIDs[1].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Items frozen successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			for _, itemID := range itemIDs {
				item, err := queries.GetItemWithID(t.Context(), db, itemID)
				require.NoError(t, err)
				require.True(t, item.Frozen)
			}
		})
	})

	t.Run("Freeze with no arguments fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("item", "freeze", "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Freeze with an invalid ID fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "freeze", "abc", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Invalid item ID: abc")
	})

	t.Run("Unfreeze clears the frozen status", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{frozen: true})

		result := run("item", "unfreeze", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Items unfrozen successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.False(t, item.Frozen)
		})
	})

	t.Run("Hide and unhide toggle the hidden status", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "hide", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Items hidden successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.True(t, item.Hidden)
		})

		result = run("item", "unhide", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Items unhidden successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.False(t, item.Hidden)
		})
	})

	t.Run("Remove deletes the item", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "remove", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Removed item "+itemIDs[0].String())
		require.Contains(t, result.Stdout, "All specified items have been successfully removed")

		result = run("item", "show", itemIDs[0].String(), "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Remove of a nonexistent item fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "remove", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Failed to remove item")
		require.Contains(t, result.Stderr, "No items have been removed")
	})

	t.Run("Copy duplicates the item", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "copy", itemIDs[0].String(), "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Item ID")
		require.Contains(t, result.Stdout, "red t-shirt")

		result = run("item", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of items listed: 2")
	})

	t.Run("Update changes the item", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "update", "--id", itemIDs[0].String(), "--description", "blue jeans", "--price", "2000", "--category", "10", "--large", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Item updated successfully")
		require.Contains(t, result.Stdout, "blue jeans")
		require.Contains(t, result.Stdout, "20.00")
		require.Contains(t, result.Stdout, "Books")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.Equal(t, "blue jeans", item.Description)
			require.Equal(t, models.MoneyInCents(2000), item.PriceInCents)
		})
	})

	t.Run("Update rejects conflicting flag pairs", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "update", "--id", itemIDs[0].String(), "--donation", "--no-donation", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "were all set")
	})

	t.Run("Update of a nonexistent item fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("item", "update", "--id", "999", "--description", "blue jeans", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Failed to update item")
	})
}

func TestItemCommands(t *testing.T) {
	t.Run("Add prints item table", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add", "--description", "red t-shirt", "--price", "1000", "--category", "1", "--seller", "100", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "added successfully")
		require.Contains(t, result.Stdout, "red t-shirt")
		require.Contains(t, result.Stdout, "10.00")
		require.Contains(t, result.Stdout, "Clothing 0-3 mos (50-56)")
	})

	t.Run("Add rejects price zero", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add", "--description", "red t-shirt", "--price", "0", "--category", "1", "--seller", "100", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Invalid price: 0 cents")
	})

	t.Run("Add rejects unknown category", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add", "--description", "red t-shirt", "--price", "1000", "--category", "999", "--seller", "100", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "No such category with ID 999")
	})

	t.Run("Add rejects unknown seller", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add", "--description", "red t-shirt", "--price", "1000", "--category", "1", "--seller", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "No user with ID 999")
	})

	t.Run("Add rejects a user that is not a seller", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("user", "add", "--id", "1", "--role", "admin", "--password", "pw1", "--db", databasePath)
		requireSuccess(t, result)

		result = run("item", "add", "--description", "red t-shirt", "--price", "1000", "--category", "1", "--seller", "1", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "is not a seller")
	})

	t.Run("Add requires all flags", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add", "--description", "red t-shirt", "--category", "1", "--seller", "100", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "required flag(s)")
	})

	t.Run("List on empty database", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No items found")
	})

	t.Run("List shows added items", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		addItem(t, run, databasePath, "red t-shirt")

		result := run("item", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "red t-shirt")
		require.Contains(t, result.Stdout, "Number of items listed: 1")
	})

	t.Run("List as CSV", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		addItem(t, run, databasePath, "red t-shirt")

		result := run("item", "list", "--format", "csv", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "item_id,seller_id,description,category,price_in_cents,donation,charity,large,sale_count")
		require.Contains(t, result.Stdout, "red t-shirt")
	})

	t.Run("Hidden items are excluded unless show-hidden is set", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		itemID := addItem(t, run, databasePath, "red t-shirt")

		result := run("item", "hide", itemID, "--db", databasePath)
		requireSuccess(t, result)

		result = run("item", "list", "--db", databasePath)
		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No items found")

		result = run("item", "list", "--show-hidden", "--db", databasePath)
		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "red t-shirt")
	})

	t.Run("Show displays item details", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		itemID := addItem(t, run, databasePath, "red t-shirt")

		result := run("item", "show", itemID, "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "red t-shirt")
		require.Contains(t, result.Stdout, "10.00")
	})

	t.Run("Show with an invalid ID fails", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "show", "abc", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Invalid item ID: abc")
	})

	t.Run("Show of a nonexistent item fails", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "show", "999", "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Add-consumable creates five items", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add-consumable", "--description", "sticker", "--price", "100", "--category", "1", "--seller", "100", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Quantity 1 -> ID")
		require.Contains(t, result.Stdout, "Quantity 5 -> ID")

		result = run("item", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of items listed: 5")
		require.Contains(t, result.Stdout, "sticker (3)")
	})

	t.Run("Add-consumable failure adds nothing", func(t *testing.T) {
		_, run, databasePath := setupItemTest(t)

		result := run("item", "add-consumable", "--description", "sticker", "--price", "100", "--category", "999", "--seller", "100", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "No such category with ID 999")

		result = run("item", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No items found")
	})
}
