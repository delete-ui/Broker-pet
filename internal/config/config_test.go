package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader(t *testing.T) {
	configContent := `
env: "test"
server:
  host: "localhost"
  port: "8080"
worker:
  processedTimeOut: 10s
postgres:
  host: "db.localhost"
  port: "5432"
  user: "testuser"
  password: "testpass"
  dbname: "testdb"
  sslmode: "disable"
redis:
  address: "redis.localhost:6379"
jwt:
  token: "testtoken"
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yaml")

	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	require.NoError(t, err)

	configName := "test_config"

	tests := []struct {
		name    string
		want    *Config
		wantErr bool
	}{
		{
			name: "successful config load",
			want: &Config{
				Env: "test",
				Server: Server{
					Host: "localhost",
					Port: "8080",
				},
				Worker: Worker{
					ProcessedTimeOut: 10 * time.Second,
				},
				Postgres: Postgres{
					Host:     "db.localhost",
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					DBName:   "testdb",
					SSLMode:  "disable",
				},
				Redis: Redis{
					Address: "redis.localhost:6379",
				},
				Jwt: Jwt{
					Token: "testtoken",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldWd)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			got, err := ConfigLoader(configName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigLoader_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		configName  string
		expectError bool
	}{
		{
			name:        "non-existent config file",
			configName:  "nonexistent_config",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConfigLoader(tt.configName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
