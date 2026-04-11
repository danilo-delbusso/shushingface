package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GroqAPIKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, os.ErrNotExist
	}
	return &Config{
		GroqAPIKey: apiKey,
	}, nil
}
