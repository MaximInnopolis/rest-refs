package http

import (
	"context"
	"net/http"
	"strings"
)

// RequestIDMiddleware adds requested endpoint to request context for further use
// This middleware extracts endpoint from request URL and stores it in the context
func (h *Handler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract endpoint information from request
		endpoint := r.URL.Path
		ctx := context.WithValue(r.Context(), "endpoint", endpoint)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireValidTokenMiddleware validates JWT from Authorization header
// This middleware checks if valid token is provided, extracts user ID from claims,
// and adds user ID to request context for further use
func (h *Handler) RequireValidTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")

		// Check if the Authorization header is missing
		if authHeader == "" {
			http.Error(w, "Не указан заголовок авторизации", http.StatusInternalServerError)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check if token is valid
		authenticated, claims, err := h.service.IsTokenValid(tokenString)
		if err != nil {
			http.Error(w, "Проблема на сервере", http.StatusInternalServerError)
			return
		}

		// If token is not valid, return unauthorized status
		if !authenticated {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["id"].(float64)
		if !ok {
			http.Error(w, "Ошибка токена", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "UserID", int(userID))
		// If token is valid pass request further
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
