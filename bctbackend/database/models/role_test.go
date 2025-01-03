package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleParsing(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleId, err := ParseRole("admin")
		require.NoError(t, err)
		require.Equal(t, AdminRoleId, roleId)
	})

	t.Run("seller", func(t *testing.T) {
		roleId, err := ParseRole("seller")
		require.NoError(t, err)
		require.Equal(t, SellerRoleId, roleId)
	})

	t.Run("cashier", func(t *testing.T) {
		roleId, err := ParseRole("cashier")
		require.NoError(t, err)
		require.Equal(t, CashierRoleId, roleId)
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := ParseRole("invalid")
		require.Error(t, err)
	})
}

func TestNameOfRole(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleName, err := NameOfRole(AdminRoleId)
		require.NoError(t, err)
		require.Equal(t, AdminName, roleName)
	})

	t.Run("seller", func(t *testing.T) {
		roleName, err := NameOfRole(SellerRoleId)
		require.NoError(t, err)
		require.Equal(t, SellerName, roleName)
	})

	t.Run("cashier", func(t *testing.T) {
		roleName, err := NameOfRole(CashierRoleId)
		require.NoError(t, err)
		require.Equal(t, CashierName, roleName)
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := NameOfRole(4)
		require.Error(t, err)
	})
}
