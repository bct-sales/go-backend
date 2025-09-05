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
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin()
			sessionId := setup.Session(user.UserId)

			err := queries.DeleteSessionWithUser(db, user.UserId)
			require.NoError(t, err)

			_, err = queries.GetSessionById(db, sessionId)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})

		t.Run("Multiple sessions, different users", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin()
			user2 := setup.Cashier()

			sessionId := setup.Session(user.UserId)
			sessionId2 := setup.Session(user2.UserId)

			err := queries.DeleteSessionWithUser(db, user.UserId)
			require.NoError(t, err)

			{
				_, err := queries.GetSessionById(db, sessionId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}

			{
				_, err := queries.GetSessionById(db, sessionId2)
				require.NoError(t, err)
			}
		})

		t.Run("Multiple sessions for one user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.Admin()

			sessionId := setup.Session(user.UserId)
			sessionId2 := setup.Session(user.UserId)

			err := queries.DeleteSessionWithUser(db, user.UserId)
			require.NoError(t, err)

			{
				_, err := queries.GetSessionById(db, sessionId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}

			{
				_, err := queries.GetSessionById(db, sessionId2)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
			}
		})
	})
}
