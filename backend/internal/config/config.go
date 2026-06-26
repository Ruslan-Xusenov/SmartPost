package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	BotToken    string
	DatabaseURL string
	RedisURL    string
	WebhookURL  string
	APIPort     int
	TWAURL      string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	// Try to load .env file from multiple possible locations
	// (supports running from backend/, backend/cmd/smartpost/, or project root)
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
		"../../../.env",
	}
	for _, p := range envPaths {
		if err := godotenv.Load(p); err == nil {
			break
		}
	}

	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		WebhookURL:  os.Getenv("WEBHOOK_URL"),
		TWAURL:      os.Getenv("TWA_URL"),
	}

	// Parse API port with default
	portStr := os.Getenv("API_PORT")
	if portStr == "" {
		cfg.APIPort = 8080
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid API_PORT: %w", err)
		}
		cfg.APIPort = port
	}

	// Validate required fields
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	return cfg, nil
}

// RedisAddr extracts host:port from the Redis URL for asynq.
// Converts "redis://host:port" → "host:port"
func (c *Config) RedisAddr() string {
	addr := c.RedisURL
	// Strip redis:// prefix
	if len(addr) > 8 && addr[:8] == "redis://" {
		addr = addr[8:]
	}
	return addr
}
