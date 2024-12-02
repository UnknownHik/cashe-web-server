package main

import (
	"log"

	"cache-web-server/internal/db"
	"cache-web-server/internal/transport"

	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env: %v", err)
	}

	// Подключаемся к базе данных
	db, err := db.InitDb()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Получаем порт и запускаем сервер
	port := transport.GetPort()
	transport.StartServer(port, db)

}
