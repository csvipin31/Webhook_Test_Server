package handler

import "webhook_test_server/persistent"

type WebhookHandler struct {
	db            persistent.DatabaseInterface
	tableNames    []string
	eventHandlers map[string]func(string, []byte) error
}

func NewWebhookHandler(db persistent.DatabaseInterface, tableNames []string) *WebhookHandler {
	handler := &WebhookHandler{
		db:            db,
		tableNames:    tableNames,
		eventHandlers: make(map[string]func(string, []byte) error),
	}
	handler.registerEventHandlers()
	return handler
}

// registerEventHandlers registers all event handlers
func (h *WebhookHandler) registerEventHandlers() {
	h.eventHandlers["order/created"] = h.OrderCreatedEventHandle
	h.eventHandlers["order/creation-failed"] = h.OrderCreationFailedEventHandle
	h.eventHandlers["order-line/cancelled"] = h.OrderLineCancelledEventHandle
	h.eventHandlers["order-line/refunded"] = h.OrderLineRefundedEventHandle
	h.eventHandlers["order-line/shipped"] = h.OrderLineShippedEventHandle
	h.eventHandlers["order-line/shipping-deleted"] = h.OrderLineShippingDeletedEventHandle
	h.eventHandlers["variant/stock-updated"] = h.HandleVariantStockUpdated
}