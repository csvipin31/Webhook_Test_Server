package persistent

import (
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
