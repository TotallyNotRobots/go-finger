package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/internal/config"
	"git.maronato.dev/maronato/finger/internal/log"
	"git.maronato.dev/maronato/finger/internal/server"
	"git.maronato.dev/maronato/finger/webfingers"
)

func getPortGenerator() func() int {
	lock := &sync.Mutex{}
	port := 8080

	return func() int {
		lock.Lock()
		defer lock.Unlock()

		port++

		return port
	}
}

func TestStartServer(t *testing.T) {
	t.Parallel()

	portGenerator := getPortGenerator()

	t.Run("starts and shuts down", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()

		cfg := config.NewConfig()
		l := log.NewLogger(&strings.Builder{}, cfg)

		ctx = log.WithLogger(ctx, l)

		// Use a new port
		cfg.Port = fmt.Sprint(portGenerator())

		// Start the server
		err := server.StartServer(ctx, cfg, nil)
		require.NoError(t, err)
	})

	t.Run("fails to start", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()

		cfg := config.NewConfig()
		l := log.NewLogger(&strings.Builder{}, cfg)

		ctx = log.WithLogger(ctx, l)

		// Use a new port
		cfg.Port = fmt.Sprint(portGenerator())

		// Use invalid host
		cfg.Host = "google.com"

		// Start the server
		err := server.StartServer(ctx, cfg, nil)
		require.Error(t, err)
	})

	t.Run("serves webfinger", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()

		cfg := config.NewConfig()
		l := log.NewLogger(&strings.Builder{}, cfg)

		ctx = log.WithLogger(ctx, l)

		// Use a new port
		cfg.Port = fmt.Sprint(portGenerator())

		resource := "acct:user@example.com"
		fingers := webfingers.WebFingers{
			resource: &webfingers.WebFinger{
				Subject: resource,
				Properties: map[string]string{
					"http://webfinger.net/rel/name": "John Doe",
				},
			},
		}

		go func() {
			// Start the server
			err := server.StartServer(ctx, cfg, fingers)
			assert.NoError(t, err)
		}()

		// Wait for the server to start
		time.Sleep(time.Millisecond * 50)

		// Create a new client
		c := http.Client{}

		// Create a new request
		r, _ := http.NewRequestWithContext(ctx,
			http.MethodGet,
			"http://"+cfg.GetAddr()+"/.well-known/webfinger?resource=acct:user@example.com",
			http.NoBody,
		)

		// Send the request
		resp, err := c.Do(r)
		require.NoError(t, err)

		defer resp.Body.Close()

		// Check the status code
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Check the response body
		fingerGot := &webfingers.WebFinger{}

		// Decode the response body
		require.NoError(t, json.NewDecoder(resp.Body).Decode(fingerGot))

		// Check the response body
		fingerWant := fingers[resource]

		require.Equal(t, fingerWant, fingerGot)
	})

	t.Run("serves healthcheck", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()

		cfg := config.NewConfig()
		l := log.NewLogger(&strings.Builder{}, cfg)

		ctx = log.WithLogger(ctx, l)

		// Use a new port
		cfg.Port = fmt.Sprint(portGenerator())

		go func() {
			// Start the server
			err := server.StartServer(ctx, cfg, nil)
			assert.NoError(t, err)
		}()

		// Wait for the server to start
		time.Sleep(time.Millisecond * 50)

		// Create a new client
		c := http.Client{}

		// Create a new request
		r, _ := http.NewRequestWithContext(ctx,
			http.MethodGet,
			"http://"+cfg.GetAddr()+"/healthz",
			http.NoBody,
		)

		// Send the request
		resp, err := c.Do(r)
		require.NoError(t, err)

		defer resp.Body.Close()

		// Check the status code
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
