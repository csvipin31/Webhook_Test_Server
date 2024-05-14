package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"webhook_test_server/model"
	"webhook_test_server/persistent"
)

type WebhookHandler struct {
    db            persistent.DatabaseInterface
    tableNames    []string
}
   

func NewWebhookHandler(db persistent.DatabaseInterface, tableName []string) *WebhookHandler {
    return &WebhookHandler{
        db:           db,
        tableNames:   tableName,
    }
}

// Health Check : ReadyHandler, LiveHandler, HealthHandler
func ReadyHandler(w http.ResponseWriter, r *http.Request) error {
    handlerName := "ReadyHandler"
    startTime, method, url := logRequestStart(r, handlerName)
    log.Println("ReadyHandler:")

    if r.Method != http.MethodGet {
        return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method Not Allowed") , "Only GET requests are accepted.")
    }

    logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
    writeJSON(w, http.StatusOK, map[string]string{"message": "Server is Ready"})
    return nil
}


func LiveHandler(w http.ResponseWriter, r *http.Request) error {
    handlerName := "LiveHandler"
    startTime, method, url := logRequestStart(r, handlerName)

    if r.Method != http.MethodGet {
        return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method Not Allowed"), "Only GET requests are accepted.")
    }

    logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
    writeJSON(w, http.StatusOK, map[string]string{"message": "Server is Live"})
    return nil
}


func HealthHandler(w http.ResponseWriter, r *http.Request) error {
    handlerName := "HealthHandler"
    startTime, method, url := logRequestStart(r, handlerName)

    if r.Method != http.MethodGet {
        return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method Not Allowed"), "Only GET requests are accepted.")
    }

    logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
    writeJSON(w, http.StatusOK, map[string]string{"message": "Server is Healthy"})
    return nil
}


