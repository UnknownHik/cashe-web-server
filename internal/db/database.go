package db

import (
	"database/sql"
	"fmt"
	"os"

	"cache-web-server/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDb создает соединение с базой данных.
func InitDb() (*sql.DB, error) {
	// Конфиг базы данных
	cfg := config.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
	}

	// Формируем строку подключения к БД
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение: %w", err)
	}

	// Проверяем, что соединение рабочее
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	// Создаем таблицы
	if err := CreateTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка при создании таблиц: %w", err)
	}

	return db, nil
}

// CreateTables создаёт необходимые таблицы в базе данных.
func CreateTables(db *sql.DB) error {
	// Запросы на создание таблиц
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL,
    		token TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			mime VARCHAR(50),
    		has_file BOOLEAN,
			public BOOLEAN DEFAULT FALSE,
			grant_login TEXT[],
    		owner VARCHAR(255),
			created TIMESTAMP DEFAULT NOW(),
			file BYTEA
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка при создании таблицы: %w", err)
		}
	}

	return nil
}
