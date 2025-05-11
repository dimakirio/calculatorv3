package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dimakirio/calculatorv1/internal/auth"
)

func AuthMiddleware(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_login", claims.Login)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
