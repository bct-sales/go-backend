package queries

import (
	models "bctrest/db/models"
	"bctrest/security"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddUser(t *testing.T) {
	for _, password := range []string{"a", "xyz"} {
		for _, salt := range []string{"fdf", "dsa"} {
			for _, userId := range []models.Id{1, 5} {
				for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
					t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
						db := openInitializedDatabase()

						hash := security.HashPassword(password, salt)
						AddUser(db, userId, roleId, 0, hash, salt)

						assert.True(t, UserWithIdExists(db, userId))
						assert.NoError(t, AuthenticateUser(db, userId, password))
					})
				}
			}
		}
	}
}

func TestAuthenticatingNonExistingUser(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	var userId models.Id = 5

	assert.False(t, UserWithIdExists(db, userId))
	assert.Error(t, AuthenticateUser(db, userId, password))
}

func TestAuthenticatingWrongPassword(t *testing.T) {
	db := openInitializedDatabase()

	password := "xyz"
	wrongPassword := "abc"
	salt := "123"
	var userId models.Id = 5
	roleId := models.SellerRoleId
	hash := security.HashPassword(password, salt)

	AddUser(db, userId, roleId, 0, hash, salt)

	assert.Error(t, AuthenticateUser(db, userId, wrongPassword))
}
