package transport

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"cache-web-server/config"
	"cache-web-server/internal/transport/auth"
	"cache-web-server/internal/transport/auth/middleware"
	"cache-web-server/internal/transport/rest"

	"github.com/go-chi/chi/v5"
)

// GetPort получает порт из переменной окружения или использует 8080 по умолчанию
func GetPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// StartServer запускает веб-сервер на указанном порту
func StartServer(port string, db *sql.DB) {
	r := chi.NewRouter()

	// Получаем adminToken
	adminToken := config.AdminToken()
	if adminToken == "" {
		log.Fatal("ADMIN_TOKEN не установлен в .env")
	}

	// Получаем JWT_SECRET из переменной окружения
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET не установлен в .env")
	}

	// Подключаем middleware для авторизации
	authMiddleware := middleware.AuthMiddleware(db, JWTSecret)

	// Обработчики для регистрации и аутентификации пользователя
	r.Post("/api/register", auth.RegisterHandler(db, adminToken))
	r.Post("/api/auth", auth.AuthHandler(db, JWTSecret))

	// Обработчики для работы с документами, требующие авторизации
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/api/docs", rest.UploadHandler(db))
		r.Get("/api/docs", rest.ListDocsHandler(db))
		r.Head("/api/docs", rest.ListDocsHandler(db))
		r.Get("/api/docs/{id}", rest.GetDocHandler(db))
		r.Head("/api/docs/{id}", rest.GetDocHandler(db))
		r.Delete("/api/docs/{id}", rest.DeleteDocHandler(db))
		r.Delete("/api/auth/{token}", rest.LogoutHandler(db))

	})

	log.Printf("Сервер запущен на порту: %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}
