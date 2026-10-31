//go:build test

package cli_tests

import (
	"database/sql"
	"testing"

	"bctbackend/database/models"
	"bctbackend/database/queries"

	"github.com/stretchr/testify/require"
)

func TestUserCommands(t *testing.T) {
	workspace := NewWorkspace(t)

	databasePath := workspace.DatabasePath()

	run := func(arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	// The database is created here; the behavior of `db init` itself is
	// covered by TestDatabaseCommands.
	result := run("db", "init", "--db", databasePath)
	requireSuccess(t, result)

	t.Run("User add", func(t *testing.T) {
		result := run("user", "add", "--id", "1", "--role", "admin", "--password", "secret123", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "User added successfully")
	})

	t.Run("User list", func(t *testing.T) {
		result := run("user", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "admin")
		require.Contains(t, result.Stdout, "secret123")
		require.Contains(t, result.Stdout, "Number of users listed: 1")
	})

	t.Run("User add with duplicate id fails", func(t *testing.T) {
		result := run("user", "add", "--id", "1", "--role", "admin", "--password", "secret123", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Failed to add user")
	})

	t.Run("User add without password fails", func(t *testing.T) {
		result := run("user", "add", "--id", "2", "--role", "admin", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "required")
	})
}

func TestUserLifecycle(t *testing.T) {
	t.Run("Show displays admin information", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(1), models.NewAdminRoleID())

		result := run("user", "show", "1", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "User Data")
		require.Contains(t, result.Stdout, "admin")
		require.Contains(t, result.Stdout, "N/A")
	})

	t.Run("Show displays a seller with items", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("user", "show", "100", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "User Data")
		require.Contains(t, result.Stdout, "Items")
		require.Contains(t, result.Stdout, "red t-shirt")
	})

	t.Run("Show of a nonexistent user fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "show", "999", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "User with ID 999 does not exist.")
	})

	t.Run("Show with an invalid ID fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "show", "abc", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Invalid user ID: abc")
	})

	t.Run("Set-password invalidates sessions", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(1), models.NewAdminRoleID())

		result := run("session", "add", "-u", "1", "-e", "1d", "--db", databasePath)
		requireSuccess(t, result)

		result = run("user", "set-password", "1", "newpass", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Password updated")
		require.Contains(t, result.Stdout, "Invalidated sessions associated with user")

		result = run("session", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "no sessions found")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			user, err := queries.GetUserWithID(t.Context(), db, models.ID(1))
			require.NoError(t, err)
			require.Equal(t, "newpass", user.Password)
		})
	})

	t.Run("Set-password with no-invalidate keeps sessions", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(1), models.NewAdminRoleID())

		result := run("session", "add", "-u", "1", "-e", "1d", "--db", databasePath)
		requireSuccess(t, result)

		result = run("user", "set-password", "1", "newpass", "--no-invalidate", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Password updated")
		require.NotContains(t, result.Stdout, "Invalidated")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			sessions, err := queries.GetSessions(t.Context(), db)
			require.NoError(t, err)
			require.Len(t, sessions, 1)
		})
	})

	t.Run("Remove deletes the user", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(5), models.NewAdminRoleID())

		result := run("user", "remove", "5", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "User removed successfully")

		result = run("user", "show", "5", "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Remove of a user with items fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("user", "remove", "100", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Failed to remove user")
	})

	t.Run("Add-sellers tops up zones", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "add-sellers", "--zones", "1,2", "--per-zone", "3", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Added the following sellers:")
		for _, sellerID := range []string{"100", "101", "102", "200", "201", "202"} {
			require.Contains(t, result.Stdout, sellerID)
		}

		result = run("user", "add-sellers", "--zones", "1,2", "--per-zone", "3", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "No sellers added, all specified zones already have enough sellers.")
	})

	t.Run("Add-sellers with invalid zones fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "add-sellers", "--zones", "abc", "--per-zone", "3", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "failed to parse zones")
	})

	t.Run("Add with generate-password assigns a password", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(1), models.NewAdminRoleID())

		result := run("user", "add", "--id", "7", "--role", "seller", "--generate-password", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "User added successfully")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			user, err := queries.GetUserWithID(t.Context(), db, models.ID(7))
			require.NoError(t, err)
			require.NotEmpty(t, user.Password)
		})
	})

	t.Run("Add with generate-password on an empty database fails", func(t *testing.T) {
		// This documents current behavior, which is arguably a bug:
		// FindUnusedPassword only considers as many dictionary entries as there
		// are used passwords, so on a database without any users the command
		// fails with "out of passwords" instead of assigning a password.
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "add", "--id", "7", "--role", "seller", "--generate-password", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "out of passwords")
	})

	t.Run("Add rejects password and generate-password together", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("user", "add", "--id", "7", "--role", "seller", "--password", "pw7", "--generate-password", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "were all set")
	})

	t.Run("Move-items moves items to the new seller", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		itemIDs := seedItemsForSeller(t, databasePath, models.ID(100), seedItemSpec{})
		seedItemsForSeller(t, databasePath, models.ID(200))

		result := run("user", "move-items", "--from", "100", "--to", "200", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Successfully moved items")

		VerifyDatabase(t, databasePath, func(t *testing.T, db *sql.DB) {
			item, err := queries.GetItemWithID(t.Context(), db, itemIDs[0])
			require.NoError(t, err)
			require.Equal(t, models.ID(200), item.SellerID)
		})
	})

	t.Run("Move-items with equal sellers is a no-op", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})

		result := run("user", "move-items", "--from", "100", "--to", "100", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Warning: this is a no-op because from and to are equal")
	})

	t.Run("Move-items refuses to merge without force-merge", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})
		seedItemsForSeller(t, databasePath, models.ID(200), seedItemSpec{})

		result := run("user", "move-items", "--from", "100", "--to", "200", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "already has items")
		require.Contains(t, result.Stderr, "Use --force-merge")

		result = run("user", "move-items", "--from", "100", "--to", "200", "--force-merge", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Successfully moved items")
	})

	t.Run("Move-items refuses to move frozen items without force-frozen", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{frozen: true})
		seedItemsForSeller(t, databasePath, models.ID(200))

		result := run("user", "move-items", "--from", "100", "--to", "200", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Use --force-frozen")

		result = run("user", "move-items", "--from", "100", "--to", "200", "--force-frozen", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Successfully moved items")
	})

	t.Run("Move-items with a nonexistent seller fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath, seedItemSpec{})
		seedItemsForSeller(t, databasePath, models.ID(200))

		result := run("user", "move-items", "--from", "999", "--to", "200", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "does not exist")
	})

	t.Run("Move-items with a non-seller fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedUser(t, databasePath, models.ID(1), models.NewAdminRoleID())
		seedItemsForSeller(t, databasePath, models.ID(200))

		result := run("user", "move-items", "--from", "1", "--to", "200", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "is not a seller")
	})

	t.Run("Move-items with no items to donate is a no-op", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItemsForSeller(t, databasePath, models.ID(100))
		seedItemsForSeller(t, databasePath, models.ID(200))

		result := run("user", "move-items", "--from", "100", "--to", "200", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "no items to donate")
	})
}
