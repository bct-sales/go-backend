//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddSession(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, roleId := range models.ListRoles() {
			testLabel := fmt.Sprintf("Role=%s", roleId.Name())

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				user := setup.User(roleId)
				expirationTime := models.Timestamp(0)
				sessionId, err := queries.AddSession(db, user.UserID, expirationTime)
				require.NoError(t, err)

				session, err := queries.GetSessionById(db, sessionId)
				require.NoError(t, err)
				require.Equal(t, sessionId, session.SessionID)
				require.Equal(t, user.UserID, session.UserID)
				require.Equal(t, expirationTime, session.ExpirationTime)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			userId := models.ID(999)
			setup.RequireNoSuchUsers(t, userId)
			expirationTime := models.Timestamp(0)
			_, err := queries.AddSession(db, userId, expirationTime)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
