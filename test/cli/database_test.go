//go:build test

package cli_tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseCommands(t *testing.T) {
	run := func(workspace *Workspace, arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	t.Run("Init creates the database", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Database file successfully created at")
		require.FileExists(t, databasePath)
	})

	t.Run("Init refuses to overwrite an existing database", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "db", "init", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "already exists")
	})

	t.Run("Init with no-categories skips the default categories", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--no-categories", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "category", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.NotContains(t, result.Stdout, "Books")
	})

	t.Run("Backup to an explicit file", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "db", "backup", "mybackup.db", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Database backup completed successfully to mybackup.db")
		require.FileExists(t, workspace.Path("mybackup.db"))
	})

	t.Run("Backup with a default timestamped file name", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "db", "backup", "--db", databasePath)

		requireSuccess(t, result)
		backups, err := filepath.Glob(workspace.Path("backup-*.db"))
		require.NoError(t, err)
		require.Len(t, backups, 1)
	})

	t.Run("Backup of a missing database fails", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "backup", "x.db", "--db", databasePath)

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "Failed to open database")
	})

	t.Run("Dummy refuses to run without overwrite", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "db", "dummy", "--db", databasePath)

		// Note: the command exits with code 0 even though it did nothing.
		requireSuccess(t, result)
		require.Contains(t, result.Stderr, "Missing --overwrite; database left untouched")
		require.Empty(t, result.Stdout)
	})

	t.Run("Dummy with overwrite populates the database", func(t *testing.T) {
		workspace := NewWorkspace(t)
		databasePath := workspace.DatabasePath()

		result := run(workspace, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = run(workspace, "db", "dummy", "--overwrite", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Resetting database")
		require.Contains(t, result.Stdout, "Adding admin user")

		result = run(workspace, "user", "list", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "admin")
		require.Contains(t, result.Stdout, "abc")
	})

	t.Run("Dummy is deterministic given a seed", func(t *testing.T) {
		csvOutputs := make([]string, 2)

		for i := range csvOutputs {
			workspace := NewWorkspace(t)
			databasePath := workspace.DatabasePath()

			result := run(workspace, "db", "init", "--db", databasePath)
			requireSuccess(t, result)

			result = run(workspace, "db", "dummy", "--overwrite", "--seed", "42", "--db", databasePath)
			requireSuccess(t, result)

			result = run(workspace, "user", "list", "--format", "csv", "--db", databasePath)
			requireSuccess(t, result)

			csvOutputs[i] = result.Stdout
		}

		require.Equal(t, csvOutputs[0], csvOutputs[1])
	})
}
