//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddSession(t *testing.T) {
	for i := 0; i < 10; i++ {
		for _, roleId := range []models.Id{models.AdminRoleId, models.CashierRoleId, models.SellerRoleId} {
			db := test.OpenInitializedDatabase()

			userId := test.AddUserToDatabase(db, roleId).UserId
			sessionId, err := queries.AddSession(db, userId)

			if assert.NoError(t, err) {
				session, err := queries.GetSessionById(db, sessionId)

				if assert.NoError(t, err) {
					assert.Equal(t, sessionId, session.SessionId)
					assert.Equal(t, userId, session.UserId)
				}
			}
		}
	}
}
