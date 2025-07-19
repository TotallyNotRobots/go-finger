package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/internal/middleware"
)

func TestWrapResponseWriter(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	wrapped := middleware.WrapResponseWriter(w)

	require.NotNil(t, wrapped)
}

func TestResponseWrapper_Status(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	wrapped := middleware.WrapResponseWriter(w)

	require.Equal(t, 0, wrapped.Status())

	wrapped.WriteHeader(http.StatusOK)

	require.Equal(t, http.StatusOK, wrapped.Status())
}

type FailWriter struct{}

func (w *FailWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("error") //nolint:err113 // We want to return an error
}

func (w *FailWriter) Header() http.Header {
	return http.Header{}
}

func (w *FailWriter) WriteHeader(_ int) {}

func TestResponseWrapper_Write(t *testing.T) {
	t.Parallel()

	t.Run("writes success messages", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		wrapped := middleware.WrapResponseWriter(w)

		size, err := wrapped.Write([]byte("test"))
		require.NoError(t, err)

		require.Equal(t, 4, size)

		require.Equal(t, http.StatusOK, wrapped.Status())
	})

	t.Run("returns error on fail write", func(t *testing.T) {
		t.Parallel()

		w := &FailWriter{}
		wrapped := middleware.WrapResponseWriter(w)

		_, err := wrapped.Write([]byte("test"))
		require.Error(t, err)
	})
}

func TestResponseWrapper_Unwrap(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	wrapped := middleware.WrapResponseWriter(w)

	require.Equal(t, w, wrapped.Unwrap())
}
