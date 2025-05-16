package repository

import (
	"Brocker-pet-project/internal/models"
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProfitRepository_AddProfitById(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewProfitRepository(db)

	tests := []struct {
		name        string
		dealId      int64
		allProfit   float64
		mock        func()
		expected    *models.ProfitSQLDeal
		expectError bool
	}{
		{
			name:      "successful add profit",
			dealId:    1,
			allProfit: 100.50,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
					AddRow(1, 1, 100.50)
				mock.ExpectQuery(`INSERT INTO clear_profit \(deals_id,all_profit\) VALUES \(\$1,\$2\) RETURNING id, deals_id,all_profit`).
					WithArgs(int64(1), 100.50).
					WillReturnRows(rows)
			},
			expected: &models.ProfitSQLDeal{
				Id:        1,
				DealId:    1,
				AllProfit: 100.50,
			},
			expectError: false,
		},
		{
			name:      "database error",
			dealId:    1,
			allProfit: 100.50,
			mock: func() {
				mock.ExpectQuery(`INSERT INTO clear_profit \(deals_id,all_profit\) VALUES \(\$1,\$2\) RETURNING id, deals_id,all_profit`).
					WithArgs(int64(1), 100.50).
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.AddProfitById(tt.dealId, tt.allProfit)

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProfitRepository_GetAllProfitInfo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewProfitRepository(db)

	tests := []struct {
		name        string
		mock        func()
		expected    *[]models.ProfitSQLDeal
		expectError bool
	}{
		{
			name: "successful get all profits",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
					AddRow(1, 1, 100.50).
					AddRow(2, 2, 200.75)
				mock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit;`).
					WillReturnRows(rows)
			},
			expected: &[]models.ProfitSQLDeal{
				{Id: 1, DealId: 1, AllProfit: 100.50},
				{Id: 2, DealId: 2, AllProfit: 200.75},
			},
			expectError: false,
		},
		{
			name: "database error",
			mock: func() {
				mock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit;`).
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "empty result",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"})
				mock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit;`).
					WillReturnRows(rows)
			},
			expected:    &[]models.ProfitSQLDeal{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			result := repo.GetAllProfitInfo(context.Background())

			if tt.expectError {
				assert.Nil(t, result)
			} else {
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					if len(*tt.expected) == 0 {
						assert.Empty(t, *result)
					} else {
						assert.Equal(t, tt.expected, result)
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
