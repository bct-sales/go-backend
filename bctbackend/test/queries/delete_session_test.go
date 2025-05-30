//go:build test

package queries

import (
	"bctbackend/database"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestDeleteSession(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	user := setup.Admin()
	expirationTime := models.Timestamp(0)
	sessionId, err := queries.AddSession(db, user.UserId, expirationTime)
	require.NoError(t, err)

	err = queries.DeleteSession(db, sessionId)
	require.NoError(t, err)

	_, err = queries.GetSessionById(db, sessionId)
	require.ErrorIs(t, err, database.ErrNoSuchSession)
}
