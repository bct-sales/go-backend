//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdatePassword(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	password1 := "xyz"
	password2 := "abc"
	newPassword1 := "123"

	user1Id, err := queries.AddUser(db, models.SellerRoleId, 0, nil, password1)
	require.NoError(t, err)

	user2Id, err := queries.AddUser(db, models.SellerRoleId, 0, nil, password2)
	require.NoError(t, err)

	err = queries.UpdateUserPassword(db, user1Id, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(db, user1Id, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(db, user2Id, password2)
	require.NoError(t, err)
}
