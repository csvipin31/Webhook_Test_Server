package handler

import (
    "net/http"
    "log"
)

// SetupRoutes configures the HTTP server routes
func SetupRoutes(webhookHandler *WebhookHandler) {
    http.HandleFunc("/ready", Make(ReadyHandler))
    http.HandleFunc("/live", Make(LiveHandler))
    http.HandleFunc("/health", Make(HealthHandler))
    http.HandleFunc("/dbhealth", Make(webhookHandler.DBHealthHandler))
    http.HandleFunc("/", Make(webhookHandler.WebhookEvents))

    // Log route configuration 
    log.Println("HTTP routes configured successfully.")
}
