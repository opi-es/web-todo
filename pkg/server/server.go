package server

import (
	"log"
	"net/http"

	"github.com/opi-es/web-todo/pkg/api"
)

func Start(port string, webDir string) {
	// Инициализируем API обработчики
	api.Init()

	// Обработчик статических файлов
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	log.Printf("Starting server on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
