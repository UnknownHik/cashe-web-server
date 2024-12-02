package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"cache-web-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(db *sql.DB, JWTSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок с токеном
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.ErrorResponse(w, 401)
				return
			}

			// Разбираем токен
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(JWTSecret), nil
			})

			if err != nil || !token.Valid {
				utils.ErrorResponse(w, 401)
				return
			}

			// Проверяем существование пользователя в БД
			var userID int
			query := `SELECT id FROM users WHERE login = $1`
			err = db.QueryRow(query, claims["login"]).Scan(&userID)
			if err != nil {
				utils.ErrorResponse(w, 401)
				return
			}

			// Добавляем пользователя в контекст
			ctx := context.WithValue(r.Context(), "login", claims["login"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
