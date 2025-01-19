//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestDeleteSession(t *testing.T) {
	db := setup.OpenInitializedDatabase()
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
