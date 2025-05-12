package middleware

import (
	jwt2 "Brocker-pet-project/pkg/jwt"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		_, err := jwt2.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
