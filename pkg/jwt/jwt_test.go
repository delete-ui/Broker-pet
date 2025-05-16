package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndValidateToken(t *testing.T) {

	userID := int64(123)

	tokenString, err := GenerateToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(userID), claims["user_id"]) // jwt библиотека конвертирует числа в float64

	exp, err := claims.GetExpirationTime()
	assert.NoError(t, err)
	assert.True(t, exp.Time.After(time.Now()))
}

func TestValidateToken_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		tokenString string
		expectError bool
	}{
		{
			name:        "empty token",
			tokenString: "",
			expectError: true,
		},
		{
			name:        "invalid token",
			tokenString: "invalid.token.string",
			expectError: true,
		},
		{
			name:        "expired token",
			tokenString: generateExpiredToken(t),
			expectError: true,
		},
		{
			name:        "wrong signing method",
			tokenString: generateTokenWithWrongSigningMethod(t),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := ValidateToken(tc.tokenString)
			if tc.expectError {
				assert.Error(t, err)
				if token != nil {
					assert.False(t, token.Valid)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			}
		})
	}
}

func generateExpiredToken(t *testing.T) string {
	claims := jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(-24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	assert.NoError(t, err)
	return tokenString
}

func generateTokenWithWrongSigningMethod(t *testing.T) string {
	// Генерируем ECDSA ключ для теста
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	claims := jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	assert.NoError(t, err)
	return tokenString
}
