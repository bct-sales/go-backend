//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestDeleteSession(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	user := setup.Admin()
	expirationTime := models.Timestamp(0)
	sessionId, err := queries.AddSession(db, user.UserId, expirationTime)
	require.NoError(t, err)

	err = queries.DeleteSession(db, sessionId)
	require.NoError(t, err)

	_, err = queries.GetSessionById(db, sessionId)
	var noSuchSessionError *queries.NoSuchSessionError
	require.ErrorAs(t, err, &noSuchSessionError)
}
