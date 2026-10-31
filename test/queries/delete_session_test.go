//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteSession(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		user := setup.Admin(t)
		expirationTime := models.Timestamp(0)
		sessionID, err := queries.AddSession(t.Context(), db, user.UserID, expirationTime)
		require.NoError(t, err)

		err = queries.DeleteSession(t.Context(), db, sessionID)
		require.NoError(t, err)

		_, err = queries.GetSessionByID(t.Context(), db, sessionID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		nonexistentSessionID := models.SessionID("nonexistent-session-id")
		err := queries.DeleteSession(t.Context(), db, nonexistentSessionID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
	})
}
