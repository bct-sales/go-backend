//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAuthenticatingSuccessfully(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId
	createdAt := models.NewTimestamp(0)
	var lastActivity *models.Timestamp = nil

	queries.AddUserWithId(db, userId, roleId, createdAt, lastActivity, password)

	actualRoleId, err := queries.AuthenticateUser(db, userId, password)

	require.NoError(t, err)

	require.Equal(t, roleId, actualRoleId)
}

func TestAuthenticatingNonExistingUser(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	userId := models.NewId(5)

	require.False(t, queries.UserWithIdExists(db, userId))

	_, err := queries.AuthenticateUser(db, userId, password)

	require.Error(t, err)
	require.IsType(t, &queries.UnknownUserError{}, err)
}

func TestAuthenticatingWrongPassword(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	wrongPassword := "abc"
	userId := models.NewId(5)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, nil, password)

	_, err := queries.AuthenticateUser(db, userId, wrongPassword)

	require.Error(t, err)
	require.IsType(t, &queries.WrongPasswordError{}, err)
}
