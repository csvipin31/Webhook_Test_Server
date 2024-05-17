package persistent

import (
	"log"
	"os"

	"webhook_test_server/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DatabaseInterface outlines the methods for database operations
type DatabaseInterface interface {
	ConnectToDatabase() error
	InitializeTables(tableName []string) error
	Close()
	CreateTableIfNotExists(tableName string) error
	CreateEventsTableIfNotExist(config TableConfig) error
	StoreData(tableName, pKey string, data interface{}) error
	DescribeTable(tableName string) error
	StoreEventData(tableName, eventType, eventId, lastUpdated, merchantId string, eventData interface{}, opts model.EventOptions) error
	StoreOrderEventData(tableName, eventType, externalOrderId, lastUpdated, merchantId string, eventData interface{}) error
	FetchByPrimaryKey(tableName, pk string) (*dynamodb.QueryOutput, error)
	FetchByGSI(tableName, gsiName string, keyConditions map[string]*dynamodb.Condition) (*dynamodb.QueryOutput, error)
	QueryOrderEventsByExternalOrderId(tableName, externalOrderId string) (*dynamodb.QueryOutput, error)
}

// Database represents the database connection.
type Database struct {
	svc *dynamodb.DynamoDB
}

type TableConfig struct {
	TableName              string                           `json:"tableName"`
	AttributeDefinitions   []*dynamodb.AttributeDefinition  `json:"attributeDefinitions"`
	KeySchema              []*dynamodb.KeySchemaElement     `json:"keySchema"`
	GlobalSecondaryIndexes []*dynamodb.GlobalSecondaryIndex `json:"globalSecondaryIndexes"`
	ReadCapacityUnits      int64                            `json:"readCapacityUnits"`
	WriteCapacityUnits     int64                            `json:"writeCapacityUnits"`
}

type Config struct {
	Tables []TableConfig `json:"tables"`
}

// NewDatabase creates a new database connection based on the environment configuration
func NewDatabase() (DatabaseInterface, error) {
	db := &Database{}
	err := db.ConnectToDatabase()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ConnectToLocalDynamoDB connects to a local DynamoDB instance.
func ConnectToLocalDynamoDB() (*dynamodb.DynamoDB, error) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	region := os.Getenv("DYNAMODB_REGION")
	log.Println("### LOCAL_DYNAMODB.", region)
	log.Println("### LOCAL_DYNAMODB.", endpoint)
	if endpoint == "" {
		endpoint = "http://localhost:8001" // Default local endpoint
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(region), // Replace with your desired region
		Endpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(
			"dummy",
			"dummy",
			""),
	}))

	return dynamodb.New(sess), nil
}

// ConnectToDatabase establishes a connection to DynamoDB, either locally or via AWS depending on the environment
func (db *Database) ConnectToDatabase() error {
	// Read role ARN and region from environment variables
	roleAvailable := CheckAWSRoleAvailability()
	log.Println("### ConnectToDatabase.", roleAvailable)

	var err error
	if roleAvailable {
		db.svc, err = ConnectToAWSDynamoDB()
	} else {
		db.svc, err = ConnectToLocalDynamoDB()
	}
	if err != nil {
		return err
	}
	log.Println("Database connected with signing region:", db.svc.SigningRegion)
	return nil
}

// Close terminates the connection to DynamoDB
func (db *Database) Close() {
	if db.svc != nil {
		db.svc.Client.Config.Credentials.Expire()
		db.svc.Client.Config.Credentials = nil
		db.svc.Client = nil
		db.svc = nil
	}
}
