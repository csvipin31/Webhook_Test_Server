package persistent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type OrderEvent struct {
    EventType       string `json:"eventType"`
    ExternalOrderID string `json:"externalOrderID"`
    LastUpdated     string `json:"lastUpdated"`
    PK              string `json:"pk"`
    SK              string `json:"sk"`
    EventData       string `json:"eventData"`
}

func ConvertDynamoItemToOrderEvent(item map[string]*dynamodb.AttributeValue) (OrderEvent, error) {
    var event OrderEvent
    err := dynamodbattribute.UnmarshalMap(item, &event)
    return event, err
}



func loadConfig(filename string) (*Config, error) {
	// Convert relative path to absolute path for clarity
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error getting absolute file path: %s\n", err)
		return nil, err
	}
	fmt.Printf("Reading configuration from: %s\n", absolutePath)

	// Read the file
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return nil, err
	}

	// Unmarshal JSON data
	var config Config
	if err = json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing JSON data: %s\n", err)
		return nil, err
	}

	return &config, nil
}