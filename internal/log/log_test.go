package log_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/internal/config"
	"git.maronato.dev/maronato/finger/internal/log"
)

func assertPanic(t *testing.T, f func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()

	// Call the function
	f()
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("defaults to info level", func(t *testing.T) {
		t.Parallel()

		cfg := config.NewConfig()

		w := &strings.Builder{}
		l := log.NewLogger(w, cfg)

		// It shouldn't log debug messages
		l.Debug("test")

		require.Empty(t, w.String())

		// It should log info messages
		l.Info("test")

		require.NotEmpty(t, w.String())
	})

	t.Run("logs debug messages if debug is enabled", func(t *testing.T) {
		t.Parallel()

		cfg := config.NewConfig()
		cfg.Debug = true

		w := &strings.Builder{}
		l := log.NewLogger(w, cfg)

		// It should log debug messages
		l.Debug("test")

		require.NotEmpty(t, w.String())
	})
}

func TestFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.NewConfig()
	l := log.NewLogger(nil, cfg)

	t.Run("panics if no logger in context", func(t *testing.T) {
		t.Parallel()

		assertPanic(t, func() {
			log.FromContext(ctx)
		})
	})

	t.Run("returns logger from context", func(t *testing.T) {
		t.Parallel()

		ctx = log.WithLogger(ctx, l)

		l2 := log.FromContext(ctx)

		require.NotNil(t, l2)
	})
}
