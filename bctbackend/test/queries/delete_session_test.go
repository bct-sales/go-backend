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
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		user := setup.Admin()
		expirationTime := models.Timestamp(0)
		sessionId, err := queries.AddSession(db, user.UserId, expirationTime)
		require.NoError(t, err)

		err = queries.DeleteSession(db, sessionId)
		require.NoError(t, err)

		_, err = queries.GetSessionById(db, sessionId)
		require.ErrorIs(t, err, dberr.ErrNoSuchSession)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentSessionId := models.SessionId("nonexistent-session-id")
		err := queries.DeleteSession(db, nonexistentSessionId)
		require.ErrorIs(t, err, dberr.ErrNoSuchSession)
	})
}
