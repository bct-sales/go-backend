//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAuthenticatingSuccessfully(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	userId := models.NewId(1)
	roleId := models.SellerRoleId
	createdAt := models.NewTimestamp(0)
	var lastActivity *models.Timestamp = nil

	queries.AddUserWithId(db, userId, roleId, createdAt, lastActivity, password)

	actualRoleId, err := queries.AuthenticateUser(db, userId, password)
	if assert.NoError(t, err) {
		assert.Equal(t, roleId, actualRoleId)
	}
}

func TestAuthenticatingNonExistingUser(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	userId := models.NewId(5)

	assert.False(t, queries.UserWithIdExists(db, userId))

	_, err := queries.AuthenticateUser(db, userId, password)

	if assert.Error(t, err) {
		assert.IsType(t, &queries.UnknownUserError{}, err)
	}
}

func TestAuthenticatingWrongPassword(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	password := "xyz"
	wrongPassword := "abc"
	userId := models.NewId(5)
	roleId := models.SellerRoleId

	queries.AddUserWithId(db, userId, roleId, 0, nil, password)

	_, err := queries.AuthenticateUser(db, userId, wrongPassword)

	if assert.Error(t, err) {
		assert.IsType(t, &queries.WrongPasswordError{}, err)
	}
}
