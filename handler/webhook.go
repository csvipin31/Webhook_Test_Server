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
	var data model.UserMessageData
    err = processJSON(r, &data)
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

func (h *WebhookHandler) HandleEventWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleEventWebhook ")
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Validate and extract merchant ID from the URL
    marketplace, err := extractMerchantId(r.URL.Path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process JSON payload
	var data model.VariantStockUpdated
    err = processJSON(r, &data)
    if err != nil {
    http.Error(w, "Failed to process JSON: "+err.Error(), http.StatusBadRequest)
    return
   }

	log.Printf("Received JSON order created: %+v", &data)
        h.handleVariantStockUpdated(marketplace, data)

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success"))
}

func (h *WebhookHandler) WebhookEvents(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebhookEvents ")
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Validate and extract merchant ID from the URL
    marketplace, err := extractMerchantId(r.URL.Path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	//Read the request body into a byte slice
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
	defer r.Body.Close() // Close the body after reading
    // Log the raw JSON body
    log.Printf("Received body: %s", body)

	//Decode the JSON into a generic map to identify the event type
    var event model.EventTypeHolder
    if err := json.Unmarshal(body, &event); err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Received event type: %s", event.Type)


    switch event.Type {
    case "order/created":
		var orderCreatedEvent model.OrderCreated
        if err := unMarshallJSON(body, &orderCreatedEvent); err != nil {
            http.Error(w, "Failed to decode order created event: "+err.Error(), http.StatusBadRequest)
            return
        }
        log.Printf("Received JSON order created: %+v", &orderCreatedEvent)
        h.handleOrderCreated(marketplace, orderCreatedEvent)
	case "variant/stock-updated":
		var variantStockUpdatedEvent model.VariantStockUpdated
        if err := json.Unmarshal(body, &variantStockUpdatedEvent); err != nil {
            http.Error(w, "Failed to decode order created event: "+err.Error(), http.StatusBadRequest)
            return
        }
        log.Printf("Received JSON order created: %+v", &variantStockUpdatedEvent)
        h.handleVariantStockUpdated(marketplace, variantStockUpdatedEvent)
    case "product/subscribed":
		var ProductSubscribedEvent model.ProductSubscribed
        if err := json.NewDecoder(r.Body).Decode(&ProductSubscribedEvent); err != nil {
            http.Error(w, "Failed to decode Product SubscribedEvent event: "+err.Error(), http.StatusBadRequest)
            return
        }
		//log.Printf("Received JSON: %s", ProductSubscribedEvent)
        h.handleProductSubscribed(marketplace,ProductSubscribedEvent)
    default:
        http.Error(w, "Unhandled event type", http.StatusBadRequest)
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

// Utility functions

func (h *WebhookHandler) handleOrderCreated(marketplace string, event model.OrderCreated) {
	log.Printf("Processing Order Created event for marketplace: %s, Event ID: %s , External Order ID: %s", marketplace, event.EventId, event.ExternalOrderID)

	// Create an instance of EventOptions
    opts := model.EventOptions{}
	// OrderCreated has an ExternalOrderID that could be empty and not necessarily part of every event.
    if event.ExternalOrderID != "" { // Check if ExternalOrderID is non-empty.
        opts.ExternalOrderId = &event.ExternalOrderID // If non-empty, set it in the options.
    }

    err := h.db.StoreEventData("EventWebhook", event.Type, event.EventId, event.LastUpdated, marketplace, event, opts)
    if err != nil {
        log.Printf("Error storing event data in handleOrderCreated: %v", err)
        return
    }

    log.Println("handleOrderCreated: Successfully processed order creation")
}

func (h *WebhookHandler) handleVariantStockUpdated(marketplace string, event model.VariantStockUpdated) {
	log.Printf("Processing handle Variant Stock Updated for marketplace: %s, Event ID: %s , deal ID: %s", marketplace, event.EventId, event.DealID)

	// Create an instance of EventOptions
    opts := model.EventOptions{}
	// OrderCreated has an ExternalOrderID that could be empty and not necessarily part of every event.
    if event.DealID != "" { // Check if ExternalOrderID is non-empty.
        opts.DealId = &event.DealID // If non-empty, set it in the options.
    }

    err := h.db.StoreEventData("EventWebhook", event.Type, event.EventId, event.LastUpdated, marketplace, event, opts)
    if err != nil {
        log.Printf("Error storing event data in handleVariantStockUpdated: %v", err)
        return
    }

    log.Println("handle Variant Stock Updated: Successfully processed order creation")
}

func (h *WebhookHandler) handleProductSubscribed(marketplace string,event model.ProductSubscribed) {
	log.Printf("Processing Product Subscried event for marketplace: %s, Event ID: %s , Deal ID: %s", marketplace, event.EventId, event.DealID)
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

// // -- Decode JSON Response from Webhook
// func processJSON(r *http.Request) (*model.UserMessageData, error) {
// 	//-- Decode JSON from Request Body
// 	//-- This only works if the webhook is set to JSON format and not XML format
// 	log.Println("Processing JSON....")
// 	decoder := json.NewDecoder(r.Body)
// 	//-- payload now becomes a structure based on the agreement struct
// 	userMessageData := model.UserMessageData{}
// 	err := decoder.Decode(&userMessageData)
// 	if err != nil {
// 		log.Println("Error: ", err)
// 		return nil, err
// 	}
// 	log.Println("Processed JSON....", &userMessageData)
// 	return &userMessageData, nil
// }


// T must be a pointer to a struct that can be unmarshalled from JSON.
// func processJSON[T any](r *http.Request) (*T, error) {
//     log.Println("Processing JSON...")
//     var data T  // T should be a struct type, not a pointer.
//     decoder := json.NewDecoder(r.Body)
//     err := decoder.Decode(&data)
//     if err != nil {
//         log.Printf("Error decoding JSON: %v", err)
//         return nil, err
//     }
//     log.Printf("Processed JSON: %+v", data)
//     return &data, nil  // Return a pointer to the data.
// }


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