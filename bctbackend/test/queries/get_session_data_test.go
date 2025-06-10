//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSessionData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, roleId := range []models.RoleId{models.NewSellerRoleId(), models.NewAdminRoleId(), models.NewCashierRoleId()} {
			testLabel := roleId.Name()

			t.Run(testLabel, func(t *testing.T) {
				t.Parallel()

				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller, sessionId := setup.LoggedIn(setup.User(roleId))
				sessionData, err := queries.GetSessionData(db, sessionId)
				require.NoError(t, err)
				require.NotNil(t, sessionData)
				require.Equal(t, seller.UserId, sessionData.UserId)
				require.Equal(t, roleId, sessionData.RoleId)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		invalidSessionId := models.SessionId("invalid-session-id")
		_, err := queries.GetSessionData(db, invalidSessionId)
		require.ErrorIs(t, err, dberr.ErrNoSuchSession)
	})
}
