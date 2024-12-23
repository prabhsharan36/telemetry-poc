package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryData struct {
	EventType string                 `json:"eventType"`
	UUID      string                 `json:"uuid"`
	CreatedAt string                 `json:"createdAt"`
	Data      map[string]interface{} `json:"data"`
}

var dbPool *pgxpool.Pool

func SaveTelemetryToDB(data TelemetryData) error {
	dataJSON, err := json.Marshal(data.Data)

	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	query := `
		INSERT INTO telemetry (event_type, created_at, uuid, data)
		VALUES ($1, $2, $3, $4)
	`
	_, err = dbPool.Exec(context.Background(), query, data.EventType, data.CreatedAt, data.UUID, dataJSON)

	if err != nil {
		return fmt.Errorf("failed to insert telemetry data: %w", err)
	}

	return nil
}

func TelemetryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)

		return
	}

	defer r.Body.Close()

	// Parse the JSON payload
	var telemetry TelemetryData

	if err := json.Unmarshal(body, &telemetry); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)

		return
	}

	fmt.Printf("Received telemetry event: %+v\n", telemetry)

	if err := SaveTelemetryToDB(telemetry); err != nil {
		http.Error(w, "Failed to save telemetry data", http.StatusInternalServerError)
		fmt.Printf("Error: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Telemetry data received and saved successfully"))
}

func main() {

	var err error
	dbURL := "postgres://tsdbadmin:ap2xxay85nzspjt6@n0p9m7tpqr.cfgn84scyg.tsdb.cloud.timescale.com:39385/tsdb?sslmode=require"
	dbPool, err = pgxpool.New(context.Background(), dbURL)

	if err != nil {
		fmt.Printf("Failed to connect to the database: %s\n", err)
		return
	}

	defer dbPool.Close()

	http.HandleFunc("/capture-event", TelemetryHandler)

	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
