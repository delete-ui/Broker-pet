package redis_test

import (
	"Brocker-pet-project/internal/config"
	"Brocker-pet-project/pkg/redis"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRedisClient(t *testing.T) {
	// Подготовка тестовых данных
	testCases := []struct {
		name     string
		cfg      *config.Config
		expected string
	}{
		{
			name: "successful client creation",
			cfg: &config.Config{
				Redis: config.Redis{
					Address: "localhost:6379",
				},
			},
			expected: "localhost:6379",
		},
		{
			name: "empty address - should use default",
			cfg: &config.Config{
				Redis: config.Redis{
					Address: "",
				},
			},
			expected: "localhost:6379", // Изменено с "" на "localhost:6379"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Вызов тестируемой функции
			client := redis.NewRedisClient(tc.cfg)

			// Проверка, что клиент не nil
			assert.NotNil(t, client, "Redis client should not be nil")

			// Проверка адреса подключения
			assert.Equal(t, tc.expected, client.Options().Addr, "Redis address should match config")

			// Дополнительная проверка: тестовый запрос к Redis (если нужно)
			if tc.cfg.Redis.Address != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				_, err := client.Ping(ctx).Result()
				if err != nil {
					t.Logf("Warning: could not ping Redis server: %v", err)
				}
			}
		})
	}
}