func (h *WebhookHandler) DBHealthHandler(w http.ResponseWriter, r *http.Request) error {
    handlerName := "DBHealthHandler"
	startTime, method, url := logRequestStart(r, handlerName)

    if r.Method != http.MethodGet {
        return NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
    }

	tableName := h.tableNames[0]
	err := h.db.DescribeTable(tableName)
	if err != nil {
		logRequestEnd(startTime, method, url, handlerName, http.StatusInternalServerError)
		return NewAPIError(http.StatusInternalServerError, err, "Database is unhealthy")
	}

	logRequestEnd(startTime, method, url, handlerName, http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Database is healthy"})
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
    tableName :="My_Table"
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

    switch event.Type {
    case "order/created":
		var orderCreatedEvent model.OrderCreated
        if err := unMarshallJSON(body, &orderCreatedEvent); err != nil {
            return NewAPIError(http.StatusBadRequest, err, "Failed to decode order created event")
        }
        log.Printf("Received JSON order created: %+v", &orderCreatedEvent)
        h.handleOrderEventCreated(h.tableNames[0],marketplace, orderCreatedEvent)
	case "variant/stock-updated":
		var variantStockUpdatedEvent model.VariantStockUpdated
        if err := json.Unmarshal(body, &variantStockUpdatedEvent); err != nil {
            return NewAPIError(http.StatusBadRequest, err, "Failed to decode variant stock updated event")
        }
        log.Printf("Received JSON order created: %+v", &variantStockUpdatedEvent)
        h.handleVariantStockUpdated(h.tableNames[1],marketplace, variantStockUpdatedEvent)
    case "product/subscribed":
		var ProductSubscribedEvent model.ProductSubscribed
        if err := json.NewDecoder(r.Body).Decode(&ProductSubscribedEvent); err != nil {
            return NewAPIError(http.StatusBadRequest, err, "Failed to decode product subscribed event")
        }
        h.handleProductSubscribed(marketplace,ProductSubscribedEvent)
    default:
       return NewAPIError(http.StatusBadRequest, fmt.Errorf("unhandled event type"), "Unhandled event type")
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
        return NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
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
        return NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
    }

    // Extract parameters from the URL or request
    externalOrderId := r.URL.Query().Get("externalOrderId")
    if  externalOrderId == "" {
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


/* Utility functions

// func (h *WebhookHandler) handleOrderCreated(tableName string,marketplace string, event model.OrderCreated) {
// 	log.Printf("Processing Order Created event for marketplace: %s, Event ID: %s , External Order ID: %s", marketplace, event.EventId, event.ExternalOrderID)

// 	// Create an instance of EventOptions
//     opts := model.EventOptions{}
// 	// OrderCreated has an ExternalOrderID that could be empty and not necessarily part of every event.
//     if event.ExternalOrderID != "" { // Check if ExternalOrderID is non-empty.
//         opts.ExternalOrderId = &event.ExternalOrderID // If non-empty, set it in the options.
//     }

//     err := h.db.StoreEventData(tableName, event.Type, event.EventId, event.LastUpdated, marketplace, event, opts)
//     if err != nil {
//         log.Printf("Error storing event data in handleOrderCreated: %v", err)
//         return
//     }

//     log.Println("handleOrderCreated: Successfully processed order creation")
// }
*/

func (h *WebhookHandler) handleOrderEventCreated(tableName string,marketplace string, event model.OrderCreated) {
	log.Printf("Processing Order Created event for marketplace: %s , External Order ID: %s", marketplace, event.ExternalOrderID)

    err := h.db.StoreOrderEventData(tableName, event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
    if err != nil {
        log.Printf("Error storing event data in handleOrderCreated: %v", err)
        return
    }

    log.Println("handleOrderCreated: Successfully processed order creation")
}

func (h *WebhookHandler) handleVariantStockUpdated(tableName string,marketplace string, event model.VariantStockUpdated) {
	log.Printf("Processing handle Variant Stock Updated for marketplace: %s, Event ID: %s , deal ID: %s", marketplace, event.EventId, event.DealID)

	// Create an instance of EventOptions
    opts := model.EventOptions{}
	// OrderCreated has an ExternalOrderID that could be empty and not necessarily part of every event.
    if event.DealID != "" { // Check if ExternalOrderID is non-empty.
        opts.DealId = &event.DealID // If non-empty, set it in the options.
    }

    err := h.db.StoreEventData(tableName, event.Type, event.EventId, event.LastUpdated, marketplace, event, opts)
    if err != nil {
        log.Printf("Error storing event data in handleVariantStockUpdated: %v", err)
        return
    }

    log.Println("handle Variant Stock Updated: Successfully processed order creation")
}

func (h *WebhookHandler) handleProductSubscribed(marketplace string,event model.ProductSubscribed) {
	// log.Printf("Processing Product Subscried event for marketplace: %s, Event ID: %s , Deal ID: %s", marketplace, event.EventId, event.DealID)
    // details := make(map[string]interface{})
    // if err := json.Unmarshal(data, &details); err != nil {
    //     return
    // }
    //h.db.StoreEventData("WebhookEvent",details["$type"].(string),details["eventId"].(string), details["lastUpdated"].(string), marketplace, details)
}

// -- Extract the merchant Id, use this it in Dynamodb
func extractMerchantId(path string) (string, error) {
	log.Println("Extract merchant Id from the Request URL :", path)
	re := regexp.MustCompile(`^/([A-Za-z0-9_]+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) != 2 {
		log.Println("issue with match")
		return "", fmt.Errorf("unable to extract merchant id : Invalid URL path: %s", path)
	}
	return matches[1], nil
}


func processJSON(r *http.Request, target interface{}) error {
log.Println("Processing JSON...")
    // Log the raw JSON data received by the server
    // Read the request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Println("Error reading request body:", err)
        return err
    }

    // Log the raw JSON data received by the server
    log.Printf("Received JSON: %s", string(body))

	// Reset the request body to its original state
    r.Body = io.NopCloser(bytes.NewBuffer(body))

    // Create a new decoder for the request body
    decoder := json.NewDecoder(r.Body)
    
    // Decode the JSON into the target structure using reflection
    if err := decoder.Decode(target); err != nil {
        log.Printf("Error decoding JSON: %v", err)
        return err
    }
    
    log.Printf("Processed JSON: %+v", target)
    return nil
}


func unMarshallJSON(body []byte, target interface{}) error {
    log.Println("Processing JSON...")
    // Log the raw JSON data received by the server
    log.Printf("Received JSON: %s", string(body))

    if err := json.Unmarshal(body, &target); err != nil {
        log.Printf("Failed to decode order created event: "+err.Error())
        return err
    }    
    log.Printf("Processed JSON: %+v", target)
    return nil
}

func logRequestStart(r *http.Request, handlerName string) (startTime time.Time, method, url string) {
	startTime = time.Now()
	method = r.Method
	url = r.URL.String()

	log.Printf("INFO: Received request: Method=%s, URL=%s, Handler=%s", method, url, handlerName)
	return startTime, method, url
}

func logRequestEnd(startTime time.Time, method, url, handlerName string, responseStatus int) {
	processingTime := time.Since(startTime)

	log.Printf("INFO: Request processed successfully: Method=%s, URL=%s, Handler=%s, ResponseStatus=%d, ProcessingTime=%s",
		method, url, handlerName, responseStatus, processingTime)
}