package main

import (
	"errors"
	"log"
	"net/http"

	"webhook_test_server/persistent"
	"webhook_test_server/handler"
)

func main() {
	var err error
	port := "8080"
	log.Println("server listening on port: ", port)

	// Initialize the database
	db, err := persistent.NewDatabase()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create the webhook handler with the database dependency
	// Create the webhook handler with the database dependency
    webhookHandler := handler.NewWebhookHandler(db)

	// Add health check endpoints
	http.HandleFunc("/ready", handler.ReadyHandler)
	http.HandleFunc("/live", handler.LiveHandler)
	http.HandleFunc("/health", handler.HealthHandler)
	http.HandleFunc("/dbhealth", webhookHandler.DBHealthHandler)
	// Webhook Handler handles post request
	http.HandleFunc("/", webhookHandler.HandleWebhook)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("failed to start HTTP server %v", err)
	}

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	}
}

// func (h *WebhookHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
// 	//commenting out the DB part in the func
// 	db, err := h.db.ConnectToDatabase()
// 	log.Println("POST Request URL : ", r.URL.Path)
// 	// Handle POST request for any valid path
// 	re := regexp.MustCompile(`^/(\d+)$`)
// 	matches := re.FindStringSubmatch(r.URL.Path)
// 	if len(matches) != 2 {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	// Log Request
// 	log.Println("Incoming Request URL : ", r.URL.Path)
// 	// Check if its PostMethod before processing the request
// 	if r.Method != http.MethodPost {
// 		log.Println("Method not allowed:", r.Method)
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Extract the key from the URL path using regular expressions
// 	id, err := extractMerchantId(r.URL.Path)
// 	log.Println("Merchant id: ", id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		log.Println(err)
// 		return
// 	}

// 	// Try and Decode JSON From Webhook
// 	data, err := processJSON(r)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		log.Println("failure to process json error ", err)
// 		return
// 	}

// 	//TODO :DB connection
// 	if db == nil {
// 		http.Error(w, "Database connection not available", http.StatusInternalServerError)
// 		log.Println("Database connection not available")
// 		return
// 	}
// 	//Create the table if it doesn't exist
// 	tableName := "billing-agreement-lifecycle-service-table-webhook-notification-us-qa" //os.Getenv("TABLE_NAME")
// 	log.Println("table name", tableName)
// 	err = db.CreateTableIfNotExists(tableName)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		log.Println("table error", err.Error())
// 		log.Fatal(err)
// 		return
// 	}
// 	//Store the data
// 	pKey := "PK#MerchantId:" + id
// 	err = db.StoreData(tableName, pKey, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		log.Println("error ", err)
// 		return
// 	}
// 	//db.Close()

// 	log.Println("Processed json :", data)

// 	// Send a response
// 	response := "Success"

// 	// Set the response headers
// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(http.StatusOK)
// 	// Write the response body
// 	_, err = w.Write([]byte(response))
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

// func (h *WebhookHandler) DBHealthHandler(w http.ResponseWriter, r *http.Request) {
// 	db, err := h.db.ConnectToDatabase()
// 	if err != nil {
// 		log.Fatalf("failed to initialize database: %v", err)
// 	}
// 	log.Println("ready table name from env var", os.Getenv("TABLE_NAME"))

// 	// Call DescribeTable to get the details of the table
// 	tableName := "billing-agreement-lifecycle-service-table-webhook-notification-us-qa" //os.Getenv("TABLE_NAME")
// 	err = db.DescribeTable(tableName)
// 	if err != nil {
// 		log.Fatalf("failed to describe table: %v", err)
// 		http.Error(w, "Database is not healthy", http.StatusInternalServerError)
// 		return
// 	}
// 	defer db.Close()

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Database is healthy"))
// }

// ReadyHandler, LiveHandler, HealthHandler
// func ReadyHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Println("ReadyHandler:")
// 	w.WriteHeader(http.StatusOK)
// 	_, err := fmt.Fprintf(w, "Ready")
// 	if err != nil {
// 		log.Println("Failed to write response:", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }

// func LiveHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Println("LiveHandler:")
// 	w.WriteHeader(http.StatusOK)

// 	_, err := fmt.Fprintf(w, "Live")
// 	if err != nil {
// 		log.Println("Failed to write response:", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }

// func HealthHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Println("HealthHandler:")
// 	w.WriteHeader(http.StatusOK)

// 	_, err := fmt.Fprintf(w, "OK")
// 	if err != nil {
// 		log.Println("Failed to write response:", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }

// // Database methods remain unchanged
// // ...

// // Helper functions remain unchanged
// // -- Extract the merchant Id, use this it in Dynamodb
// func extractMerchantId(path string) (string, error) {
// 	log.Println("Extract merchant Id from the Request URL :", path)
// 	re := regexp.MustCompile(`^/(\d+)$`)
// 	matches := re.FindStringSubmatch(path)
// 	if len(matches) != 2 {
// 		log.Println("issue with match")
// 		return "", fmt.Errorf("Unable to extract merchant id : Invalid URL path: %s", path)
// 	}
// 	return matches[1], nil
// }

// // -- Decode JSON Response from Webhook
// func processJSON(r *http.Request) (*model.AgreementData, error) {
// 	//-- Decode JSON from Request Body
// 	//-- This only works if the webhook is set to JSON format and not XML format
// 	log.Println("Processing JSON....")
// 	decoder := json.NewDecoder(r.Body)
// 	//-- payload now becomes a structure based on the agreement struct
// 	agreementData := model.AgreementData{}
// 	err := decoder.Decode(&agreementData)
// 	if err != nil {
// 		log.Println("Error: ", err)
// 		return nil, err
// 	}
// 	log.Println("Processed JSON....", &agreementData)
// 	return &agreementData, nil
// }