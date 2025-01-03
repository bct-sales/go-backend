package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleParsing(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleId, err := ParseRole("admin")

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, AdminRoleId, roleId) {
			return
		}
	})

	t.Run("seller", func(t *testing.T) {
		roleId, err := ParseRole("seller")

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, SellerRoleId, roleId) {
			return
		}
	})

	t.Run("cashier", func(t *testing.T) {
		roleId, err := ParseRole("cashier")

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, CashierRoleId, roleId) {
			return
		}
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := ParseRole("invalid")

		if !assert.Error(t, err) {
			return
		}
	})
}

func TestNameOfRole(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		roleName, err := NameOfRole(AdminRoleId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, AdminName, roleName) {
			return
		}
	})

	t.Run("seller", func(t *testing.T) {
		roleName, err := NameOfRole(SellerRoleId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, SellerName, roleName) {
			return
		}
	})

	t.Run("cashier", func(t *testing.T) {
		roleName, err := NameOfRole(CashierRoleId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, CashierName, roleName) {
			return
		}
	})

	t.Run("unknown", func(t *testing.T) {
		_, err := NameOfRole(4)

		if !assert.Error(t, err) {
			return
		}
	})
}
