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
		for _, roleId := range models.ListRoles() {
			testLabel := roleId.Name()

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.User(roleId), aux.WithExpiration(100))
				sessionData, err := queries.GetSessionData(db, sessionId, setup.Clock.Now())
				expectedExpirationTime := setup.Clock.Now() + models.Timestamp(100)

				require.NoError(t, err)
				require.NotNil(t, sessionData)
				require.Equal(t, seller.UserID, sessionData.UserId)
				require.Equal(t, roleId, sessionData.RoleId)
				require.Equal(t, expectedExpirationTime, sessionData.ExpirationTime)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid session ID", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidSessionId := models.SessionID("invalid-session-id")
			_, err := queries.GetSessionData(db, invalidSessionId, setup.Clock.Now())
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})

		t.Run("Expired session", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			sessionId := setup.Session(seller.UserID, aux.WithExpiration(10))

			// Advance time to ensure session is expired
			setup.Clock.Advance(11)

			_, err := queries.GetSessionData(db, sessionId, setup.Clock.Now())
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
		})
	})
}
