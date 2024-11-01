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
