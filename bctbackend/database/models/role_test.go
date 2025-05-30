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
		require.Equal(t, AdminRoleId, roleId)
	})

	t.Run("seller", func(t *testing.T) {
		t.Parallel()
		roleId, err := ParseRole("seller")
		require.NoError(t, err)
		require.Equal(t, SellerRoleId, roleId)
	})

	t.Run("cashier", func(t *testing.T) {
		t.Parallel()
		roleId, err := ParseRole("cashier")
		require.NoError(t, err)
		require.Equal(t, CashierRoleId, roleId)
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRole("invalid")
		require.Error(t, err)
	})
}

func TestNameOfRole(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()
		roleName, err := NameOfRole(AdminRoleId)
		require.NoError(t, err)
		require.Equal(t, AdminName, roleName)
	})

	t.Run("seller", func(t *testing.T) {
		t.Parallel()
		roleName, err := NameOfRole(SellerRoleId)
		require.NoError(t, err)
		require.Equal(t, SellerName, roleName)
	})

	t.Run("cashier", func(t *testing.T) {
		t.Parallel()
		roleName, err := NameOfRole(CashierRoleId)
		require.NoError(t, err)
		require.Equal(t, CashierName, roleName)
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		_, err := NameOfRole(4)
		require.Error(t, err)
	})
}
