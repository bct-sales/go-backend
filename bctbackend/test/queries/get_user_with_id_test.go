//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	lastActivity := models.Timestamp(2)
	user := models.User{
		Password:     "xyz",
		UserId:       models.Id(1),
		RoleId:       models.NewSellerRoleId(),
		CreatedAt:    models.Timestamp(1),
		LastActivity: &lastActivity,
	}

	user.UserId = setup.User(user.RoleId, aux.WithCreatedAt(user.CreatedAt), aux.WithLastActivity(*user.LastActivity), aux.WithPassword(user.Password)).UserId

	actual, err := queries.GetUserWithId(db, user.UserId)
	require.NoError(t, err)
	require.Equal(t, user, *actual)
}
