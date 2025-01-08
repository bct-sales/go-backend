//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddSession(t *testing.T) {
	for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		userId := test.AddUserToDatabase(db, roleId).UserId
		expirationTime := models.Timestamp(0)
		sessionId, err := queries.AddSession(db, userId, expirationTime)
		require.NoError(t, err)

		session, err := queries.GetSessionById(db, sessionId)
		require.NoError(t, err)
		require.Equal(t, sessionId, session.SessionId)
		require.Equal(t, userId, session.UserId)
		require.Equal(t, expirationTime, session.ExpirationTime)
	}
}
