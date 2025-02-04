package config

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DbPool *pgxpool.Pool

func ConnectDatabase() {
	dbURL := GetEnv("DATABASE_URL", "postgres://tsdbadmin:gzja911fi7kzuzgg@abl2a0r9wr.ahdfzt5ekj.tsdb.cloud.timescale.com:37861/tsdb?sslmode=require")

	pool, err := pgxpool.New(context.Background(), dbURL)

	if err != nil {
		fmt.Printf("❌ Failed to connect to the database: %s\n", err)
	} else {
		log.Println("✅ Connected to TimescaleDB")
	}

	DbPool = pool
}
