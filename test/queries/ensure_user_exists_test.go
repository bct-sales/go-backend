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

func TestEnsureUserExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		user := setup.Seller()

		err := queries.EnsureUserExists(db, user.UserID)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentUserID := models.ID(999)
		setup.RequireNoSuchUsers(t, nonexistentUserID)

		err := queries.EnsureUserExists(db, nonexistentUserID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})
}
