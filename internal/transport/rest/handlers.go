package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cache-web-server/internal/models"
	"cache-web-server/internal/utils"

	"github.com/go-chi/chi/v5"
)

// UploadHandler обрабатывает загрузку нового документа
func UploadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.ErrorResponse(w, 405)
			return
		}

		// Извлекаем логин текущего пользователя из контекста
		login := r.Context().Value("login").(string)

		// Парсим multipart-запрос с лимитом в 10 Мб
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.ErrorResponse(w, 400)
			return
		}

		// Достаем метаданные
		metaData := r.FormValue("meta")
		if metaData == "" {
			utils.ErrorResponse(w, 400)
			return
		}

		var meta models.Meta
		if err := json.Unmarshal([]byte(metaData), &meta); err != nil {
			utils.ErrorResponse(w, 400)
			return
		}

		if meta.Token == "" || meta.Name == "" {
			utils.ErrorResponse(w, 400)
			return
		}

		// Достаем json
		jsonData := r.FormValue("json")
		var jsonParsed map[string]interface{}
		if jsonData != "" {
			if err := json.Unmarshal([]byte(jsonData), &jsonParsed); err != nil {
				utils.ErrorResponse(w, 400)
				return
			}
		}

		var docID string
		// Достаем файл
		if meta.File {
			file, _, err := r.FormFile("file")
			if err != nil {
				utils.ErrorResponse(w, 400)
				return
			}
			defer file.Close()

			// Читаем содержимое файла в память
			fileData, err := io.ReadAll(file)
			if err != nil {
				utils.ErrorResponse(w, 500)
				return
			}

			// Вставляем данные в таблицу
			query := `INSERT INTO documents (id, name, mime, has_file, public, grant_login, owner, file)
				  VALUES ($1, $2, $3, TRUE, $4, $5, $6, $7) RETURNING id`
			err = db.QueryRow(query, meta.Token, meta.Name, meta.Mime, meta.Public, meta.Grant, login, fileData).Scan(&docID)
			if err != nil {
				fmt.Println(err)
				utils.ErrorResponse(w, 500)
				return
			}

			utils.UploadResponse(w, jsonParsed, meta.Name)
		} else {
			query := `INSERT INTO documents (id, name, mime, has_file, public, grant_login, owner)
				  VALUES ($1, $2, $3, TRUE, $4, $5, $6) RETURNING id`
			err := db.QueryRow(query, meta.Token, meta.Name, meta.Mime, meta.Public, meta.Grant, login).Scan(&docID)
			if err != nil {
				fmt.Println(err)
				utils.ErrorResponse(w, 500)
				return
			}

			utils.UploadResponse(w, jsonParsed, meta.Name)
		}
	}
}

// ListDocsHandler обрабатывает получение списка документов
func ListDocsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			utils.ErrorResponse(w, 405)
			return
		}

		// Извлекаем логин текущего пользователя из контекста
		userLogin := r.Context().Value("login").(string)

		// Читаем параметры запроса
		login := r.URL.Query().Get("login")
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		limitStr := r.URL.Query().Get("limit")

		// Составляем базовый SQL-запрос для получения документов
		query := `SELECT id, name, mime, has_file, public, created, grant_login FROM documents`
		shardQuery := []string{}
		params := []interface{}{}

		// Если передан логин, фильтруем документы по нему, иначе показываем только свои
		if login != "" {
			shardQuery = append(shardQuery, fmt.Sprintf("login = $%d", len(params)+1))
			params = append(params, login)
		} else {
			shardQuery = append(shardQuery, fmt.Sprintf("owner = $%d", len(params)+1))
			params = append(params, userLogin)
		}

		// Если передан параметр key и value, добавляем фильтрацию по ним
		if key != "" && value != "" {
			shardQuery = append(shardQuery, fmt.Sprintf("%s = $%d", key, len(params)+1))
			params = append(params, value)
		}

		if len(shardQuery) > 0 {
			query += " WHERE " + strings.Join(shardQuery, " AND ") + " ORDER BY name, created"
		}

		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				utils.ErrorResponse(w, 400) // Неверный лимит
				return
			}
			query += " LIMIT $" + strconv.Itoa(len(params)+1)
			params = append(params, limit)
		}

		// Выполняем запрос к базе данных
		rows, err := db.Query(query, params...)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}
		defer rows.Close()

		var docs []models.Document
		// Читаем строки из результата запроса
		for rows.Next() {
			var doc models.Document
			var grant string
			if err := rows.Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public, &doc.Created, &grant); err != nil {
				fmt.Println(err)
				utils.ErrorResponse(w, 500)
				return
			}

			doc.Grant = strings.Split(grant, ",")
			docs = append(docs, doc)
		}

		utils.DataResponse(w, docs)
	}
}

// GetDocHandler обрабатывает получение одного документа
func GetDocHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			utils.ErrorResponse(w, 405)
			return
		}

		// Получаем ID документа из параметров
		id := chi.URLParam(r, "id")
		if id == "" {
			utils.ErrorResponse(w, 400)
			return
		}

		// Читаем документ из базы
		query := `SELECT id, name, mime, has_file, public, created, grant_login, file FROM documents WHERE id = $1`
		var doc models.Document
		var file []byte
		var grant string
		err := db.QueryRow(query, id).Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public, &doc.Created, &grant, &file)
		if err != nil {
			fmt.Println(err)
			utils.ErrorResponse(w, 400)
			return
		}
		doc.Grant = strings.Split(grant, ",")

		w.Header().Set("Content-Type", doc.Mime)
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		if doc.File {
			w.Write(file)
			_, err = w.Write(file)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			utils.DataResponse(w, []models.Document{doc})
		}

	}
}

// DeleteDocHandler обрабатывает удаление документа
func DeleteDocHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodDelete {
			utils.ErrorResponse(w, 405)
			return
		}

		// Получаем ID документа из параметров
		id := chi.URLParam(r, "id")

		// Удаляем документ из базы
		query := `DELETE FROM documents WHERE id = $1`
		_, err := db.Exec(query, id)
		if err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		utils.ActResponse(w, id, true)
	}
}

// LogoutHandler завершает сессию пользователя
func LogoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodDelete {
			utils.ErrorResponse(w, 405)
			return
		}

		// Получаем токен из параметров
		token := chi.URLParam(r, "token")

		// Очищаем токен в базе
		query := `UPDATE users SET token = NULL WHERE token = $1`
		if _, err := db.Exec(query, token); err != nil {
			utils.ErrorResponse(w, 500)
			return
		}

		utils.ActResponse(w, token, true)
	}
}
