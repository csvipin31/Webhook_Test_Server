package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"webhook_test_server/handler"
	"webhook_test_server/persistent"

	"github.com/joho/godotenv"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("SERVER_PORT")
	log.Println("server listening on port: ", port)

	// Initialize the database
	db, err := persistent.NewDatabase()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize tables on startup
	tableName := os.Getenv("DYNAMODB_ORDER_TABLE_NAME")
    if err := db.InitializeTables(tableName); err != nil {
        log.Fatalf("failed to initialize tables:%s: %v", tableName,err)
    }


	// Create the webhook handler with the database dependency
    webhookHandler := handler.NewWebhookHandler(db,tableName)
    handler.SetupRoutes(webhookHandler)

	log.Printf("Server starting on port: %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("failed to start HTTP server %v", err)
	}

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	}
}