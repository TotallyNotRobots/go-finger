package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/handler"
	"git.maronato.dev/maronato/finger/internal/config"
	"git.maronato.dev/maronato/finger/internal/log"
	"git.maronato.dev/maronato/finger/webfingers"
)

func TestWebfingerHandler(t *testing.T) {
	t.Parallel()

	fingers := webfingers.WebFingers{
		"acct:user@example.com": {
			Subject: "acct:user@example.com",
			Links: []webfingers.Link{
				{
					Rel:  "http://webfinger.net/rel/profile-page",
					Href: "https://example.com/user",
				},
			},
			Properties: map[string]string{
				"http://webfinger.net/rel/name": "John Doe",
			},
		},
		"acct:other@example.com": {
			Subject: "acct:other@example.com",
			Properties: map[string]string{
				"http://webfinger.net/rel/name": "Jane Doe",
			},
		},
		"https://example.com/user": {
			Subject: "https://example.com/user",
			Properties: map[string]string{
				"http://webfinger.net/rel/name": "John Baz",
			},
		},
	}

	tests := []struct {
		name            string
		resource        string
		wantCode        int
		alternateMethod string
	}{
		{
			name:     "valid resource",
			resource: "acct:user@example.com",
			wantCode: http.StatusOK,
		},
		{
			name:     "other valid resource",
			resource: "acct:other@example.com",
			wantCode: http.StatusOK,
		},
		{
			name:     "url resource",
			resource: "https://example.com/user",
			wantCode: http.StatusOK,
		},
		{
			name:     "resource missing acct:",
			resource: "user@example.com",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "resource missing",
			resource: "",
			wantCode: http.StatusBadRequest,
		},
		{
			name:            "invalid method",
			resource:        "acct:user@example.com",
			wantCode:        http.StatusMethodNotAllowed,
			alternateMethod: http.MethodPost,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cfg := config.NewConfig()
			l := log.NewLogger(&strings.Builder{}, cfg)

			ctx = log.WithLogger(ctx, l)

			// Create a new request
			r, _ := http.NewRequestWithContext(
				ctx,
				tc.alternateMethod,
				"/.well-known/webfinger?resource="+tc.resource,
				http.NoBody,
			)

			// Create a new response
			w := httptest.NewRecorder()

			// Create a new handler
			h := handler.WebfingerHandler(fingers)

			// Serve the request
			h.ServeHTTP(w, r)

			require.Equal(t, tc.wantCode, w.Code)

			// If the status code is 200, check the response body
			if tc.wantCode == http.StatusOK {
				// Check the content type
				require.Equal(t, "application/jrd+json", w.Header().Get("Content-Type"))

				fingerWant := fingers[tc.resource]
				fingerGot := &webfingers.WebFinger{}

				// Decode the response body
				require.NoError(t, json.NewDecoder(w.Body).Decode(fingerGot))

				//  Sort links

				sort.Slice(fingerGot.Links, func(i, j int) bool {
					return fingerGot.Links[i].Rel < fingerGot.Links[j].Rel
				})

				sort.Slice(fingerWant.Links, func(i, j int) bool {
					return fingerWant.Links[i].Rel < fingerWant.Links[j].Rel
				})

				// Check the response body
				require.Equal(t, fingerWant, fingerGot)
			}
		})
	}
}

func BenchmarkWebfingerHandler(b *testing.B) {
	fingers, err := webfingers.NewWebFingers(
		webfingers.Resources{
			"user@example.com": {
				"prop1": "value1",
			},
		},
		nil,
	)
	require.NoError(b, err)

	h := handler.WebfingerHandler(fingers)
	r := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:user@example.com", http.NoBody)

	for range b.N {
		w := httptest.NewRecorder()

		h.ServeHTTP(w, r)

		require.Equal(b, http.StatusOK, w.Code)
	}
}
