//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestDeleteSession(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	userId := test.AddUserToDatabase(db, models.AdminRoleId).UserId
	expirationTime := models.Timestamp(0)
	sessionId, err := queries.AddSession(db, userId, expirationTime)
	require.NoError(t, err)

	err = queries.DeleteSession(db, sessionId)
	require.NoError(t, err)

	_, err = queries.GetSessionById(db, sessionId)
	var noSuchSessionError *queries.NoSuchSessionError
	require.ErrorAs(t, err, &noSuchSessionError)
}

func TestDeleteExpiredSessions(t *testing.T) {
	for cutoff := 0; cutoff < 100; cutoff += 10 {
		testLabel := fmt.Sprintf("cutoff=%d", cutoff)
		t.Run(testLabel, func(t *testing.T) {
			db := test.OpenInitializedDatabase()
			defer db.Close()

			userId := test.AddUserToDatabase(db, models.AdminRoleId).UserId
			expiredSessions := []models.SessionId{}
			unexpiredSessions := []models.SessionId{}

			for i := 0; i < 100; i++ {
				expirationTime := models.Timestamp(0)
				sessionId, err := queries.AddSession(db, userId, expirationTime)
				require.NoError(t, err)

				if expirationTime < models.Timestamp(cutoff) {
					expiredSessions = append(expiredSessions, sessionId)
				} else {
					unexpiredSessions = append(unexpiredSessions, sessionId)
				}
			}

			err := queries.DeleteExpiredSessions(db, models.Timestamp(cutoff))
			require.NoError(t, err)

			for _, sessionId := range expiredSessions {
				_, err := queries.GetSessionById(db, sessionId)
				require.Error(t, err)
			}

			for _, sessionId := range unexpiredSessions {
				_, err := queries.GetSessionById(db, sessionId)
				require.NoError(t, err)
			}
		})
	}
}
