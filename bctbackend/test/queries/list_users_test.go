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

func TestListUsers(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, nil, password)

	users := []*models.User{}
	err := queries.GetUsers(db, queries.CollectTo(&users))
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, userId, users[0].UserId)
	require.Equal(t, roleId, users[0].RoleId)
	require.Equal(t, password, users[0].Password)
}
