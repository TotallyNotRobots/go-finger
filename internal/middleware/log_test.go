package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/internal/config"
	"git.maronato.dev/maronato/finger/internal/log"
	"git.maronato.dev/maronato/finger/internal/middleware"
)

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.NewConfig()

	stdout := &strings.Builder{}

	l := log.NewLogger(stdout, cfg)
	ctx = log.WithLogger(ctx, l)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/", http.NoBody)

	require.Empty(t, stdout.String())

	middleware.RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, stdout.String())
}
