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

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, nil, password)

	user, err := queries.GetUserWithId(db, userId)
	require.NoError(t, err)
	require.Equal(t, userId, user.UserId)
	require.Equal(t, roleId, user.RoleId)
}
