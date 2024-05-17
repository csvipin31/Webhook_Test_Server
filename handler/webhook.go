package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"webhook_test_server/model"
	"webhook_test_server/persistent"
)

func (h *WebhookHandler) WebhookEvents(w http.ResponseWriter, r *http.Request) error {
	handlerName := "WebhookEvents"
	startTime, method, url := logRequestStart(r, handlerName)
	if r.Method != http.MethodPost {
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only POST requests are accepted.")
	}

	// Validate and extract merchant ID from the URL
	marketplace, err := extractMerchantId(r.URL.Path)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Invalid merchant ID format.")
	}

	//Read the request body into a byte slice
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, err, "Error reading request body")
	}

	defer r.Body.Close() // Close the body after reading
	// Log the raw JSON body
	log.Printf("Received body: %s", body)

	//Decode the JSON into a generic map to identify the event type
	var event model.EventTypeHolder
	if err := json.Unmarshal(body, &event); err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Failed to decode JSON:")
	}

	log.Printf("Received event type: %s", event.Type)

	handler, found := h.eventHandlers[event.Type]

	if !found {
		log.Printf("No handler found for event type: %s", event.Type)
		return NewAPIError(http.StatusBadRequest, err, fmt.Sprintf("Unhandled event type: %s", event.Type))
	}

	// Handle the event
	if err := handler(marketplace, body); err != nil {
		logRequestEnd(startTime, method, url, handlerName, http.StatusInternalServerError)
		return NewAPIError(http.StatusBadRequest, err, "Failed to handle event")
	}

	// Log success and write the response
	logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Success"})
	return nil
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) error {
	handlerName := "HandleWebhook"
	startTime, method, url := logRequestStart(r, handlerName)
	if r.Method != http.MethodPost {
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only POST requests are accepted.")
	}

	// Validate and extract merchant ID from the URL
	id, err := extractMerchantId(r.URL.Path)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Invalid merchant ID format.")
	}

	// Process JSON payload
	var data model.UserMessageData
	err = processJSON(r, &data)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Failed to process JSON.")
	}

	if err := validateByType(&data); err != nil {
		// Handle validation error
		fmt.Fprintf(w, "Validation error: %v", err)
		return NewAPIError(http.StatusBadRequest, err, "Validation error :Type does not match check payload")
	}

	// Interact with the database
	tableName := "My_Table"
	pKey := "PK#MerchantId:" + id

	// Ensure the table exists or create if it does not exist
	if err := h.db.CreateTableIfNotExists(tableName); err != nil {
		log.Printf("Error ensuring table exists: %v", err)
		return NewAPIError(http.StatusInternalServerError, err, "Database table creation failed.")
	}

	// Store the data in the database
	if err := h.db.StoreData(tableName, pKey, data); err != nil {
		return NewAPIError(http.StatusInternalServerError, err, "Failed to store data.")
	}

	// Log success and write the response
	logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Success"})
	return nil
}

func (h *WebhookHandler) GetOrderEventsByPK(w http.ResponseWriter, r *http.Request) error {
	handlerName := "GetOrderEventsByPK"
	startTime, method, url := logRequestStart(r, handlerName)

	if r.Method != http.MethodGet {
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
	}

	// Extract parameters from the URL or request
	merchantId := r.URL.Query().Get("merchantId")
	externalOrderId := r.URL.Query().Get("externalOrderId")
	if merchantId == "" || externalOrderId == "" {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("missing merchantId or externalOrderId parameter"), "Missing merchantId or externalOrderId parameter")
	}

	// Construct the primary key
	pk := fmt.Sprintf("#PK#%s#%s", merchantId, externalOrderId)

	// Fetch data based on primary key without requiring SK
	tableName := h.tableNames[0]
	result, err := h.db.FetchByPrimaryKey(tableName, pk)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Failed to fetch order events:")
	}

	// Check if items were found
	if len(result.Items) == 0 {
		http.Error(w, "Order events not found", http.StatusNotFound)
		return NewAPIError(http.StatusNotFound, fmt.Errorf("order event Not found for PK : %s", pk), "Order events not found")
	}

	// Convert DynamoDB items to OrderEvent structs
	var orderEvents []persistent.OrderEvent
	for _, item := range result.Items {
		event, err := persistent.ConvertDynamoItemToOrderEvent(item)
		if err != nil {
			return NewAPIError(http.StatusInternalServerError, err, "Failed to parse order event")
		}
		orderEvents = append(orderEvents, event)
	}

	// Write the result to the response
	writeJSON(w, http.StatusOK, orderEvents)
	logRequestEnd(startTime, method, url, handlerName, http.StatusOK)

	return nil
}

func (h *WebhookHandler) GetOrderByExternalID(w http.ResponseWriter, r *http.Request) error {
	handlerName := "GetOrderByExternalID"
	startTime, method, url := logRequestStart(r, handlerName)

	if r.Method != http.MethodGet {
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
	}

	// Extract parameters from the URL or request
	externalOrderId := r.URL.Query().Get("externalOrderId")
	if externalOrderId == "" {
		return NewAPIError(http.StatusBadRequest, fmt.Errorf("missing externalOrderId parameter"), "Missing externalOrderId parameter")
	}

	// Fetch data based on primary key without requiring SK
	tableName := h.tableNames[0]
	result, err := h.db.QueryOrderEventsByExternalOrderId(tableName, externalOrderId)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, err, "Failed to fetch order events: by external order Id")
	}

	// Check if items were found
	if len(result.Items) == 0 {
		return NewAPIError(http.StatusNotFound, fmt.Errorf("order event Not found for PK : %s", externalOrderId), "Order events not found")
	}

	// Convert DynamoDB items to OrderEvent structs
	var orderEvents []persistent.OrderEvent
	for _, item := range result.Items {
		event, err := persistent.ConvertDynamoItemToOrderEvent(item)
		if err != nil {
			return NewAPIError(http.StatusInternalServerError, err, "Failed to parse order event")
		}
		orderEvents = append(orderEvents, event)
	}

	// Write the result to the response
	writeJSON(w, http.StatusOK, orderEvents)
	logRequestEnd(startTime, method, url, handlerName, http.StatusOK)

	return nil
}
