package handler

import (
    "fmt"
    "log"
    "net/http"
    "regexp"
	"encoding/json"

	"webhook_test_server/persistent"
	"webhook_test_server/model"

)

type WebhookHandler struct {
    db persistent.DatabaseInterface
}

func NewWebhookHandler(db persistent.DatabaseInterface) *WebhookHandler {
    return &WebhookHandler{db: db}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Validate and extract merchant ID from the URL
    id, err := extractMerchantId(r.URL.Path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process JSON payload
    data, err := processJSON(r)
    if err != nil {
        http.Error(w, "Failed to process JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Interact with the database
    tableName := "My_Table"
    pKey := "PK#MerchantId:" + id

	// Ensure the table exists or create if it does not exist
    if err := h.db.CreateTableIfNotExists(tableName); err != nil {
        log.Printf("Error ensuring table exists: %v", err)
        http.Error(w, "Failed to ensure table exists: "+err.Error(), http.StatusInternalServerError)
        return
    }

	// Store the data in the database
    if err := h.db.StoreData(tableName, pKey, data); err != nil {
        http.Error(w, "Failed to store data: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success"))
}

func (h *WebhookHandler) DBHealthHandler(w http.ResponseWriter, r *http.Request) {

    tableName := "My_Table"
	if err := h.db.DescribeTable(tableName); err != nil {
        http.Error(w, "Database is not healthy: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Database is healthy"))
}

// ReadyHandler, LiveHandler, HealthHandler
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ReadyHandler:")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "Ready")
	if err != nil {
		log.Println("Failed to write response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LiveHandler:")
	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintf(w, "Live")
	if err != nil {
		log.Println("Failed to write response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HealthHandler:")
	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintf(w, "OK")
	if err != nil {
		log.Println("Failed to write response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Utility functions (extractMerchantId, processJSON) remain unchanged

// Helper functions remain unchanged
// -- Extract the merchant Id, use this it in Dynamodb
func extractMerchantId(path string) (string, error) {
	log.Println("Extract merchant Id from the Request URL :", path)
	re := regexp.MustCompile(`^/(\d+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) != 2 {
		log.Println("issue with match")
		return "", fmt.Errorf("unable to extract merchant id : Invalid URL path: %s", path)
	}
	return matches[1], nil
}

// -- Decode JSON Response from Webhook
func processJSON(r *http.Request) (*model.UserMessageData, error) {
	//-- Decode JSON from Request Body
	//-- This only works if the webhook is set to JSON format and not XML format
	log.Println("Processing JSON....")
	decoder := json.NewDecoder(r.Body)
	//-- payload now becomes a structure based on the agreement struct
	userMessageData := model.UserMessageData{}
	err := decoder.Decode(&userMessageData)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	log.Println("Processed JSON....", &userMessageData)
	return &userMessageData, nil
}