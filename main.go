package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"telemetry/config"

	"github.com/gin-gonic/gin"
)

type TelemetryData struct {
	EventType string                 `json:"eventType"`
	UUID      string                 `json:"uuid"`
	CreatedAt string                 `json:"createdAt"`
	Data      map[string]interface{} `json:"data"`
}

func SaveTelemetryToDB(data TelemetryData) error {
	dataJSON, err := json.Marshal(data.Data)

	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	query := `
		INSERT INTO telemetry_data (event_type, created_at, uuid, data)
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
	config.LoadEnv()
	port := config.GetEnv("PORT", "8080")
	r := gin.Default()
	r.GET("/", handlers.HomeHandler)

	http.HandleFunc("/capture-event", TelemetryHandler)

	log.Printf("Server running on port %s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}

	defer dbPool.Close()
}
