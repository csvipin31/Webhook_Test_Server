package persistent

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

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