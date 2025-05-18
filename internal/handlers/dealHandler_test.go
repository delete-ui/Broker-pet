package handlers

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestDealHandler_NewDealPost_Success(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, redisMock := setupMockRedis()
	logger := zap.NewNop()

	dealRepo := repository.NewDealRepository(db, redisClient)
	handler := NewDealHandler(dealRepo, redisClient, logger)

	// Test data
	newDeal := models.Deal{
		Title:    "Test Deal",
		Expenses: 100,
		Profit:   200,
	}
	expectedDeal := models.Deal{
		Id:       1,
		Title:    "Test Deal",
		Expenses: 100,
		Profit:   200,
		Status:   "not processed",
	}

	// Mock expectations
	dbMock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(newDeal.Title, newDeal.Expenses, newDeal.Profit, "not processed").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
			AddRow(expectedDeal.Id, expectedDeal.Title, expectedDeal.Expenses, expectedDeal.Profit, expectedDeal.Status))

	redisMock.ExpectDel("notProcessedDeals:all", "processedDeals:all", "allDeals:get").SetVal(1)

	// Create request
	body, _ := json.Marshal(newDeal)
	req := httptest.NewRequest(http.MethodPost, "/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.NewDealPost(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())

	var response models.Deal
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, expectedDeal, response)
}

func TestDealHandler_AllProcessedDealsGet_Cached(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	redisClient, redisMock := setupMockRedis()
	logger := zap.NewNop()

	dealRepo := repository.NewDealRepository(db, redisClient)
	handler := NewDealHandler(dealRepo, redisClient, logger)

	// Test data
	cachedDeals := []models.Deal{
		{Id: 1, Title: "Deal 1", Status: "processed"},
		{Id: 2, Title: "Deal 2", Status: "processed"},
	}
	cachedData, _ := json.Marshal(cachedDeals)

	// Mock expectations
	redisMock.ExpectGet("processedDeals:all").SetVal(string(cachedData))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/deals/processed", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllProcessedDealsGet(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, redisMock.ExpectationsWereMet())

	var response []models.Deal
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, cachedDeals, response)
}

func TestDealHandler_AllNotProcessedDealsGet_NotCached(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, redisMock := setupMockRedis()
	logger := zap.NewNop()

	dealRepo := repository.NewDealRepository(db, redisClient)
	handler := NewDealHandler(dealRepo, redisClient, logger)

	// Test data
	deals := []models.Deal{
		{Id: 1, Title: "Deal 1", Status: "not processed"},
		{Id: 2, Title: "Deal 2", Status: "not processed"},
	}
	expectedJSON, _ := json.Marshal(deals)

	// Mock expectations
	redisMock.ExpectGet("notProcessedDeals:all").RedisNil()

	dbMock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
		WithArgs("not processed").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
			AddRow(deals[0].Id, deals[0].Title, deals[0].Expenses, deals[0].Profit, deals[0].Status).
			AddRow(deals[1].Id, deals[1].Title, deals[1].Expenses, deals[1].Profit, deals[1].Status))

	// Исправленная часть - используем точное значение JSON вместо mock.Anything
	redisMock.ExpectSet("notProcessedDeals:all", expectedJSON, 5*time.Minute).SetVal("OK")

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/deals/not-processed", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllNotProcessedDealsGet(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())

	var response []models.Deal
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, deals, response)
}

func TestDealHandler_AllDealsGet_WrongMethod(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	redisClient, _ := setupMockRedis()
	logger := zap.NewNop()

	dealRepo := repository.NewDealRepository(db, redisClient)
	handler := NewDealHandler(dealRepo, redisClient, logger)

	// Create request with wrong method
	req := httptest.NewRequest(http.MethodPost, "/deals", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllDealsGet(w, req)

	// Verify
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
