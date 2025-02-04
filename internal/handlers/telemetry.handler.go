package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"telemetry/internal/config"

	"github.com/go-playground/validator/v10"
)

type TelemetryData struct {
	EventType string                 `json:"eventType" validate:"required,oneof=OwnerOutreach OwnerProfileReviewed OwnerListingApproval OwnerPropertyAvailabilityCheck OwnerVisitConfirmation TenantVisitConfirmation VisitNPS TransactionNPS"`
	UUID      string                 `json:"uuid" validate:"required,uuid"`
	CreatedAt string                 `json:"createdAt" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Data      map[string]interface{} `json:"data" validate:"required"`
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
	var telemetryData TelemetryData

	if err := json.Unmarshal(body, &telemetryData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)

		return
	}

	fmt.Printf("Received telemetry event: %+v\n", telemetryData)

	validate = validator.New()
	validate.RegisterStructValidation(validateEventData, telemetryData)
	err = validate.Struct(telemetryData)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("Field '%s' validation failed: %s\n", err.Field(), err.Tag())
		}

		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("✅ Telemetry data is valid!")
	}

	if err := SaveTelemetryToDB(telemetryData); err != nil {
		http.Error(w, "Failed to save telemetry data", http.StatusInternalServerError)
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("Saved telemetry event: %+v\n", err)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Telemetry data received and saved successfully"))
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

	_, err = config.DbPool.Exec(context.Background(), query, data.EventType, data.CreatedAt, data.UUID, dataJSON)

	if err != nil {
		return fmt.Errorf("failed to insert telemetry data: %w", err)
	}

	return nil
}

var validate *validator.Validate

func validateEventData(sl validator.StructLevel) {
	td := sl.Current().Interface().(TelemetryData)

	requiredFields := map[string][]string{
		"OwnerOutreach":                  {"leadId", "responseText"},
		"OwnerProfileReviewed":           {"proposalId", "reviewResponse"},
		"OwnerListingApproval":           {"propertyId", "approvalResponse"},
		"OwnerPropertyAvailabilityCheck": {"propertyId", "isAvailable"},
		"OwnerVisitConfirmation":         {"visitId", "confirmationResponse"},
		"TenantVisitConfirmation":        {"visitId", "confirmationResponse"},
		"VisitNPS":                       {"visitId", "score"},
		"TransactionNPS":                 {"dealId", "score"},
	}

	fields, exists := requiredFields[td.EventType]
	if !exists {
		sl.ReportError(td.EventType, "EventType", "eventType", "invalidEventType", "")
		return
	}

	for _, field := range fields {
		if _, ok := td.Data[field]; !ok {
			sl.ReportError(td.Data, field, field, "requiredField", field)
		}
	}

	if td.EventType == "OwnerVisitConfirmation" || td.EventType == "TenantVisitConfirmation" {
		validResponses := []string{"confirm", "not available for rent", "reschedule", "cancel visit"}
		if val, ok := td.Data["confirmationResponse"]; ok {
			response, _ := val.(string)
			if !contains(validResponses, response) {
				sl.ReportError(td.Data, "confirmationResponse", "confirmationResponse", "invalidConfirmationResponse", "")
			}
		}
	}

	if td.EventType == "VisitNPS" || td.EventType == "TransactionNPS" {
		validScores := []string{"average", "excellent", "poor"}
		if val, ok := td.Data["score"]; ok {
			score, _ := val.(string)
			if !contains(validScores, score) {
				sl.ReportError(td.Data, "score", "score", "invalidScore", "")
			}
		}
	}
}

func contains(list []string, str string) bool {
	for _, v := range list {
		if strings.EqualFold(v, str) {
			return true
		}
	}
	return false
}
