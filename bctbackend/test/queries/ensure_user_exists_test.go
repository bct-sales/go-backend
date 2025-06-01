//go:build test

package queries

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestEnsureUserExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		user := setup.Seller()

		err := queries.EnsureUserExists(db, user.UserId)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentUserId := models.Id(999)
		setup.RequireNoSuchUsers(t, nonexistentUserId)

		err := queries.EnsureUserExists(db, nonexistentUserId)
		require.ErrorIs(t, err, database.ErrNoSuchUser)
	})
}
