//go:build test

package cli_tests

import (
	"strings"
	"testing"

	"bctbackend/database/models"

	"github.com/stretchr/testify/require"
)

// requireTableRow asserts that the stdout of a table command contains a row
// whose text includes cellText, and that the same row includes expectedSubstring.
// Table commands must be run with --no-color for this to be reliable.
func requireTableRow(t *testing.T, stdout string, cellText string, expectedSubstring string) {
	t.Helper()

	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, cellText) {
			require.Contains(t, line, expectedSubstring)
			return
		}
	}

	require.Fail(t, "no table row contains "+cellText, "stdout:\n%s", stdout)
}

func TestCategoryCommands(t *testing.T) {
	t.Run("List shows the default categories", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("category", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Clothing 0-3 mos (50-56)")
		require.Contains(t, result.Stdout, "Books")
	})

	t.Run("Add creates a category", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("category", "add", "--id", "50", "--name", "Vinyl records", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Category successfully added")

		result = run("category", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Vinyl records")
	})

	t.Run("Add with a duplicate ID fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("category", "add", "--id", "1", "--name", "Duplicate", "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Rename changes the category name", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("category", "add", "--id", "50", "--name", "Vinyl records", "--db", databasePath)
		requireSuccess(t, result)

		result = run("category", "rename", "--id", "50", "--name", "45 rpm", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Category successfully renamed")

		result = run("category", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "45 rpm")
		require.NotContains(t, result.Stdout, "Vinyl records")
	})

	t.Run("Rename of a nonexistent category fails", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)

		result := run("category", "rename", "--id", "999", "--name", "X", "--db", databasePath)

		requireFailure(t, result)
	})

	t.Run("Count counts visible items per category", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath,
			seedItemSpec{categoryID: models.ID(1)},
			seedItemSpec{categoryID: models.ID(1)},
			seedItemSpec{categoryID: models.ID(10)},
			seedItemSpec{categoryID: models.ID(11), hidden: true},
		)

		result := run("category", "count", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		requireTableRow(t, result.Stdout, "Clothing 0-3 mos (50-56)", "2")
		requireTableRow(t, result.Stdout, "Books", "1")
		// All categories are listed, even with zero items; the hidden item is not counted.
		requireTableRow(t, result.Stdout, "Toys", "0")
	})

	t.Run("Count with include-hidden counts hidden items", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		seedItems(t, databasePath,
			seedItemSpec{categoryID: models.ID(1)},
			seedItemSpec{categoryID: models.ID(10)},
			seedItemSpec{categoryID: models.ID(11), hidden: true},
		)

		result := run("category", "count", "--include-hidden", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		requireTableRow(t, result.Stdout, "Toys", "1")
	})
}
