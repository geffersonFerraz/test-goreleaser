package main

import (
	"os"
	"test-goreleaser/server"
)

func main() {
	listen := os.Getenv("SERVER_IP")
	if listen == "" {
		listen = "0.0.0.0"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewHttpServer(listen, port)
	srv.Listen()
}
