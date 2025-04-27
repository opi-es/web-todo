package db

import (
	"database/sql"
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

// AddTask добавляет задачу в базу данных
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

// Tasks возвращает список задач
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

func GetTask(id string) (*Task, error) {
	var task Task
	err := db.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?",
		id,
	).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	return &task, nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`UPDATE scheduler 
		SET date = ?, title = ?, comment = ?, repeat = ? 
		WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID,
	)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return tx.Commit()
}

// DeleteTask удаляет задачу по ID
func DeleteTask(id string) error {
	res, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// UpdateDate обновляет только дату задачи
func UpdateDate(id string, newDate string) error {
	res, err := db.Exec(
		"UPDATE scheduler SET date = ? WHERE id = ?",
		newDate, id,
	)
	if err != nil {
		return fmt.Errorf("update date failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
