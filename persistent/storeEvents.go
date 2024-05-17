package persistent

import (
	"encoding/json"
	"fmt"
	"log"
	"webhook_test_server/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// StoreData stores data in a specified DynamoDB table
func (db *Database) StoreData(tableName, pKey string, data interface{}) error {
	// First, marshal the data into a map[string]*dynamodb.AttributeValue
	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return err
	}

	// Add the primary key to the attribute value map
	av["PrimaryKey"] = &dynamodb.AttributeValue{S: aws.String(pKey)}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Perform the PutItem operation
	_, err = db.svc.PutItem(input)
	if err != nil {
		log.Printf("Failed to put item in table %s: %v", tableName, err)
		return err
	}

	log.Printf("Data successfully stored in table %s", tableName)
	return nil
}

/*
func (db *Database) CreateEventsTableIfNotExists(tableName string) error {
	log.Printf("CreateEventsTableIfNotExists")
	// Check if the table already exists
	exists, err := db.tableExists(tableName)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Table %s already exists", tableName)
		return nil
	}

	// Define table attributes and schema
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: aws.String("S"), // Partition key
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: aws.String("S"), // Sort key
			},
			{
				AttributeName: aws.String("DealId"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("ExternalOrderId"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("DealIdIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("DealId"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("SK"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
			{
				IndexName: aws.String("ExternalOrderIdIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("ExternalOrderId"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("SK"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
		},
	}

	// Create the table
	_, err = db.svc.CreateTable(input)
	if err != nil {
		return err
	}
	log.Printf("Table %s created successfully", tableName)
	return nil
}
*/
// StoreData stores data in the WebhookEvents table in DynamoDB.

func (db *Database) StoreEventData(tableName, eventType, eventId, lastUpdated, merchantId string, eventData interface{}, opts model.EventOptions) error {
	log.Printf("StoreEventData")
	// Prepare the primary key and sort key
	pk := fmt.Sprintf("PK%s#%s#%s", merchantId, eventType, eventId)
	sk := fmt.Sprintf("SK%s", lastUpdated)

	// Marshal the entire event data into a JSON string for the EventData attribute
	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}

	// Prepare the attribute values for DynamoDB
	item := map[string]*dynamodb.AttributeValue{
		"PK":        {S: aws.String(pk)},
		"SK":        {S: aws.String(sk)},
		"EventID":   {S: aws.String(eventId)},
		"EventType": {S: aws.String(eventType)},
		"EventData": {S: aws.String(string(eventDataJSON))},
	}

	// Add DealId and ExternalOrderId to the item if available
	if opts.DealId != nil {
		item["DealId"] = &dynamodb.AttributeValue{S: aws.String(*opts.DealId)}
	}
	if opts.ExternalOrderId != nil {
		item["ExternalOrderId"] = &dynamodb.AttributeValue{S: aws.String(*opts.ExternalOrderId)}
	}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}

	// Perform the PutItem operation
	_, err = db.svc.PutItem(input)
	if err != nil {
		log.Printf("Failed to put item in table :%v, %v", tableName, err)
		return err
	}

	log.Printf("Data successfully stored in table : %v", tableName)
	return nil
}

// StoreData stores data in the WebhookEvents table in DynamoDB.
func (db *Database) StoreOrderEventData(tableName, eventType, externalOrderId, lastUpdated, merchantId string, eventData interface{}) error {
	log.Printf("StoreEventData")
	// Prepare the primary key and sort key
	pk := fmt.Sprintf("#PK#%s#%s", merchantId, externalOrderId)
	sk := fmt.Sprintf("#SK#%s#%s", lastUpdated, eventType)

	// Marshal the entire event data into a JSON string for the EventData attribute
	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}

	// Prepare the attribute values for DynamoDB
	item := map[string]*dynamodb.AttributeValue{
		"PK":              {S: aws.String(pk)},
		"SK":              {S: aws.String(sk)},
		"ExternalOrderId": {S: aws.String(externalOrderId)},
		"LastUpdated":     {S: aws.String(lastUpdated)},
		"EventType":       {S: aws.String(eventType)},
		"EventData":       {S: aws.String(string(eventDataJSON))},
	}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}

	// Perform the PutItem operation
	_, err = db.svc.PutItem(input)
	if err != nil {
		log.Printf("Failed to put item in table :%v, %v", tableName, err)
		return err
	}

	log.Printf("Data successfully stored in table : %v", tableName)
	return nil
}