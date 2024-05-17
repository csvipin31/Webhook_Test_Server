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
)

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

// TODO : Remove later
/*
func unMarshallJSON(body []byte, target interface{}) error {
	log.Println("Processing JSON...")
	// Log the raw JSON data received by the server
	log.Printf("Received JSON: %s", string(body))

	if err := json.Unmarshal(body, &target); err != nil {
		log.Printf("Failed to decode order created event: " + err.Error())
		return err
	}
	log.Printf("Processed JSON: %+v", target)
	return nil
}
*/

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