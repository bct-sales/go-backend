//go:build test

package queries

import (
	"testing"

	dberr "bctbackend/database/errors"

	"github.com/stretchr/testify/require"
)

func requireDatabaseWrappedError(t *testing.T, err error, expectedWrapped error) {
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	var wrappedErr *dberr.ErrDatabase
	require.ErrorAsf(t, err, &wrappedErr, "expected error %v to be wrapped in ErrDatabase", err)
	require.ErrorIs(t, err, expectedWrapped)
}
