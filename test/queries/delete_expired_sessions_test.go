//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredSessions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for cutoff := 0; cutoff < 100; cutoff += 10 {
			testLabel := fmt.Sprintf("cutoff=%d", cutoff)
			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				user := setup.Admin()
				expiredSessions := []models.SessionID{}
				unexpiredSessions := []models.SessionID{}

				for i := 0; i < 100; i++ {
					expirationTime := models.Timestamp(0)
					sessionID, err := queries.AddSession(db, user.UserID, expirationTime)
					require.NoError(t, err)

					if expirationTime < models.Timestamp(cutoff) {
						expiredSessions = append(expiredSessions, sessionID)
					} else {
						unexpiredSessions = append(unexpiredSessions, sessionID)
					}
				}

				err := queries.DeleteExpiredSessions(db, models.Timestamp(cutoff))
				require.NoError(t, err)

				for _, sessionID := range expiredSessions {
					_, err := queries.GetSessionByID(db, sessionID)
					requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSession)
				}

				for _, sessionID := range unexpiredSessions {
					_, err := queries.GetSessionByID(db, sessionID)
					require.NoError(t, err)
				}
			})
		}
	})
}
