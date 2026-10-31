//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteSessionWithUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single session", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin(t)
			sessionID := setup.Session(t, user.UserID)

			err := queries.DeleteSessionWithUser(t.Context(), db, user.UserID)
			require.NoError(t, err)

			_, err = queries.GetSessionByID(t.Context(), db, sessionID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})

		t.Run("Multiple sessions, different users", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin(t)
			user2 := setup.Cashier(t)

			sessionID := setup.Session(t, user.UserID)
			sessionID2 := setup.Session(t, user2.UserID)

			err := queries.DeleteSessionWithUser(t.Context(), db, user.UserID)
			require.NoError(t, err)

			{
				_, err := queries.GetSessionByID(t.Context(), db, sessionID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}

			{
				_, err := queries.GetSessionByID(t.Context(), db, sessionID2)
				require.NoError(t, err)
			}
		})

		t.Run("Multiple sessions for one user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin(t)

			sessionID := setup.Session(t, user.UserID)
			sessionID2 := setup.Session(t, user.UserID)

			err := queries.DeleteSessionWithUser(t.Context(), db, user.UserID)
			require.NoError(t, err)

			{
				_, err := queries.GetSessionByID(t.Context(), db, sessionID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}

			{
				_, err := queries.GetSessionByID(t.Context(), db, sessionID2)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}
		})
	})
}
