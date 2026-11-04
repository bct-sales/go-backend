//go:build test

package cli_tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	workspace := NewWorkspace(t)

	run := func(arguments ...string) CommandResult {
		return workspace.Run(t, arguments...)
	}

	t.Run("No arguments prints help", func(t *testing.T) {
		result := run()

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Usage:")
	})

	t.Run("Unknown command fails", func(t *testing.T) {
		result := run("frobnicate")

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "unknown command")
	})

	t.Run("Unknown flag fails", func(t *testing.T) {
		result := run("--frobnicate")

		requireFailure(t, result)
		require.Contains(t, result.Stderr, "unknown flag")
	})

	t.Run("Without database configuration commands fail cleanly", func(t *testing.T) {
		freshWorkspace := NewWorkspace(t)

		result := freshWorkspace.Run(t, "user", "list")

		// The default value of the --db flag never reaches the configuration;
		// a database path must come from --db, the config file or the environment.
		requireFailure(t, result)
		require.Contains(t, result.Stderr, "database path is not set in configuration")
	})

	t.Run("Database path can come from the BCT_DATABASE environment variable", func(t *testing.T) {
		freshWorkspace := NewWorkspace(t)
		databasePath := freshWorkspace.DatabasePath()

		result := freshWorkspace.Run(t, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = freshWorkspace.RunWithEnv(t, []string{"BCT_DATABASE=" + databasePath}, "user", "list")

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of users listed: 0")
	})

	t.Run("Database path can come from the configuration file", func(t *testing.T) {
		freshWorkspace := NewWorkspace(t)
		databasePath := freshWorkspace.DatabasePath()

		result := freshWorkspace.Run(t, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		configuration := "database: '" + databasePath + "'\n"
		require.NoError(t, os.WriteFile(freshWorkspace.Path("bctconfig.yaml"), []byte(configuration), 0o644))

		result = freshWorkspace.Run(t, "user", "list")

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of users listed: 0")
	})

	t.Run("Verbose mode enables debug logging", func(t *testing.T) {
		result := run("-v", "version")

		requireSuccess(t, result)
		require.Contains(t, result.Stderr, "Verbose mode enabled")
	})

	t.Run("No-color removes ANSI escape sequences", func(t *testing.T) {
		freshWorkspace := NewWorkspace(t)
		databasePath := freshWorkspace.DatabasePath()

		result := freshWorkspace.Run(t, "db", "init", "--db", databasePath)
		requireSuccess(t, result)

		result = freshWorkspace.Run(t, "user", "add", "--id", "1", "--role", "admin", "--password", "secret123", "--db", databasePath)
		requireSuccess(t, result)

		result = freshWorkspace.Run(t, "user", "list", "--no-color", "--db", databasePath)

		requireSuccess(t, result)
		require.Contains(t, result.Stdout, "Number of users listed: 1")
		require.NotContains(t, result.Stdout, "\x1b[")
	})
}
