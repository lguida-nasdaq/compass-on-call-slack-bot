package main

import (
	"log"
	"os"
	"strconv"

	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/app"
	"github.com/metriodev/pompiers/internal/server"
)

func main() {
	apiUser := os.Getenv("USER_EMAIL")
	apiKey := os.Getenv("API_KEY")
	cloudId := os.Getenv("ATLASSIAN_CLOUD_ID")

	compassClient := api.NewCompassClient(apiUser, apiKey, cloudId)
	jiraClient := api.NewJiraClient(apiUser, apiKey)

	app := app.NewApp(compassClient, jiraClient)
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Failed to get available port: %v", err)
	}
	host := os.Getenv("HOST")

	srv := server.NewServer(app, host, port, os.Getenv("SLACK_SIGNING_SECRET"))
	if err := srv.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
