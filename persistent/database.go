package persistent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"webhook_test_server/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sts"
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

func (db *Database) InitializeTables(tableNames []string) error {
	log.Printf("Initialize the dynamodb Tables")
	config, err := loadConfig("persistent/table.json")
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	ReplaceTableNames(config, tableNames)
	// Example for a single table, repeat for others or make it dynamic based on configuration
	for _, tableConfig := range config.Tables {
		err := db.CreateEventsTableIfNotExist(tableConfig)
		if err != nil {
			log.Printf("Failed to create table %s: %s", tableConfig.TableName, err)
		}
	}
	return nil
}

// ReplaceTableNames replaces the table names in the JSON configuration with the table names from the tableNames slice.
func ReplaceTableNames(config *Config, tableNames []string) {
	if len(config.Tables) != len(tableNames) {
		log.Fatalf("The number of table names in the environment does not match the number of tables in the JSON configuration")
	}

	for i := range config.Tables {
		config.Tables[i].TableName = tableNames[i]
	}
}

// TokenFetcher is a custom implementation of the TokenFetcher interface.
type TokenFetcher struct {
	webIdentityToken string
}

// FetchToken returns the web identity token bytes.
func (tf *TokenFetcher) FetchToken(_ credentials.Context) ([]byte, error) {
	return []byte(tf.webIdentityToken), nil
}

// CustomProvider is a custom implementation of credentials.Provider that wraps the *stscreds.WebIdentityRoleProvider.
type CustomProvider struct {
	provider *stscreds.WebIdentityRoleProvider
}

// Retrieve returns the AWS credentials.
func (p *CustomProvider) Retrieve() (credentials.Value, error) {
	return p.provider.Retrieve()
}

// IsExpired returns whether the underlying credentials are expired or not.
func (p *CustomProvider) IsExpired() bool {
	return p.provider.IsExpired()
}

// CheckAWSRoleAvailability checks if the AWS role is available.
func CheckAWSRoleAvailability() bool {
	myRoleArn := os.Getenv("AWS_ROLE_ARN")
	log.Println("### my Role Arn.", myRoleArn)
	if myRoleArn == "" {
		return false
	}

	sess := session.Must(session.NewSession())
	log.Println("sess", sess)
	secret := stscreds.NewCredentials(sess, myRoleArn)
	stsSvc := sts.New(sess, &aws.Config{Credentials: secret})
	input := &sts.GetCallerIdentityInput{}
	_, err := stsSvc.GetCallerIdentity(input)
	if err != nil {
		log.Println("error", err)
	}
	return err == nil
}

// ConnectToAWSDynamoDB connects to AWS DynamoDB.
func ConnectToAWSDynamoDB() (*dynamodb.DynamoDB, error) {
	roleARN := os.Getenv("AWS_ROLE_ARN")
	webIdentityTokenPath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	if roleARN == "" || webIdentityTokenPath == "" {
		return nil, errors.New("AWS_ROLE_ARN or AWS_WEB_IDENTITY_TOKEN_FILE is not set")
	}

	// Read the web identity token from the file
	webIdentityToken, err := os.ReadFile(webIdentityTokenPath)
	if err != nil {
		log.Println("Error reading the web identity token file:", err)
		return nil, err
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a new STS client to perform AWS STS operations
	stsClient := sts.New(sess)

	// Create a TokenFetcher instance with the web identity token
	tokenFetcher := &TokenFetcher{
		webIdentityToken: string(webIdentityToken),
	}

	// Create a custom AWS credentials provider using the web identity token and role ARN
	provider := &CustomProvider{
		provider: stscreds.NewWebIdentityRoleProviderWithOptions(stsClient, roleARN, "WebIdentitySession", tokenFetcher),
	}

	credsValue, err := provider.Retrieve()
	if err != nil {
		log.Println("Error retrieving AWS credentials:", err)
		return nil, err
	}

	// Print the AWS credentials obtained through web identity federation
	log.Println("Access Key ID:", credsValue.AccessKeyID)
	// log.Println("Secret Access Key:", credsValue.SecretAccessKey)
	// log.Println("Session Token:", credsValue.SessionToken)

	dynamoDBClient := dynamodb.New(sess, &aws.Config{Credentials: credentials.NewCredentials(provider)})

	// For example, scan the table
	result, err := dynamoDBClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("My_Table"),
	})
	if err != nil {
		log.Println("Error scanning table:", err)
		return nil, err
	}

	log.Println("Items:")
	for _, item := range result.Items {
		log.Println(item)
	}

	return dynamoDBClient, nil
}

