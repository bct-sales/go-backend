package queries

import (
	models "bctbackend/db/models"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddUser(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, userId := range []models.Id{1, 5} {
			for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
				t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
					db := openInitializedDatabase()

					AddUser(db, userId, roleId, 0, password)

					assert.True(t, UserWithIdExists(db, userId))
					assert.NoError(t, AuthenticateUser(db, userId, password))
				})
			}
		}
	}
}

func TestAuthenticatingSuccessfully(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId

	AddUser(db, userId, roleId, 0, password)

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

	AddUser(db, userId, roleId, 0, password)

	assert.Error(t, AuthenticateUser(db, userId, wrongPassword))
}
