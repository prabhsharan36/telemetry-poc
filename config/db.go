package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DbPool *pgxpool.Pool

func LoadEnv() {
	err := godotenv.Load()
	fmt.Print(err)
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func ConnectDatabase() {
	// dbURL := GetEnv("DATABASE_URL", "")

	pool, err := pgxpool.New(context.Background(), "postgres://tsdbadmin:gzja911fi7kzuzgg@abl2a0r9wr.ahdfzt5ekj.tsdb.cloud.timescale.com:37861/tsdb?sslmode=require")

	if err != nil {
		fmt.Printf("❌ Failed to connect to the database: %s\n", err)
	}

	DbPool = pool
	log.Println("✅ Connected to TimescaleDB")
}
