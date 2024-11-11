//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddUserWithId(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, userId := range []models.Id{1, 5} {
			for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
				t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
					db := OpenInitializedDatabase()

					err := queries.AddUserWithId(db, userId, roleId, 0, password)

					if assert.NoError(t, err) {
						assert.True(t, queries.UserWithIdExists(db, userId))
						assert.NoError(t, queries.AuthenticateUser(db, userId, password))
					}
				})
			}
		}
	}
}

func TestAddUser(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
			t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
				db := OpenInitializedDatabase()

				userId, err := queries.AddUser(db, roleId, 0, password)

				if assert.NoError(t, err) {
					assert.True(t, queries.UserWithIdExists(db, userId))
				}
			})
		}
	}
}

func TestAuthenticatingSuccessfully(t *testing.T) {
	db := OpenInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, password)

	assert.NoError(t, queries.AuthenticateUser(db, userId, password))
}

func TestAuthenticatingNonExistingUser(t *testing.T) {
	db := OpenInitializedDatabase()

	password := "xyz"
	userId := models.NewId(5)

	assert.False(t, queries.UserWithIdExists(db, userId))
	assert.Error(t, queries.AuthenticateUser(db, userId, password))
}

func TestAuthenticatingWrongPassword(t *testing.T) {
	db := OpenInitializedDatabase()

	password := "xyz"
	wrongPassword := "abc"
	userId := models.NewId(5)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, password)

	assert.Error(t, queries.AuthenticateUser(db, userId, wrongPassword))
}

func TestGetUser(t *testing.T) {
	db := OpenInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, password)

	user, err := queries.GetUserWithId(db, userId)

	if assert.NoError(t, err) {
		assert.Equal(t, userId, user.UserId)
		assert.Equal(t, roleId, user.RoleId)
	}
}

func TestListUsers(t *testing.T) {
	db := OpenInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, password)

	users, err := queries.ListUsers(db)

	if assert.NoError(t, err) {
		assert.Len(t, users, 1)
		assert.Equal(t, userId, users[0].UserId)
		assert.Equal(t, roleId, users[0].RoleId)
		assert.Equal(t, password, users[0].Password)
	}
}

func TestUpdatePassword(t *testing.T) {
	db := OpenInitializedDatabase()

	password1 := "xyz"
	password2 := "abc"
	newPassword1 := "123"

	user1Id, err := queries.AddUser(db, models.SellerRoleId, 0, password1)

	if assert.NoError(t, err) {
		user2Id, err := queries.AddUser(db, models.SellerRoleId, 0, password2)

		if assert.NoError(t, err) {
			err := queries.UpdateUserPassword(db, user1Id, newPassword1)

			if assert.NoError(t, err) {
				assert.NoError(t, queries.AuthenticateUser(db, user1Id, newPassword1))
				assert.NoError(t, queries.AuthenticateUser(db, user2Id, password2))
			}
		}
	}
}
