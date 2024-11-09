package queries

import (
	models "bctbackend/database/models"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddSession(t *testing.T) {
	for i := 0; i < 10; i++ {
		for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
			db := openInitializedDatabase()

			userId := addTestUser(db, roleId)
			sessionId, err := AddSession(db, userId)

			if assert.NoError(t, err) {
				session, err := GetSessionById(db, sessionId)

				if assert.NoError(t, err) {
					assert.Equal(t, sessionId, session.SessionId)
					assert.Equal(t, userId, session.UserId)
				}
			}
		}
	}
}
