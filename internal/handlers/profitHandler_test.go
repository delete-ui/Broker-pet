package handlers

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProfitHandler_AllClearProfitGET_Success(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	profitRepo := repository.NewProfitRepository(db)
	handler := NewProfitHandler(profitRepo, logger)

	// Test data
	expectedProfits := []models.ProfitSQLDeal{
		{Id: 1, DealId: 1, AllProfit: 100},
		{Id: 2, DealId: 2, AllProfit: 200},
	}

	// Mock expectations
	rows := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
		AddRow(expectedProfits[0].Id, expectedProfits[0].DealId, expectedProfits[0].AllProfit).
		AddRow(expectedProfits[1].Id, expectedProfits[1].DealId, expectedProfits[1].AllProfit)

	dbMock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit`).
		WillReturnRows(rows)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/profits", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllClearProfitGET(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())

	var response []models.ProfitSQLDeal
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedProfits, response)
}

func TestProfitHandler_AllClearProfitGET_WrongMethod(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	profitRepo := repository.NewProfitRepository(db)
	handler := NewProfitHandler(profitRepo, logger)

	// Create request with wrong method
	req := httptest.NewRequest(http.MethodPost, "/profits", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllClearProfitGET(w, req)

	// Verify
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestProfitHandler_AllClearProfitGET_DBError(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	profitRepo := repository.NewProfitRepository(db)
	handler := NewProfitHandler(profitRepo, logger)

	// Mock expectations
	dbMock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit`).
		WillReturnError(sql.ErrNoRows)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/profits", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.AllClearProfitGET(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestProfitHandler_AllClearProfitGET_EncodeError(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	// Create logger that will observe logs
	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore)

	profitRepo := repository.NewProfitRepository(db)
	handler := NewProfitHandler(profitRepo, observedLogger)

	// Test data
	testProfits := []models.ProfitSQLDeal{
		{Id: 1, DealId: 1, AllProfit: 100},
	}

	// Mock database response
	rows := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
		AddRow(testProfits[0].Id, testProfits[0].DealId, testProfits[0].AllProfit)

	dbMock.ExpectQuery(`SELECT id, deals_id, all_profit FROM clear_profit`).
		WillReturnRows(rows)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/profits", nil)

	// Create response writer that will fail on Write
	w := &failingResponseWriter{
		header: make(http.Header),
	}

	// Call handler
	handler.AllClearProfitGET(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.statusCode, "Expected status code 500")
	assert.Equal(t, 1, observedLogs.Len(), "Expected 1 error log")
	if observedLogs.Len() > 0 {
		assert.Contains(t, observedLogs.All()[0].Message, "Error encoding response")
	}
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// failingResponseWriter fails on Write to simulate json encode error
type failingResponseWriter struct {
	header     http.Header
	statusCode int
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	// Return error to simulate encoding failure
	return 0, errors.New("forced write error")
}

func (w *failingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

type errorWriter struct{}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, assert.AnError
}

// Переменная для подмены json.NewEncoder в тестах
var jsonNewEncoder = func(w http.ResponseWriter) *json.Encoder {
	return json.NewEncoder(w)
}
