package models

// User модель пользователя
type User struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
	Token string `json:"token"`
}

// Meta модель метаданных документа
type Meta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

// Document модель для представления документа
type Document struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Mime    string   `json:"mime"`
	File    bool     `json:"file"`
	Public  bool     `json:"public"`
	Created string   `json:"created"`
	Grant   []string `json:"grant"`
}

// APIResponse общая модель для всех методов
type APIResponse struct {
	Error    *Error                 `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
	Data     *Data                  `json:"data,omitempty"`
}

// Error модель ответа с ошибкой
type Error struct {
	Code int    `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

// Data модель ответа с данными
type Data struct {
	Docs []Document  `json:"doc,omitempty"`
	JSON interface{} `json:"json,omitempty"`
	File string      `json:"file,omitempty"`
}
