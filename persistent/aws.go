package persistent

import (
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sts"
)

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
