package main

import (
	"webserver/config"
	"webserver/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.New()

	if err != nil {
		log.Fatalln("Failed to parse config", err)
	}

	srv, err := server.New(cfg)

	if err != nil {
		log.Fatalln("Failed to create server", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalln("Failed to shutdown gracefully!", err)
	}

	log.Println("Server shutting down...")
}
