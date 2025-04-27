package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/opi-es/web-todo/pkg/db"
)

// Общие вспомогательные функции
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

// Структуры ответов
type TasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

type taskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type taskResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// Обработчики маршрутов

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeJSONError(w, "Failed to fetch tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, TasksResponse{Tasks: tasks})
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			writeJSONError(w, "Invalid now date format", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Валидация ID
	if task.ID == "" {
		writeJSONError(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Валидация заголовка
	if task.Title == "" {
		writeJSONError(w, "Task title is required", http.StatusBadRequest)
		return
	}

	// Валидация даты
	if _, err := time.Parse("20060102", task.Date); err != nil {
		writeJSONError(w, "Invalid date format, expected YYYYMMDD", http.StatusBadRequest)
		return
	}

	// Валидация правила повторения (если указано)
	if task.Repeat != "" {
		now := time.Now()
		if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
			writeJSONError(w, "Invalid repeat format: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		writeJSONError(w, "Task title is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	currentDate := now.Format("20060102")

	if req.Date == "" {
		req.Date = currentDate
	} else if _, err := time.Parse("20060102", req.Date); err != nil {
		writeJSONError(w, "Invalid date format, expected YYYYMMDD", http.StatusBadRequest)
		return
	}

	if req.Date < currentDate && req.Repeat != "" {
		nextDate, err := NextDate(now, req.Date, req.Repeat)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Date = nextDate
	} else if req.Date < currentDate && req.Repeat == "" {
		req.Date = currentDate
	}

	task := &db.Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	id, err := db.AddTask(task)
	if err != nil {
		writeJSONError(w, "Failed to add task to database", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, taskResponse{ID: id})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := db.UpdateDate(id, nextDate); err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
