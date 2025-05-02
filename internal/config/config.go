package config

import (
	"os"
)

type Config struct {
	User          string
	APIKey        string
	SigningSecret string
	CloudID       string
}

func LoadConfig() (Config, error) {
	return Config{
		User:          os.Getenv("JIRA_USER"),
		APIKey:        os.Getenv("JIRA_API_KEY"),
		SigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
		CloudID:       os.Getenv("JIRA_CLOUD_ID"),
	}, nil
}
