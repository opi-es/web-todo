package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/opi-es/web-todo/pkg/db"
)

type taskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type taskResponse struct {
	ID    string `json:"id,omitempty"` // Меняем int64 на string
	Error string `json:"error,omitempty"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "Invalid JSON format"})
		return
	}

	// Обработка даты
	now := time.Now()
	currentDate := now.Format(dateLayout)

	if req.Date == "" {
		req.Date = currentDate
	}

	// Проверка формата даты
	_, err = time.Parse(dateLayout, req.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "Invalid date format, expected YYYYMMDD"})
		return
	}

	// Если дата в прошлом и есть правило повторения, вычисляем следующую дату
	if req.Date < currentDate && req.Repeat != "" {
		nextDate, err := NextDate(now, req.Date, req.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, taskResponse{Error: err.Error()})
			return
		}
		req.Date = nextDate
	} else if req.Date < currentDate && req.Repeat == "" {
		req.Date = currentDate
	}

	// Проверка правила повторения (если указано)
	if req.Repeat != "" {
		_, err := NextDate(now, req.Date, req.Repeat)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, taskResponse{Error: err.Error()})
			return
		}
	}

	// Создаем задачу в базе данных
	task := &db.Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	id, err := db.AddTask(task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, taskResponse{Error: "Failed to add task to database"})
		return
	}

	writeJSON(w, http.StatusCreated, taskResponse{ID: id})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
