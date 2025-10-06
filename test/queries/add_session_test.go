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
		for _, roleID := range models.ListRoles() {
			testLabel := fmt.Sprintf("Role=%s", roleID.Name())

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				user := setup.User(roleID)
				expirationTime := models.Timestamp(0)
				sessionID, err := queries.AddSession(db, user.UserID, expirationTime)
				require.NoError(t, err)

				session, err := queries.GetSessionByID(db, sessionID)
				require.NoError(t, err)
				require.Equal(t, sessionID, session.SessionID)
				require.Equal(t, user.UserID, session.UserID)
				require.Equal(t, expirationTime, session.ExpirationTime)
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentUserID := setup.GenerateNonexistentUserID(t)
			expirationTime := models.Timestamp(0)
			_, err := queries.AddSession(db, nonexistentUserID, expirationTime)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
