package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"webhook_test_server/model"
)

// OrderLineCancelledHandler handles order line cancelled events
func (h *WebhookHandler) OrderCreatedEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Creation event ")
	var event model.OrderCreated
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order created event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Created event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

// OrderLineCancelledHandler handles order line cancelled events
func (h *WebhookHandler) OrderCreationFailedEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Creation Failed event ")
	var event model.OrderCreationFailed
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order creation failed event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Creation Failed event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

// OrderLineCancelledHandler handles order line cancelled events
func (h *WebhookHandler) OrderLineCancelledEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Line Cancelled event")
	var event model.OrderLineCancelled
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order line cancelled event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Line Cancelled event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

// OrderLineRefundedHandler handles order line refunded events
func (h *WebhookHandler) OrderLineRefundedEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Line Refunded event")
	var event model.OrderLineRefunded
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order line refunded event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Line Refunded event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

// OrderLineShippedHandler handles order line shipped events
func (h *WebhookHandler) OrderLineShippedEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Line Shipped event ")
	var event model.OrderLineShipped
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order line shipped event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Line Shipped event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

// OrderLineShippingDeletedHandler handles order line shipping deleted events
func (h *WebhookHandler) OrderLineShippingDeletedEventHandle(marketplace string, body []byte) error {
	log.Printf("Processing Order Line Shipping Deleted event")
	var event model.OrderLineShippingDeleted
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode order line shipping deleted event: %w", err)
	}
	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}

	log.Printf("Storing Order Line Shipping Deleted event for marketplace: %s, External Order ID: %s", marketplace, event.ExternalOrderID)
	return h.db.StoreOrderEventData(h.tableNames[0], event.Type, event.ExternalOrderID, event.LastUpdated, marketplace, event)
}

func (h *WebhookHandler) HandleVariantStockUpdated(marketplace string, body []byte) error {
	log.Printf("Processing handle Variant Stock Updated event")
	var event model.VariantStockUpdated
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode Variant Stoc kUpdated event: %w", err)
	}
	log.Printf("Processing handle Variant Stock Updated for marketplace: %s, Event ID: %s , deal ID: %s", marketplace, event.EventId, event.DealID)

	// Validate the struct to make sure all required fields are present and correct
	if err := validateByType(&event); err != nil {
		log.Printf("Validation error for Order Created event: %v", err)
		return fmt.Errorf("validation error for order created event: %w", err)
	}
	// Create an instance of EventOptions
	opts := model.EventOptions{}
	if event.DealID != "" { // Check if ExternalOrderID is non-empty.
		opts.DealId = &event.DealID // If non-empty, set it in the options.
	}

	return h.db.StoreEventData(h.tableNames[1], event.Type, event.EventId, event.LastUpdated, marketplace, event, opts)
}