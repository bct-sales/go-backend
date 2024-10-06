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
	password := "test"
	salt := "xxx"
	hash := security.HashPassword(password, salt)

	for _, userId := range []models.Id{1, 5} {
		for _, roleId := range []models.Id{models.AdminRoleId} {
			t.Run(fmt.Sprintf("With role id %d", roleId), func(t *testing.T) {
				db := openInitializedDatabase()

				AddUser(db, userId, roleId, 0, hash, salt)

				assert.True(t, UserWithIdExists(db, userId))
			})
		}
	}
}
