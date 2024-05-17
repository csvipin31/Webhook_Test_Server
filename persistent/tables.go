package persistent

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

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