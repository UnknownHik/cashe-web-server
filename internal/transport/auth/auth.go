package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"cache-web-server/internal/models"
	"cache-web-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler обрабатывает POST запрос для регистрации нового пользователя
func RegisterHandler(db *sql.DB, adminToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.ErrorResponse(w, 405)
			return
		}

		// Проверяем токен администратора
		if r.Header.Get("Authorization") != adminToken {
			utils.ErrorResponse(w, 403)
			return
		}

		var req models.User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.ErrorResponse(w, 400)
			return
		}

		// Проверяем формат логина
		if !regexp.MustCompile(`^[a-zA-Z0-9]{8,}$`).MatchString(req.Login) {
			utils.ErrorResponse(w, 400)
			return
		}

		// Проверяем формат пароля
		if !validPass(req.Pswd) {
			utils.ErrorResponse(w, 400)
			return
		}

		// Хэшируем пароль
		hashedPassword, err := hashPassword(req.Pswd)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		// Добавляем пользователя в базу
		var userID int
		query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
		err = db.QueryRow(query, req.Login, hashedPassword).Scan(&userID)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		utils.ActResponse(w, "login", req.Login)
	}
}

// validPass проверяет пароль на соответствие требованиям
func validPass(pswd string) bool {
	if len(pswd) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pswd)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(pswd)
	hasNumber := regexp.MustCompile(`\d`).MatchString(pswd)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(pswd)
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// hashPassword хэширует пароль
func hashPassword(pswd string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pswd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка при хэшировании пароля: %w", err)
	}
	return string(hashedPass), nil
}

// AuthHandler обрабатывает POST запрос для аутентификации пользователя
func AuthHandler(db *sql.DB, secretKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.ErrorResponse(w, 405)
			return
		}

		var req models.User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.ErrorResponse(w, 400)
			return
		}

		// Ищем пользователя в базе
		var hashedPassword string
		var userID int
		var token *string
		query := `SELECT id, password, token FROM users WHERE login = $1`
		err := db.QueryRow(query, req.Login).Scan(&userID, &hashedPassword, &token)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		// Сравниваем хэш пароля
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Pswd)); err != nil {
			utils.ErrorResponse(w, 401)
			return
		}

		// Генерируем токен
		tokenString, err := generateJWT(req.Login, secretKey)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		// Обновляем токен в базе данных для этого пользователя
		query = `UPDATE users SET token = $1 WHERE id = $2`
		_, err = db.Exec(query, tokenString, userID)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		utils.ActResponse(w, "token", tokenString)
	}
}

// generateJWT создает JWT для пользователя
func generateJWT(login string, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"login": login,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
