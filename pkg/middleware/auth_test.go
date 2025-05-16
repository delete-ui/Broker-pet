package middleware

import (
	"Brocker-pet-project/pkg/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	// Генерируем валидный тестовый токен
	validToken, err := jwt.GenerateToken(1)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый обработчик, который будет вызван после middleware
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Создаем запрос с тестовым токеном
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Применяем middleware к тестовому обработчику
			middleware := AuthMiddleware(handler)
			middleware.ServeHTTP(rr, req)

			// Проверяем статус код
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
