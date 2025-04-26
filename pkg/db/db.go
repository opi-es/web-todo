package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var (
	db *sql.DB
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);

CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
`

func Init(dbFile string) error {
	// Проверяем существование файла
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	// Открываем базу данных
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	// Проверяем соединение
	if err = db.Ping(); err != nil {
		return err
	}

	// Если файла не было, создаем схему
	if install {
		if _, err = db.Exec(schema); err != nil {
			return err
		}
		log.Printf("Created new DB file %s with schema", dbFile)
	}

	return nil
}

// GetDB возвращает соединение с базой данных
func GetDB() *sql.DB {
	return db
}
