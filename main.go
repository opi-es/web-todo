package main

import (
	"os"

	"github.com/opi-es/web-todo/pkg/server"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	server.Start(port, webDir)
}
