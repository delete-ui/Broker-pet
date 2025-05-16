package database

import (
	"Brocker-pet-project/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		Postgres: config.Postgres{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "1106",
			DBName:   "pet_project",
			SSLMode:  "disable",
		},
	}

	t.Run("InitDB should initialize database connection", func(t *testing.T) {
		InitDB(cfg)

		// Проверяем, что DB не nil
		assert.NotNil(t, DB, "DB should not be nil after initialization")

		// Проверяем, что соединение действительно работает
		err := DB.Ping()
		assert.NoError(t, err, "Should be able to ping database")
	})

	t.Run("ReturnDB should return initialized DB instance", func(t *testing.T) {
		db := ReturnDB()

		assert.Equal(t, DB, db, "ReturnDB should return the same DB instance")
		assert.NotNil(t, db, "Returned DB should not be nil")
	})

	t.Run("CloseDB should close database connection", func(t *testing.T) {
		// Инициализируем соединение если еще не инициализировано
		if DB == nil {
			InitDB(cfg)
		}

		err := CloseDB()
		require.NoError(t, err, "Should close database without error")

		// Проверяем, что соединение действительно закрыто
		err = DB.Ping()
		assert.Error(t, err, "Ping should fail after closing connection")
	})
}
