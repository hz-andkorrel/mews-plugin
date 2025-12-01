package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

// WebhookPayload is the incoming webhook from Mews
type WebhookPayload struct {
	EnterpriseID  string  `json:"EnterpriseId"`
	IntegrationID string  `json:"IntegrationId"`
	Events        []Event `json:"Events"`
}

// Event represents a single event in the webhook
type Event struct {
	Discriminator string          `json:"Discriminator"`
	Value         json.RawMessage `json:"Value"`
}

// EntityUpdated contains the entity ID
type EntityUpdated struct {
	ID string `json:"Id"`
}

// Global Redis client
var redisClient *redis.Client
var redisChannel string
var ctx = context.Background()

func main() {
	log.Println("Mews Webhook Integration Plugin Started")

	// Initialize Redis connection
	initRedis()

	// Setup HTTP server for webhooks
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/render", handleRender)

	port := getEnvOrDefault("WEBHOOK_PORT", "8080")
	log.Printf("Webhook server listening on port %s", port)
	log.Println("Waiting for Mews webhooks...")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start webhook server: %v", err)
	}
}

func initRedis() {
	redisAddr := getEnvOrDefault("REDIS_ADDR", "host.docker.internal:6379")
	redisChannel = getEnvOrDefault("REDIS_CHANNEL", "hotel.events")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	// Test connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis at %s: %v", redisAddr, err)
		log.Println("Webhook server will continue without event publishing")
	} else {
		log.Printf("✓ Connected to Redis event bus at %s", redisAddr)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for Angular frontend
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and serve the HTML file from the frontend folder
	html, err := os.ReadFile("../frontend/index.html")
	if err != nil {
		log.Printf("Error reading ../frontend/index.html: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(html)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the webhook payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process events
	processEvents(payload.Events)

	// Respond to Mews
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func processEvents(events []Event) {
	for _, event := range events {
		if event.Discriminator == "ServiceOrderUpdated" {
			var entity EntityUpdated
			if err := json.Unmarshal(event.Value, &entity); err != nil {
				log.Printf("Error parsing event: %v", err)
				continue
			}

			// Log the check-in
			log.Printf("Guest checked in - Reservation ID: %s", entity.ID)

			// Publish event to Redis event bus
			publishToEventBus("guest.checked_in", entity.ID)
		}
	}
}

func publishToEventBus(eventType string, reservationID string) {
	if redisClient == nil {
		log.Println("Redis client not initialized, skipping event publish")
		return
	}

	// Create event payload
	eventPayload := map[string]string{
		"type":           eventType,
		"reservation_id": reservationID,
	}

	eventJSON, err := json.Marshal(eventPayload)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	// Publish to Redis
	err = redisClient.Publish(ctx, redisChannel, eventJSON).Err()
	if err != nil {
		log.Printf("Failed to publish event to Redis: %v", err)
	} else {
		log.Printf("✓ Published to event bus: %s (ID: %s)", eventType, reservationID)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
