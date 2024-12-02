package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"cache-web-server/internal/models"
)

// HTTP-статусы по заданию
var httpStatus = map[int]string{
	http.StatusOK:                  "Все ок",
	http.StatusBadRequest:          "Некорректные параметры",
	http.StatusUnauthorized:        "Не авторизован",
	http.StatusForbidden:           "Нет прав доступа",
	http.StatusMethodNotAllowed:    "Неверный метод запроса",
	http.StatusInternalServerError: "Нежданчик",
	http.StatusNotImplemented:      "Метод не реализован",
}

// ErrorResponse формирует ответ с ошибкой
func ErrorResponse(w http.ResponseWriter, code int) {
	errResp := models.APIResponse{
		Error: &models.Error{
			Code: code,
			Text: httpStatus[code],
		},
	}

	WriteJSONResponse(w, code, errResp)
}

// ActResponse формирует ответ подтверждения действия
func ActResponse(w http.ResponseWriter, key string, value interface{}) {
	actResp := models.APIResponse{
		Response: map[string]interface{}{
			key: value,
		},
	}

	WriteJSONResponse(w, 200, actResp)
}

// UploadResponse формирует ответ с данными загруженного файла
func UploadResponse(w http.ResponseWriter, jsonParsed map[string]interface{}, name string) {
	uploadResp := models.APIResponse{
		Data: &models.Data{
			JSON: jsonParsed,
			File: name,
		},
	}

	WriteJSONResponse(w, 200, uploadResp)
}

// DataResponse формирует ответ с данными
func DataResponse(w http.ResponseWriter, docs []models.Document) {
	dataResp := models.APIResponse{
		Data: &models.Data{
			Docs: docs,
		},
	}

	WriteJSONResponse(w, 200, dataResp)
}

// WriteJSONResponse отправляет JSON-ответ клиенту
func WriteJSONResponse(w http.ResponseWriter, statusCode int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Не удалось закодировать ответ: %v", err)
	}
}
