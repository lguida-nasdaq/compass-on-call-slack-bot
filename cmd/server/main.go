package main

import (
	"log"

	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/config"
	"github.com/metriodev/pompiers/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	compassClient := api.NewCompassClient(cfg)
	jiraClient := api.NewJiraClient(cfg)

	srv := server.NewServer(cfg, compassClient, jiraClient)
	if err := srv.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
