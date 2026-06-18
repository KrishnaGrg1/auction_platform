package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT              string
	AUTH_PORT         string
	AUCTION_PORT      string
	DB_URL            string
	JWT_SECRET        string
	RESEND_API_KEY    string
	FRONTEND_URL      string
	RESEND_EMAIL_FROM string
	REDIS_URL         string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using system env")
	}

	return &Config{
		AUTH_PORT:         getEnv("AUTH_PORT", "8081"),
		AUCTION_PORT:      getEnv("AUCTION_PORT", "8080"),
		DB_URL:            getEnv("GOOSE_DBSTRING", ""),
		JWT_SECRET:        getEnv("JWT_SECRET", ""),
		RESEND_API_KEY:    getEnv("RESEND_API_KEY", ""),
		FRONTEND_URL:      getEnv("FRONTEND_URL", ""),
		RESEND_EMAIL_FROM: getEnv("RESEND_EMAIL_FROM", ""),
		REDIS_URL:         getEnv("REDIS_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback

}
