package config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Config struct to hold Mastodon and RSS feed details
type Config struct {
	LastRunTime               time.Time `json:"last_run_time"`
	BaseUrl                   string    `json:"baseUrl"`
	FeedEndpoint              string    `json:"feedEndpoint"`
	DoesMetaOgHasRelativePath bool      `json:"doesMetaOgHasRelativePath"`
	Socials                   []string  `json:"socials"`
}

func LoadConfig(filename string) Config {
	configFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}
	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Error parsing configuration JSON: %v", err)
	}
	return config
}

func SaveConfig(filename string, config Config) {
	configFile, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("Error marshalling configuration: %v", err)
	}
	if err := os.WriteFile(filename, configFile, 0644); err != nil {
		log.Fatalf("Error writing configuration file: %v", err)
	}
}
