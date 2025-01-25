//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddUserWithId(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, userId := range []models.Id{1, 5} {
			for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
				t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
					db := OpenInitializedDatabase()
					defer db.Close()

					err := queries.AddUserWithId(db, userId, roleId, 0, nil, password)
					require.NoError(t, err)
					require.True(t, queries.UserWithIdExists(db, userId))

					actualRoleId, err := queries.AuthenticateUser(db, userId, password)
					require.NoError(t, err)
					require.Equal(t, roleId, actualRoleId)
				})
			}
		}
	}
}

func TestAddUserWithExistingId(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	userId := models.NewId(1)
	roleId := models.SellerRoleId
	password := "xyz"
	createdAt := models.Timestamp(0)
	var lastAccess *models.Timestamp = nil

	{
		err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
		require.NoError(t, err)
	}

	{
		err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
		var userIdAlreadyInUseError *queries.UserIdAlreadyInUseError
		require.ErrorAs(t, err, &userIdAlreadyInUseError)
	}
}

func TestAddUserWithIdWithInvalidRole(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	userId := models.NewId(1)
	roleId := models.Id(10)
	password := "xyz"
	createdAt := models.Timestamp(0)
	var lastAccess *models.Timestamp = nil

	require.False(t, models.IsValidRole(roleId), "sanity test: role id should be invalid")

	{
		err := queries.AddUserWithId(db, userId, roleId, createdAt, lastAccess, password)
		var noSuchRoleError *queries.NoSuchRoleError
		require.ErrorAs(t, err, &noSuchRoleError)
	}
}

func TestAddUser(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
			t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
				db := OpenInitializedDatabase()
				defer db.Close()

				userId, err := queries.AddUser(db, roleId, 0, nil, password)
				require.NoError(t, err)
				require.True(t, queries.UserWithIdExists(db, userId))
			})
		}
	}
}

func TestAddUserWithInvalidRole(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	roleId := models.Id(10)
	password := "xyz"
	createdAt := models.Timestamp(0)
	var lastActivity *models.Timestamp = nil

	require.False(t, models.IsValidRole(roleId), "sanity test: role id should be invalid")

	{
		_, err := queries.AddUser(db, roleId, createdAt, lastActivity, password)
		var noSuchRoleError *queries.NoSuchRoleError
		require.ErrorAs(t, err, &noSuchRoleError)
	}
}

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

func TestUpdatePassword(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	password1 := "xyz"
	password2 := "abc"
	newPassword1 := "123"

	user1Id, err := queries.AddUser(db, models.SellerRoleId, 0, nil, password1)
	require.NoError(t, err)

	user2Id, err := queries.AddUser(db, models.SellerRoleId, 0, nil, password2)
	require.NoError(t, err)

	err = queries.UpdateUserPassword(db, user1Id, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(db, user1Id, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(db, user2Id, password2)
	require.NoError(t, err)
}
