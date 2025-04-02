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

func TestGetUser(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	lastActivity := models.Timestamp(2)
	user := models.User{
		Password:     "xyz",
		UserId:       models.NewId(1),
		RoleId:       models.SellerRoleId,
		CreatedAt:    models.Timestamp(1),
		LastActivity: &lastActivity,
	}

	user.UserId = AddUserToDatabase(db, user.RoleId, WithCreatedAt(user.CreatedAt), WithLastActivity(*user.LastActivity), WithPassword(user.Password)).UserId

	actual, err := queries.GetUserWithId(db, user.UserId)
	require.NoError(t, err)
	require.Equal(t, user, *actual)
}
