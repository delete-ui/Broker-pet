package handlers

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"Brocker-pet-project/pkg/jwt"
	"bytes"
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

func TestUserHandler_NewUserPost_Success(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Test data
	newUser := models.User{
		Username: "testuser",
		Password: "testpass",
	}
	expectedResponse := models.NewUserResponse{
		Id:       1,
		Username: "testuser",
	}

	// Mock expectations
	dbMock.ExpectQuery(`INSERT INTO users`).
		WithArgs(newUser.Username, newUser.Password).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(expectedResponse.Id, expectedResponse.Username))

	// Create request
	body, _ := json.Marshal(newUser)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.NewUserPost(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())

	var response models.NewUserResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Check logs
	assert.Equal(t, 1, observedLogs.FilterMessage("User post request successfully handled").Len())
}

func TestUserHandler_NewUserPost_InvalidMethod(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Create request with wrong method
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.NewUserPost(w, req)

	// Verify
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUserHandler_NewUserPost_InvalidContentType(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Create request with wrong content type
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Call handler
	handler.NewUserPost(w, req)

	// Verify
	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

func TestUserHandler_NewUserPost_DecodeError(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.NewUserPost(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 1, observedLogs.FilterMessage("Error decoding user").Len())
}

func TestUserHandler_NewUserPost_DBError(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Test data
	newUser := models.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Mock expectations - return error
	dbMock.ExpectQuery(`INSERT INTO users`).
		WithArgs(newUser.Username, newUser.Password).
		WillReturnError(errors.New("database error"))

	// Create request
	body, _ := json.Marshal(newUser)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.NewUserPost(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Equal(t, 1, observedLogs.FilterMessage("Error creating new user").Len())
}

func TestUserHandler_LoginIn_Success(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Test data
	loginUser := models.User{
		Username: "testuser",
		Password: "testpass",
	}
	dbUser := models.User{
		Id:       1,
		Username: "testuser",
		Password: "testpass",
	}

	// Исправленный запрос - должен соответствовать тому, что в обработчике
	dbMock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
		WithArgs(loginUser.Username, loginUser.Password).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password"}).
			AddRow(dbUser.Id, dbUser.Username, dbUser.Password))

	// Create request
	body, _ := json.Marshal(loginUser)
	req := httptest.NewRequest(http.MethodGet, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call handler
	handler.LoginIn(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])

	// Check logs
	assert.Equal(t, 1, observedLogs.FilterMessage("User get request successfully handled").Len())
}

func TestUserHandler_LoginIn_UserNotFound(t *testing.T) {
	// Setup
	db, dbMock := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Test data
	loginUser := models.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Исправленный запрос
	dbMock.ExpectQuery(`SELECT \* FROM users WHERE username=\$1 AND password=\$2`).
		WithArgs(loginUser.Username, loginUser.Password).
		WillReturnError(sql.ErrNoRows)

	// Create request
	body, _ := json.Marshal(loginUser)
	req := httptest.NewRequest(http.MethodGet, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call handler
	handler.LoginIn(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Equal(t, 1, observedLogs.FilterMessage("Error getting user by username").Len())
}

func TestUserHandler_LoginIn_InvalidMethod(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	logger := zap.NewNop()
	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Create request with wrong method
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.LoginIn(w, req)

	// Verify
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUserHandler_LoginIn_DecodeError(t *testing.T) {
	// Setup
	db, _ := setupMockDB(t)
	defer db.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	logger := zap.New(observedZapCore)

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo, logger)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodGet, "/login", bytes.NewReader([]byte("{invalid}")))
	w := httptest.NewRecorder()

	// Call handler
	handler.LoginIn(w, req)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 1, observedLogs.FilterMessage("Error decoding user").Len())
}

// Helper to mock jwt.GenerateToken
var jwtGenerate = jwt.GenerateToken
