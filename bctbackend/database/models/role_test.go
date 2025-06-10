package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleParsing(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()
		roleId, err := ParseRole("admin")
		require.NoError(t, err)
		require.Equal(t, NewAdminRoleId(), roleId)
	})

	t.Run("seller", func(t *testing.T) {
		t.Parallel()
		roleId, err := ParseRole("seller")
		require.NoError(t, err)
		require.Equal(t, NewSellerRoleId(), roleId)
	})

	t.Run("cashier", func(t *testing.T) {
		t.Parallel()
		roleId, err := ParseRole("cashier")
		require.NoError(t, err)
		require.Equal(t, NewCashierRoleId(), roleId)
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRole("invalid")
		require.Error(t, err)
	})
}
