package repository

import (
	"Brocker-pet-project/internal/models"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_NewUser(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name        string
		username    string
		password    string
		mock        func()
		expected    *models.NewUserResponse
		expectError bool
	}{
		{
			name:     "successful user creation",
			username: "testuser",
			password: "testpass",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username"}).
					AddRow(1, "testuser")
				mock.ExpectQuery(`INSERT INTO users \(username,password\) VALUES \(\$1,\$2\) RETURNING id,username`).
					WithArgs("testuser", "testpass").
					WillReturnRows(rows)
			},
			expected: &models.NewUserResponse{
				Id:       1,
				Username: "testuser",
			},
			expectError: false,
		},
		{
			name:     "database error",
			username: "testuser",
			password: "testpass",
			mock: func() {
				mock.ExpectQuery(`INSERT INTO users \(username,password\) VALUES \(\$1,\$2\) RETURNING id,username`).
					WithArgs("testuser", "testpass").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "scan error - missing columns",
			username: "testuser",
			password: "testpass",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1) // missing username
				mock.ExpectQuery(`INSERT INTO users \(username,password\) VALUES \(\$1,\$2\) RETURNING id,username`).
					WithArgs("testuser", "testpass").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.NewUser(tt.username, tt.password)

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetUserByUsername(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name        string
		username    string
		password    string
		mock        func()
		expected    *models.User
		expectError bool
	}{
		{
			name:     "successful get user",
			username: "testuser",
			password: "testpass",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "password"}).
					AddRow(1, "testuser", "testpass")
				mock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
					WithArgs("testuser", "testpass").
					WillReturnRows(rows)
			},
			expected: &models.User{
				Id:       1,
				Username: "testuser",
				Password: "testpass",
			},
			expectError: false,
		},
		{
			name:     "user not found",
			username: "testuser",
			password: "testpass",
			mock: func() {
				mock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
					WithArgs("testuser", "testpass").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "database error",
			username: "testuser",
			password: "testpass",
			mock: func() {
				mock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
					WithArgs("testuser", "testpass").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "scan error - missing columns",
			username: "testuser",
			password: "testpass",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "username"}). // missing password
											AddRow(1, "testuser")
				mock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
					WithArgs("testuser", "testpass").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.GetUserByUsername(tt.username, tt.password)

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
