package fingerreader_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"git.maronato.dev/maronato/finger/internal/config"
	"git.maronato.dev/maronato/finger/internal/fingerreader"
	"git.maronato.dev/maronato/finger/internal/log"
	"git.maronato.dev/maronato/finger/webfingers"
)

func newTempFile(t *testing.T, content string) (name string, remove func()) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "finger-test")
	require.NoError(t, err)

	_, err = f.WriteString(content)
	require.NoError(t, err)

	return f.Name(), func() {
		err = os.Remove(f.Name())
		require.NoError(t, err)
	}
}

func TestNewFingerReader(t *testing.T) {
	t.Parallel()

	require.NotNil(t, fingerreader.NewFingerReader())
}

func TestFingerReader_ReadFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		urnsContent    string
		fingersContent string
		useURNFile     bool
		useFingerFile  bool
		wantErr        bool
	}{
		{
			name:           "reads files",
			urnsContent:    "name: https://schema/name\nprofile: https://schema/profile",
			fingersContent: "user@example.com:\n  name: John Doe",
			useURNFile:     true,
			useFingerFile:  true,
			wantErr:        false,
		},
		{
			name:           "errors on missing URNs file",
			urnsContent:    "invalid",
			fingersContent: "user@example.com:\n  name: John Doe",
			useURNFile:     false,
			useFingerFile:  true,
			wantErr:        true,
		},
		{
			name:           "errors on missing fingers file",
			urnsContent:    "name: https://schema/name\nprofile: https://schema/profile",
			fingersContent: "invalid",
			useFingerFile:  false,
			useURNFile:     true,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.NewConfig()

			urnsFileName, urnsCleanup := newTempFile(t, tc.urnsContent)
			defer urnsCleanup()

			fingersFileName, fingersCleanup := newTempFile(t, tc.fingersContent)
			defer fingersCleanup()

			if !tc.useURNFile {
				cfg.URNPath = "invalid"
			} else {
				cfg.URNPath = urnsFileName
			}

			if !tc.useFingerFile {
				cfg.FingerPath = "invalid"
			} else {
				cfg.FingerPath = fingersFileName
			}

			f := fingerreader.NewFingerReader()

			err := f.ReadFiles(cfg)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, []byte(tc.urnsContent), f.URNSFile)
				require.Equal(t, []byte(tc.fingersContent), f.FingersFile)
			}
		})
	}
}

func TestReadFingerFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		urnsContent    string
		fingersContent string
		wantURN        webfingers.URNAliases
		wantFinger     webfingers.Resources
		returns        webfingers.WebFingers
		wantErr        bool
	}{
		{
			name:           "reads files",
			urnsContent:    "name: https://schema/name\nprofile: https://schema/profile",
			fingersContent: "user@example.com:\n  name: John Doe",
			wantURN: webfingers.URNAliases{
				"name":    "https://schema/name",
				"profile": "https://schema/profile",
			},
			wantFinger: webfingers.Resources{
				"user@example.com": {
					"name": "John Doe",
				},
			},
			returns: webfingers.WebFingers{
				"acct:user@example.com": {
					Subject: "acct:user@example.com",
					Properties: map[string]string{
						"https://schema/name": "John Doe",
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "uses custom URNs",
			urnsContent:    "favorite_food: https://schema/favorite_food",
			fingersContent: "user@example.com:\n  favorite_food: Apple",
			wantURN: webfingers.URNAliases{
				"favorite_food": "https://schema/favorite_food",
			},
			wantFinger: webfingers.Resources{
				"user@example.com": {
					"https://schema/favorite_food": "Apple",
				},
			},
			wantErr: false,
		},
		{
			name:           "errors on invalid URNs file",
			urnsContent:    "invalid",
			fingersContent: "user@example.com:\n  name: John Doe",
			wantErr:        true,
		},
		{
			name:           "errors on invalid fingers file",
			urnsContent:    "name: https://schema/name\nprofile: https://schema/profile",
			fingersContent: "invalid",
			wantErr:        true,
		},
		{
			name:           "errors on invalid URNs values",
			urnsContent:    "name: invalid",
			fingersContent: "user@example.com:\n  name: John Doe",
			wantErr:        true,
		},
		{
			name:           "errors on invalid fingers values",
			urnsContent:    "name: https://schema/name\nprofile: https://schema/profile",
			fingersContent: "invalid:\n  name: John Doe",
			wantErr:        true,
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

			f := fingerreader.NewFingerReader()

			f.FingersFile = []byte(tc.fingersContent)
			f.URNSFile = []byte(tc.urnsContent)

			got, err := f.ReadFingerFile(ctx)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tc.returns != nil {
					require.Equal(t, tc.returns, got)
				}
			}
		})
	}
}
