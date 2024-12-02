package config

import "os"

// DBConfig содержит настройки для подключения к базе данных.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// AdminToken получается админ токен их переменной окружения
func AdminToken() string {
	token := os.Getenv("ADMIN_TOKEN")
	return token
}
