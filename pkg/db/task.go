package db

import (
	"fmt"
	"strconv"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в базу данных (существующий код без изменений)
func AddTask(task *Task) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		return "", fmt.Errorf("insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("get last insert ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return strconv.FormatInt(id, 10), nil
}

// Tasks возвращает список задач (НОВАЯ функция)
func Tasks(limit int, search string) ([]*Task, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	var query string
	var args []interface{}

	// Проверяем, является ли search датой в формате 02.01.2006
	if t, err := time.Parse("02.01.2006", search); err == nil {
		query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`
		args = []interface{}{t.Format("20060102"), limit}
	} else if search != "" {
		// Поиск по подстроке
		query = `SELECT id, date, title, comment, repeat FROM scheduler 
                WHERE title LIKE ? OR comment LIKE ? 
                ORDER BY date LIMIT ?`
		searchTerm := "%" + search + "%"
		args = []interface{}{searchTerm, searchTerm, limit}
	} else {
		// Базовый запрос
		query = `SELECT id, date, title, comment, repeat FROM scheduler 
                ORDER BY date LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		tasks = append(tasks, &t)
	}

	if tasks == nil {
		tasks = []*Task{} // Для {"tasks": []} вместо null
	}

	return tasks, nil
}
