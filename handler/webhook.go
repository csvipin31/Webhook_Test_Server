package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

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
	log.Println("ReadyHandler:")
     if r.Method != http.MethodGet {
        sendAPIError(w, NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE"))
        return fmt.Errorf("method not allowed")
    }
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "Ready")
	if err != nil {
        sendAPIError(w, NewAPIError( http.StatusInternalServerError, err, "Internal Server Error"))
        return err
	}
    return nil
}

func LiveHandler(w http.ResponseWriter, r *http.Request) error {
	log.Println("LiveHandler:")
     if r.Method != http.MethodGet {
        sendAPIError(w, NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE"))
        return fmt.Errorf("method not allowed")
    }

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "Live")
	if err != nil {
		sendAPIError(w, NewAPIError( http.StatusInternalServerError, err, "Internal Server Error"))
        return err
	}
    return nil
}

func HealthHandler(w http.ResponseWriter, r *http.Request) error {
	log.Println("HealthHandler:")
     if r.Method != http.MethodGet {
        sendAPIError(w, NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE"))
        return fmt.Errorf("method not allowed")
    }

	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintf(w, "OK")
	if err != nil {
		sendAPIError(w, NewAPIError( http.StatusInternalServerError, err, "Internal Server Error"))
        return err
	}
    return nil
}

func (h *WebhookHandler) DBHealthHandler(w http.ResponseWriter, r *http.Request) error {
    log.Println("DBHealthHandler:")
    if r.Method != http.MethodGet {
        sendAPIError(w, NewAPIError( http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE"))
        return fmt.Errorf("method not allowed")
    }

	if err := h.db.DescribeTable(h.tableNames[0]); err != nil {
        http.Error(w, "Database is not healthy: "+err.Error(), http.StatusInternalServerError)
        sendAPIError(w, NewAPIError( http.StatusInternalServerError, err, "Database is not healthy:unable to describe table"))
        return err
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Database is healthy"))
    return nil
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) error {
    if r.Method != http.MethodPost {
        sendAPIError(w, NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only POST requests are accepted."))
        return fmt.Errorf("method not allowed")
    }

    // Validate and extract merchant ID from the URL
    id, err := extractMerchantId(r.URL.Path)
    if err != nil {
        sendAPIError(w, NewAPIError(http.StatusBadRequest, err, "Invalid merchant ID format."))
        return err
    }

    // Process JSON payload
	var data model.UserMessageData
    err = processJSON(r, &data)
    if err != nil {
        sendAPIError(w, NewAPIError(http.StatusBadRequest, err, "Failed to process JSON."))
        return err
   }

   if err := validateByType(&data); err != nil {
        // Handle validation error
        fmt.Fprintf(w, "Validation error: %v", err)
        sendAPIError(w, NewAPIError(http.StatusBadRequest, err, "Validation error :Type does not match check payload"))
        return err
    }

    // Interact with the database
    tableName :="My_Table"
    pKey := "PK#MerchantId:" + id

	// Ensure the table exists or create if it does not exist
    if err := h.db.CreateTableIfNotExists(tableName); err != nil {
        log.Printf("Error ensuring table exists: %v", err)
        sendAPIError(w, NewAPIError(http.StatusInternalServerError, err, "Database table creation failed."))
        return err
    }

	// Store the data in the database
    if err := h.db.StoreData(tableName, pKey, data); err != nil {
        sendAPIError(w, NewAPIError(http.StatusInternalServerError, err, "Failed to store data."))
        return err
    }

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success"))
    return nil
}

func (h *WebhookHandler) WebhookEvents(w http.ResponseWriter, r *http.Request) error {
	log.Printf("WebhookEvents ")
    if r.Method != http.MethodPost {
        sendAPIError(w, NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only POST requests are accepted."))
        return fmt.Errorf("method not allowed")
    }

    // Validate and extract merchant ID from the URL
    marketplace, err := extractMerchantId(r.URL.Path)
    if err != nil {
        sendAPIError(w, NewAPIError(http.StatusBadRequest, err, "Invalid merchant ID format."))
        return err
    }

	//Read the request body into a byte slice
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
        sendAPIError(w, NewAPIError(http.StatusInternalServerError, err, "Error reading request body"))
        return err
    }
    
	defer r.Body.Close() // Close the body after reading
    // Log the raw JSON body
    log.Printf("Received body: %s", body)

	//Decode the JSON into a generic map to identify the event type
    var event model.EventTypeHolder
    if err := json.Unmarshal(body, &event); err != nil {
        sendAPIError(w, NewAPIError(http.StatusBadRequest, err, "Failed to decode JSON:"))
        return err
    }

    log.Printf("Received event type: %s", event.Type)

    switch event.Type {
    case "order/created":
		var orderCreatedEvent model.OrderCreated
        if err := unMarshallJSON(body, &orderCreatedEvent); err != nil {
            http.Error(w, "Failed to decode order created event: "+err.Error(), http.StatusBadRequest)
            return err
        }
        log.Printf("Received JSON order created: %+v", &orderCreatedEvent)
        h.handleOrderEventCreated(h.tableNames[0],marketplace, orderCreatedEvent)
	case "variant/stock-updated":
		var variantStockUpdatedEvent model.VariantStockUpdated
        if err := json.Unmarshal(body, &variantStockUpdatedEvent); err != nil {
            http.Error(w, "Failed to decode order created event: "+err.Error(), http.StatusBadRequest)
            return err
        }
        log.Printf("Received JSON order created: %+v", &variantStockUpdatedEvent)
        h.handleVariantStockUpdated(h.tableNames[1],marketplace, variantStockUpdatedEvent)
    case "product/subscribed":
		var ProductSubscribedEvent model.ProductSubscribed
        if err := json.NewDecoder(r.Body).Decode(&ProductSubscribedEvent); err != nil {
            http.Error(w, "Failed to decode Product SubscribedEvent event: "+err.Error(), http.StatusBadRequest)
            return err
        }
        h.handleProductSubscribed(marketplace,ProductSubscribedEvent)
    default:
        http.Error(w, "Unhandled event type", http.StatusBadRequest)
    }

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success"))
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
