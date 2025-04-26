package main

import (
	"log"
	"os"

	"github.com/opi-es/web-todo/pkg/db"
	"github.com/opi-es/web-todo/pkg/server"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
	defaultDB   = "scheduler.db"
)

func main() {
	// Получаем настройки из переменных окружения
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = defaultDB
	}

	// Инициализируем базу данных
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Запускаем сервер
	server.Start(port, webDir)
}
