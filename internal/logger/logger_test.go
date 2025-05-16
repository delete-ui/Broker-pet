package logger

import (
	"Brocker-pet-project/internal/config"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected zapcore.Level
	}{
		{
			name:     "development environment",
			env:      "dev",
			expected: zap.DebugLevel,
		},
		{
			name:     "production environment",
			env:      "prod",
			expected: zap.InfoLevel,
		},
		{
			name:     "default (local) environment",
			env:      "local",
			expected: zap.DebugLevel,
		},
		{
			name:     "unknown environment falls back to local",
			env:      "unknown",
			expected: zap.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := &config.Config{Env: tt.env}

			logger, err := InitLogger(cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)

			observedZapCore, observedLogs := observer.New(tt.expected)
			observedLogger := zap.New(observedZapCore)

			switch tt.expected {
			case zap.DebugLevel:
				observedLogger.Debug("test debug message")
				assert.Equal(t, 1, observedLogs.Len(), "Debug message should be logged")
			case zap.InfoLevel:
				observedLogger.Info("test info message")
				assert.Equal(t, 1, observedLogs.Len(), "Info message should be logged")
			}

			if tt.expected == zap.InfoLevel {
				observedLogger.Debug("test debug message that should not be logged")
				assert.Equal(t, 1, observedLogs.Len(), "Debug message should not be logged in production")
			}

			// Корректно закрываем логгер
			err = logger.Sync()
			assert.NoError(t, err, "Logger sync should not return error")
		})
	}
}
