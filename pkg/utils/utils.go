package utils

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	// Load env vars from .env
	return godotenv.Load()
}

func GetOptionalEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetRequiredEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("Env variable " + key + " required.")
}
