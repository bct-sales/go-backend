//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddSession(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			user := setup.User(roleId)
			expirationTime := models.Timestamp(0)
			sessionId, err := queries.AddSession(db, user.UserId, expirationTime)
			require.NoError(t, err)

			session, err := queries.GetSessionById(db, sessionId)
			require.NoError(t, err)
			require.Equal(t, sessionId, session.SessionId)
			require.Equal(t, user.UserId, session.UserId)
			require.Equal(t, expirationTime, session.ExpirationTime)
		}
	})
}
