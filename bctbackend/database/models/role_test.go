package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleParsing(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleId, err := ParseRole("admin")

		if assert.NoError(t, err) {
			assert.Equal(t, AdminRoleId, roleId)
		}
	})

	t.Run("seller", func(t *testing.T) {
		roleId, err := ParseRole("seller")

		if assert.NoError(t, err) {
			assert.Equal(t, SellerRoleId, roleId)
		}
	})

	t.Run("cashier", func(t *testing.T) {
		roleId, err := ParseRole("cashier")

		if assert.NoError(t, err) {
			assert.Equal(t, CashierRoleId, roleId)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := ParseRole("invalid")

		assert.Error(t, err)
	})
}

func TestRoleToString(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleName, err := RoleToString(AdminRoleId)

		if assert.NoError(t, err) {
			assert.Equal(t, AdminName, roleName)
		}
	})

	t.Run("seller", func(t *testing.T) {
		roleName, err := RoleToString(SellerRoleId)

		if assert.NoError(t, err) {
			assert.Equal(t, SellerName, roleName)
		}
	})

	t.Run("cashier", func(t *testing.T) {
		roleName, err := RoleToString(CashierRoleId)

		if assert.NoError(t, err) {
			assert.Equal(t, CashierName, roleName)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := RoleToString(4)

		assert.Error(t, err)
	})
}
