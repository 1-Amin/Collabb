package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/collabb/backend/internal/auth"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		var tokenStr string
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else {
			// WebSocket upgrade passes token as query param
			tokenStr = r.URL.Query().Get("token")
		}
		if tokenStr == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