// ConnectToLocalDynamoDB connects to a local DynamoDB instance.
func ConnectToLocalDynamoDB() (*dynamodb.DynamoDB, error) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	region := os.Getenv("DYNAMODB_REGION")
	log.Println("### LOCAL_DYNAMODB.", region)
	log.Println("### LOCAL_DYNAMODB.", endpoint)
	if endpoint == "" {
		endpoint = "http://localhost:8000" // Default local endpoint
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

// CreateTableIfNotExists checks if a table exists and creates it if it does not
func (db *Database) CreateTableIfNotExists(tableName string) error {
	// First, check if the table already exists
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
				AttributeName: aws.String("PrimaryKey"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PrimaryKey"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
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

// tableExists checks the existence of a table
func (db *Database) tableExists(tableName string) (bool, error) {
	input := &dynamodb.ListTablesInput{}

	// Loop through all tables in the account to check for existence
	for {
		result, err := db.svc.ListTables(input)
		if err != nil {
			return false, err
		}
		for _, name := range result.TableNames {
			if *name == tableName {
				return true, nil
			}
		}
		// Check if there are more tables beyond the returned set
		if result.LastEvaluatedTableName == nil {
			break
		}
		input.ExclusiveStartTableName = result.LastEvaluatedTableName
	}

	return false, nil
}

// DescribeTable checks details of a specified table
func (db *Database) DescribeTable(tableName string) error {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}

	result, err := db.svc.DescribeTable(input)
	if err != nil {
		log.Printf("Error describing table %s: %v", tableName, err)
		return err
	}

	// Output some of the important information about the table
	table := result.Table
	log.Printf("Table Description for %s:", tableName)
	log.Printf("Status: %s", *table.TableStatus)
	log.Printf("Item Count: %d", *table.ItemCount)
	log.Printf("Provisioned Read Capacity Units: %d", *table.ProvisionedThroughput.ReadCapacityUnits)
	log.Printf("Provisioned Write Capacity Units: %d", *table.ProvisionedThroughput.WriteCapacityUnits)

	return nil
}

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

func (db *Database) CreateEventsTableIfNotExist(config TableConfig) error {
	log.Printf("CreateEventsTableIfNotExists")
	// Check if the table already exists
	exists, err := db.tableExists(config.TableName)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Table %s already exists", config.TableName)
		return nil
	}

	// Ensure each GSI has a properly defined Projection
	for i := range config.GlobalSecondaryIndexes {
		if config.GlobalSecondaryIndexes[i].Projection == nil {
			config.GlobalSecondaryIndexes[i].Projection = &dynamodb.Projection{
				ProjectionType: aws.String("ALL"), // or "KEYS_ONLY", "INCLUDE"
				// Uncomment and specify the attributes if ProjectionType is "INCLUDE"
				// NonKeyAttributes: aws.StringSlice([]string{"Attribute1", "Attribute2"}),

			}
		}

		// Check and set default Provisioned Throughput if it's not specified
		if config.GlobalSecondaryIndexes[i].ProvisionedThroughput == nil {
			config.GlobalSecondaryIndexes[i].ProvisionedThroughput = &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10), // Default read capacity
				WriteCapacityUnits: aws.Int64(10), // Default write capacity
			}
		}
	}

	// Create the table
	_, err = db.svc.CreateTable(&dynamodb.CreateTableInput{
		TableName:              aws.String(config.TableName),
		AttributeDefinitions:   config.AttributeDefinitions,
		KeySchema:              config.KeySchema,
		GlobalSecondaryIndexes: config.GlobalSecondaryIndexes,
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
			WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
		},
	})

	if err != nil {
		return err
	}

	log.Printf("Table %s created successfully", config.TableName)
	return nil
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

func (db *Database) FetchByPrimaryKey(tableName, pk string) (*dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("#pk = :pkval"),
		ExpressionAttributeNames: map[string]*string{
			"#pk": aws.String("PK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pkval": {
				S: aws.String(pk),
			},
		},
		ScanIndexForward: aws.Bool(false), // Set to false if you want to sort in descending order
	}

	result, err := db.svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch items by primary key without SK: %w", err)
	}

	return result, nil
}

func (db *Database) FetchByGSI(tableName, gsiName string, keyConditions map[string]*dynamodb.Condition) (*dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		TableName:     aws.String(tableName),
		IndexName:     aws.String(gsiName),
		KeyConditions: keyConditions,
	}

	result, err := db.svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item by GSI: %w", err)
	}

	return result, nil
}

func (db *Database) QueryOrderEventsByExternalOrderId(tableName, externalOrderId string) (*dynamodb.QueryOutput, error) {
	keyConditions := map[string]*dynamodb.Condition{
		"ExternalOrderId": {
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{
				{
					S: aws.String(externalOrderId),
				},
			},
		},
	}

	return db.FetchByGSI(tableName, "ExternalOrderIdIndex", keyConditions)
}
