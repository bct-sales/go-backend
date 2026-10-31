//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSessionData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, roleID := range models.ListRoles() {
			testLabel := roleID.Name()

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller, sessionID := setup.LoggedIn(t, setup.User(t, roleID), aux.WithExpiration(100))
				sessionData, err := queries.GetSessionData(t.Context(), db, sessionID, setup.Clock.Now())
				expectedExpirationTime := setup.Clock.Now() + models.Timestamp(100)

				require.NoError(t, err)
				require.NotNil(t, sessionData)
				require.Equal(t, seller.UserID, sessionData.UserID)
				require.Equal(t, roleID, sessionData.RoleID)
				require.Equal(t, expectedExpirationTime, sessionData.ExpirationTime)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid session ID", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			invalidSessionID := models.SessionID("invalid-session-id")
			_, err := queries.GetSessionData(t.Context(), db, invalidSessionID, setup.Clock.Now())
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})

		t.Run("Expired session", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			sessionID := setup.Session(t, seller.UserID, aux.WithExpiration(10))

			// Advance time to ensure session is expired
			setup.Clock.Advance(11)

			_, err := queries.GetSessionData(t.Context(), db, sessionID, setup.Clock.Now())
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})
	})
}
