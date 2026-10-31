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
	setup, db := NewDatabaseFixture(t, WithDefaultCategories)
	defer setup.Close()

	password1 := "xyz"
	password2 := "abc"
	newPassword1 := "123"

	user1ID, err := queries.AddUser(t.Context(), db, models.NewSellerRoleID(), 0, nil, password1)
	require.NoError(t, err)

	user2ID, err := queries.AddUser(t.Context(), db, models.NewSellerRoleID(), 0, nil, password2)
	require.NoError(t, err)

	err = queries.UpdateUserPassword(t.Context(), db, user1ID, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(t.Context(), db, user1ID, newPassword1)
	require.NoError(t, err)

	_, err = queries.AuthenticateUser(t.Context(), db, user2ID, password2)
	require.NoError(t, err)
}
