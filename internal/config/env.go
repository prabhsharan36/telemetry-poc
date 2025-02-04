package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func LoadEnv() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	environmentPath := filepath.Join(dir, ".env")
	err = godotenv.Load(environmentPath)
	if err != nil {
		color.Red("❌ Error loading .env file")
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
