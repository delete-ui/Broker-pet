package repository

import (
	"Brocker-pet-project/internal/models"
	"context"
	"database/sql"
	"errors"
	"github.com/redis/go-redis/v9"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock
}

func setupMockRedis() (*redis.Client, redismock.ClientMock) {
	client, mock := redismock.NewClientMock()
	return client, mock
}

func TestDealRepository_CreateNewDeal(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	redisClient, _ := setupMockRedis()

	repo := NewDealRepository(db, redisClient)

	tests := []struct {
		name        string
		title       string
		expenses    float64
		profit      float64
		mock        func()
		expected    *models.Deal
		expectError bool
	}{
		{
			name:     "successful creation",
			title:    "Test Deal",
			expenses: 100,
			profit:   200,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
					AddRow(1, "Test Deal", 100, 200, "not processed")
				mock.ExpectQuery(`INSERT INTO transactions`).
					WithArgs("Test Deal", 100.0, 200.0, "not processed").
					WillReturnRows(rows)
			},
			expected: &models.Deal{
				Id:       1,
				Title:    "Test Deal",
				Expenses: 100,
				Profit:   200,
				Status:   "not processed",
			},
			expectError: false,
		},
		{
			name:     "database error",
			title:    "Test Deal",
			expenses: 100,
			profit:   200,
			mock: func() {
				mock.ExpectQuery(`INSERT INTO transactions`).
					WithArgs("Test Deal", 100.0, 200.0, "not processed").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.CreateNewDeal(tt.title, tt.expenses, tt.profit)

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDealRepository_GetAllNotProcessedDeals(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	redisClient, _ := setupMockRedis()

	repo := NewDealRepository(db, redisClient)

	tests := []struct {
		name        string
		mock        func()
		expected    *[]models.Deal
		expectError bool
	}{
		{
			name: "successful fetch",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
					AddRow(1, "Deal 1", 100, 200, "not processed").
					AddRow(2, "Deal 2", 150, 300, "not processed")
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
					WithArgs("not processed").
					WillReturnRows(rows)
			},
			expected: &[]models.Deal{
				{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "not processed"},
				{Id: 2, Title: "Deal 2", Expenses: 150, Profit: 300, Status: "not processed"},
			},
			expectError: false,
		},
		{
			name: "database error",
			mock: func() {
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
					WithArgs("not processed").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.GetAllNotProcessedDeals(context.Background())

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDealRepository_MarkTransactionAsProcessed(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	redisClient, redisMock := setupMockRedis()

	repo := NewDealRepository(db, redisClient)

	tests := []struct {
		name        string
		id          int64
		mock        func()
		redisMock   func()
		expected    *models.Deal
		expectError bool
	}{
		{
			name: "successful mark as processed",
			id:   1,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
					AddRow(1, "Deal 1", 100, 200, "processed")
				mock.ExpectQuery(`UPDATE transactions`).
					WithArgs("processed", int64(1)).
					WillReturnRows(rows)
			},
			redisMock: func() {
				redisMock.ExpectDel("notProcessedDeals:all", "processedDeals:all", "allDeals:get").SetVal(1)
			},
			expected: &models.Deal{
				Id:       1,
				Title:    "Deal 1",
				Expenses: 100,
				Profit:   200,
				Status:   "processed",
			},
			expectError: false,
		},
		{
			name: "database error",
			id:   1,
			mock: func() {
				mock.ExpectQuery(`UPDATE transactions`).
					WithArgs("processed", int64(1)).
					WillReturnError(errors.New("database error"))
			},
			redisMock:   func() {},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			tt.redisMock()

			result := repo.MarkTransactionAsProcessed(tt.id)

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
				assert.NoError(t, redisMock.ExpectationsWereMet())
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDealRepository_GetAllProcessedDeals(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	redisClient, _ := setupMockRedis()

	repo := NewDealRepository(db, redisClient)

	tests := []struct {
		name        string
		mock        func()
		expected    *[]models.Deal
		expectError bool
	}{
		{
			name: "successful fetch",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
					AddRow(1, "Deal 1", 100, 200, "processed").
					AddRow(2, "Deal 2", 150, 300, "processed")
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
					WithArgs("processed").
					WillReturnRows(rows)
			},
			expected: &[]models.Deal{
				{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "processed"},
				{Id: 2, Title: "Deal 2", Expenses: 150, Profit: 300, Status: "processed"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.GetAllProcessedDeals(context.Background())
			assert.Equal(t, tt.expected, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDealRepository_GetAllDeals(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	redisClient, _ := setupMockRedis()

	repo := NewDealRepository(db, redisClient)

	tests := []struct {
		name        string
		mock        func()
		expected    *[]models.Deal
		expectError bool
	}{
		{
			name: "successful fetch",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
					AddRow(1, "Deal 1", 100, 200, "processed").
					AddRow(2, "Deal 2", 150, 300, "not processed")
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE id!=0`).
					WillReturnRows(rows)
			},
			expected: &[]models.Deal{
				{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "processed"},
				{Id: 2, Title: "Deal 2", Expenses: 150, Profit: 300, Status: "not processed"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.GetAllDeals(context.Background())
			assert.Equal(t, tt.expected, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
