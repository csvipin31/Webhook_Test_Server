package handler

import (
	"fmt"
	"log"
	"net/http"
)

// Health Check : ReadyHandler, LiveHandler, HealthHandler
func ReadyHandler(w http.ResponseWriter, r *http.Request) error {
	handlerName := "ReadyHandler"
	startTime, method, url := logRequestStart(r, handlerName)
	log.Println("ReadyHandler:")

	if r.Method != http.MethodGet {
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method Not Allowed"), "Only GET requests are accepted.")
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
		return NewAPIError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"), "Only GET allowed, using wrong method TYPE")
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