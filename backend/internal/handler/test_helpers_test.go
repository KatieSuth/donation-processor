package handler_test

import (
	"testing"

	"github.com/KatieSuth/donation-processor/backend/internal/handler"
	"github.com/KatieSuth/donation-processor/backend/internal/store"
	"github.com/stretchr/testify/require"
)

// newTestHandler constructs a real *Handler with controlled dependencies.
func newTestHandler(t *testing.T, s store.Store) *handler.Handler {
	t.Helper()
	require.NotNil(t, s)
	return handler.New("http://localhost:3000", s)
}
