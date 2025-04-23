package server

import (
	"log"
	"net/http"
)

func Start(port string, webDir string) {
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	log.Printf("Starting server on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
