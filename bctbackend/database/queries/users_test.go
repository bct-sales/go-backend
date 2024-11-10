//go:build test

package queries

import (
	models "bctbackend/database/models"
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
					db := openInitializedDatabase()

					err := AddUserWithId(db, userId, roleId, 0, password)

					if assert.NoError(t, err) {
						assert.True(t, UserWithIdExists(db, userId))
						assert.NoError(t, AuthenticateUser(db, userId, password))
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
				db := openInitializedDatabase()

				userId, err := AddUser(db, roleId, 0, password)

				if assert.NoError(t, err) {
					assert.True(t, UserWithIdExists(db, userId))
				}
			})
		}
	}
}

func TestAuthenticatingSuccessfully(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	AddUserWithId(db, userId, roleId, 0, password)

	assert.NoError(t, AuthenticateUser(db, userId, password))
}

func TestAuthenticatingNonExistingUser(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	userId := models.NewId(5)

	assert.False(t, UserWithIdExists(db, userId))
	assert.Error(t, AuthenticateUser(db, userId, password))
}

func TestAuthenticatingWrongPassword(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	wrongPassword := "abc"
	userId := models.NewId(5)
	roleId := models.SellerRoleId

	AddUserWithId(db, userId, roleId, 0, password)

	assert.Error(t, AuthenticateUser(db, userId, wrongPassword))
}

func TestGetUser(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	AddUserWithId(db, userId, roleId, 0, password)

	user, err := GetUserWithId(db, userId)

	if assert.NoError(t, err) {
		assert.Equal(t, userId, user.UserId)
		assert.Equal(t, roleId, user.RoleId)
	}
}

func TestListUsers(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	AddUserWithId(db, userId, roleId, 0, password)

	users, err := ListUsers(db)

	if assert.NoError(t, err) {
		assert.Len(t, users, 1)
		assert.Equal(t, userId, users[0].UserId)
		assert.Equal(t, roleId, users[0].RoleId)
		assert.Equal(t, password, users[0].Password)
	}
}

func TestUpdatePassword(t *testing.T) {
	db := openInitializedDatabase()

	password1 := "xyz"
	password2 := "abc"
	newPassword1 := "123"

	user1Id, err := AddUser(db, models.SellerRoleId, 0, password1)

	if assert.NoError(t, err) {
		user2Id, err := AddUser(db, models.SellerRoleId, 0, password2)

		if assert.NoError(t, err) {
			err := UpdateUserPassword(db, user1Id, newPassword1)

			if assert.NoError(t, err) {
				assert.NoError(t, AuthenticateUser(db, user1Id, newPassword1))
				assert.NoError(t, AuthenticateUser(db, user2Id, password2))
			}
		}
	}
}
