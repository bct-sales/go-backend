//go:build test

package cli_tests

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// setupSessionTest creates a workspace with an initialized database and an
// admin user with ID 1.
func setupSessionTest(t *testing.T) (func(...string) CommandResult, string) {
	t.Helper()

	run, databasePath := setupDatabaseTest(t)

	requireSuccess(t, run("user", "add", "--id", "1", "--role", "admin", "--password", "pw1", "--db", databasePath))

	return run, databasePath
}

// addSession adds a session for user 1 with the given expiration and returns
// the session ID (the command prints it without any decoration).
func addSession(t *testing.T, run func(...string) CommandResult, databasePath string, expiration string) string {
	t.Helper()

	result := run("session", "add", "-u", "1", "-e", expiration, "--db", databasePath)
	requireSuccess(t, result)

	sessionID := strings.TrimSpace(result.Stdout)
	require.NotEmpty(t, sessionID)

	return sessionID
}

func TestSessionCommands(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		t.Run("Add prints the bare session ID", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)

			result := run("session", "add", "-u", "1", "-e", "24h", "--db", databasePath)

			requireSuccess(t, result)
			require.Equal(t, strings.TrimSpace(result.Stdout), result.Stdout)
			require.NotEmpty(t, result.Stdout)
		})

		t.Run("Add with a nonexistent user fails", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)

			result := run("session", "add", "-u", "999", "-e", "1h", "--db", databasePath)

			requireFailure(t, result)
		})

		t.Run("Add with an invalid expiration fails", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)

			result := run("session", "add", "-u", "1", "-e", "1x", "--db", databasePath)

			requireFailure(t, result)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("List on empty database", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)

			result := run("session", "list", "--db", databasePath)

			requireSuccess(t, result)
			require.Contains(t, result.Stdout, "no sessions found")
		})

		t.Run("List shows added sessions", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)
			sessionID1 := addSession(t, run, databasePath, "24h")
			sessionID2 := addSession(t, run, databasePath, "48h")

			result := run("session", "list", "--no-color", "--db", databasePath)

			requireSuccess(t, result)
			require.Contains(t, result.Stdout, sessionID1)
			require.Contains(t, result.Stdout, sessionID2)
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("Remove by session ID", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)
			sessionID1 := addSession(t, run, databasePath, "24h")
			sessionID2 := addSession(t, run, databasePath, "48h")

			result := run("session", "remove", "-s", sessionID1, "--db", databasePath)

			requireSuccess(t, result)
			require.Empty(t, result.Stdout)

			result = run("session", "list", "--no-color", "--db", databasePath)

			requireSuccess(t, result)
			require.NotContains(t, result.Stdout, sessionID1)
			require.Contains(t, result.Stdout, sessionID2)
		})

		t.Run("Remove by user ID", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)
			addSession(t, run, databasePath, "24h")
			addSession(t, run, databasePath, "48h")

			result := run("session", "remove", "-u", "1", "--db", databasePath)

			requireSuccess(t, result)

			result = run("session", "list", "--db", databasePath)

			requireSuccess(t, result)
			require.Contains(t, result.Stdout, "no sessions found")
		})

		t.Run("Remove without flags fails", func(t *testing.T) {
			// This documents current behavior: although --all defaults to true,
			// cobra's one-required validation fires because the flag was not
			// explicitly given, so a bare `session remove` fails cleanly instead
			// of deleting every session.
			run, databasePath := setupSessionTest(t)

			result := run("session", "remove", "--db", databasePath)

			requireFailure(t, result)
			require.Contains(t, result.Stderr, "at least one of the flags in the group [session user all] is required")
		})

		t.Run("Remove with all removes every session", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)
			addSession(t, run, databasePath, "24h")
			addSession(t, run, databasePath, "48h")

			result := run("session", "remove", "--all", "--db", databasePath)

			requireSuccess(t, result)

			result = run("session", "list", "--db", databasePath)

			requireSuccess(t, result)
			require.Contains(t, result.Stdout, "no sessions found")
		})
	})

	t.Run("Prune", func(t *testing.T) {
		t.Run("Prune removes only expired sessions", func(t *testing.T) {
			run, databasePath := setupSessionTest(t)
			expiredSessionID := addSession(t, run, databasePath, "1s")
			activeSessionID := addSession(t, run, databasePath, "24h")

			time.Sleep(2 * time.Second)

			result := run("session", "prune", "--db", databasePath)

			requireSuccess(t, result)
			require.Empty(t, result.Stdout)

			result = run("session", "list", "--no-color", "--db", databasePath)

			requireSuccess(t, result)
			require.NotContains(t, result.Stdout, expiredSessionID)
			require.Contains(t, result.Stdout, activeSessionID)
		})
	})
}
