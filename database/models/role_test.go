package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleParsing(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleID, err := ParseRole("admin")
		require.NoError(t, err)
		require.Equal(t, NewAdminRoleID(), roleID)
	})

	t.Run("seller", func(t *testing.T) {
		roleID, err := ParseRole("seller")
		require.NoError(t, err)
		require.Equal(t, NewSellerRoleID(), roleID)
	})

	t.Run("cashier", func(t *testing.T) {
		roleID, err := ParseRole("cashier")
		require.NoError(t, err)
		require.Equal(t, NewCashierRoleID(), roleID)
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := ParseRole("invalid")
		require.Error(t, err)
	})
}
