//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateLastActivity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, roleId := range models.ListRoles() {
			for _, lastActivity := range []models.Timestamp{500, 1000, 2000} {
				testLabel := fmt.Sprintf("Role: %s, LastActivity: %d", roleId, lastActivity)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.User(roleId, aux.WithNoLastActivity())

					queries.UpdateLastActivity(db, seller.UserID, lastActivity)

					userData, err := queries.GetUserWithId(db, seller.UserID)
					require.NoError(t, err)
					require.NotNil(t, userData)
					require.NotNil(t, userData.LastActivity)
					require.Equal(t, lastActivity, *userData.LastActivity)
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			userId := models.ID(999)
			setup.RequireNoSuchUsers(t, userId)

			err := queries.UpdateLastActivity(db, userId, 1000)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})
	})
}
