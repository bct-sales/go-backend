//go:build test

package path

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	t.Run("Root", func(t *testing.T) {
		path := NewRootPath("test")

		require.Equal(t, "test", path.String())
	})

	t.Run("Login", func(t *testing.T) {
		require.Equal(t, "/api/v1/login", Root().Login().String())
	})
}
