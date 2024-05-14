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

	// Define the environment variable keys
    envVars := []string{
        "DYNAMODB_ORDER_TABLE_NAME",
        "DYNAMODB_PRODUCT_TABLE_NAME",
    }

    // Load the table names from environment variables
    tableNames := LoadTableNames(envVars...)

    // Example usage: Print the loaded table names
    for _, tableName := range tableNames {
        log.Printf("Loaded table name: %s", tableName)
    }
    if err := db.InitializeTables(tableNames); err != nil {
        log.Fatalf("failed to initialize tables:%s: %v", tableNames,err)
    }


	// Create the webhook handler with the database dependency
    webhookHandler := handler.NewWebhookHandler(db,tableNames)
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


func LoadTableNames(envVars ...string) []string {
    var tableNames []string

    for _, envVar := range envVars {
        tableName := os.Getenv(envVar)
        if tableName != "" {
            tableNames = append(tableNames, tableName)
        }
    }

    if len(tableNames) == 0 {
        log.Fatalf("No valid table names provided in environment variables")
    }

    return tableNames
}