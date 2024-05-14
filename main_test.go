package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"webhook_test_server/handler"
	"webhook_test_server/model"
	"webhook_test_server/persistent"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for the DatabaseInterface
type MockDB struct {
    mock.Mock
}

func (m *MockDB) ConnectToDatabase() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockDB) Close() {
    m.Called()
}

func (m *MockDB) CreateTableIfNotExists(tableName string) error {
    args := m.Called(tableName)
    return args.Error(0)
}

func (m *MockDB) CreateEventsTableIfNotExist(config persistent.TableConfig) error {
    args := m.Called(config)
    return args.Error(0)
}

func (m *MockDB) StoreData(tableName, pKey string, data interface{}) error {
    args := m.Called(tableName, pKey, data)
    return args.Error(0)
}

func (m *MockDB) DescribeTable(tableName string) error {
    args := m.Called(tableName)
    return args.Error(0)
}

func (m *MockDB) InitializeTables(tableName []string) error {
    args := m.Called(tableName)
    return args.Error(0)
}

func (m *MockDB) StoreEventData(tableName, eventType, eventId, lastUpdated, merchantId string, eventData interface{}, opts model.EventOptions) error {
    args := m.Called(tableName, eventType, eventId, lastUpdated, merchantId, eventData,opts)
    return args.Error(0)
}

func (m *MockDB) StoreOrderEventData(tableName, eventType, externalOrderId, lastUpdated, merchantId string, eventData interface{}) error {
    args := m.Called(tableName, eventType, externalOrderId, lastUpdated, merchantId, eventData)
    return args.Error(0)
}
func (m *MockDB) FetchByPrimaryKey(tableName, pk string) (*dynamodb.QueryOutput, error) {
    args := m.Called(tableName, pk)
    return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (m *MockDB) FetchByGSI(tableName, gsiName string, keyConditions map[string]*dynamodb.Condition) (*dynamodb.QueryOutput, error) {
    args := m.Called(tableName, gsiName, keyConditions)
    return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (m *MockDB) QueryOrderEventsByExternalOrderId(tableName, externalOrderId string) (*dynamodb.QueryOutput, error) {
    args := m.Called(tableName, externalOrderId)
    return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}


// TestHandleWebhook tests the webhook handler function
func TestHandleWebhook(t *testing.T) {
    db := new(MockDB)
   
    // Create a UserMessageData instance with the data you expect to receive
    userMessageData := model.UserMessageData{
        Token: "wew1212121xewewbdgfhgdf",
        AgreementStatus: "INACTIVE",
        Reason: []string{"test reason", "reason 2"},
        UserMessage: "Your Payment method is expired",
    }
    jsonData, err := json.Marshal(userMessageData)
    if err != nil {
        t.Fatal(err) // Handle errors with JSON marshaling
    }

    tableNames := []string{"My_Table"}
    fmt.Println("TableName before mock setup:", tableNames[0]) 
     // Setting up the expected call with mock for CreateTableIfNotExists
    db.On("CreateTableIfNotExists", tableNames[0]).Return(nil)
    // Setting up the expected call with mock
    db.On("StoreData", 
        tableNames[0], 
        "PK#MerchantId:45", 
        mock.AnythingOfType("model.UserMessageData")).Return(nil)

    
    handler := handler.NewWebhookHandler(db,tableNames)
   
    // Setting up a request
    req := httptest.NewRequest("POST", "/45", bytes.NewReader(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Calling the handler
    handler.HandleWebhook(w, req)

    // Check response
    res := w.Result()
    defer res.Body.Close()
    assert.Equal(t, http.StatusOK, res.StatusCode, "Expected status OK")

    /// Check that the mock was called as expected
    db.AssertExpectations(t)
}

// TestWebhookEvents tests the webhook handler function
func TestWebhookVariantStockUpdateEvents(t *testing.T) {
    t.Skip()
    db := new(MockDB)
    tableNames := []string{"EventWebhook"}

    // Initialize the handler
    handler := handler.NewWebhookHandler(db,tableNames)

    // Setup a sample dynamic event for testing
    variantStockUpdatedEvent := model.VariantStockUpdated{
        BaseEvent: model.BaseEvent{
            Type:       "variant/stock-updated",
            EventId:    "529c8a0d-4b85-495a-a54c-6031995d9c2a",
            LastUpdated: "2024-05-07T01:47:00.138Z",
        },
        DealID:    "378397",
        VariantID: nil,
        Stock:     16,
    }

    jsonData, err := json.Marshal(variantStockUpdatedEvent)
    if err != nil {
        t.Fatal(err) // Handle errors with JSON marshaling
    }
    log.Printf("jsonData %s", jsonData)

    // Mock expected database interactions
    db.On("StoreEventData",
        tableNames[0],
        "variant/stock-updated",
        "529c8a0d-4b85-495a-a54c-6031995d9c2a",
        "2024-05-07T01:47:00.138Z",
        "BIGW",
        mock.Anything,
        mock.AnythingOfType("model.EventOptions")).Return(nil)

    // Setup a HTTP request for POST method
    req := httptest.NewRequest("POST", "/BIGW", bytes.NewReader(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Call the handler
    handler.WebhookEvents(w, req)

    // Check the response
    res := w.Result()
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK; got %v", res.StatusCode)
    }

    // Verify the body of the response
    body, _ := io.ReadAll(res.Body)
    expectedBody := "Success"
    if string(body) != expectedBody {
        t.Errorf("Expected body %s; got %s", expectedBody, string(body))
    }

    /// Check that the mock was called as expected
    db.AssertExpectations(t)
}

// TestWebhookEvents tests the webhook handler function
func TestWebhookOrderCreatedEvents(t *testing.T) {
    db := new(MockDB)
    tableNames := []string{"EventWebhook"}

    // Initialize the handler
    handler := handler.NewWebhookHandler(db,tableNames)

    // Setup a sample dynamic event for testing
    orderCreated := model.OrderCreated{
        BaseEvent: model.BaseEvent{
            Type:        "order/created",
            EventId:     "48b4a0d1-2a95-4308-9a45-00c65b6e70e4",
            LastUpdated: "2024-05-03T03:48:13.506Z",
        },
        ExternalOrderID: "auto-test-3aef291d-1bf0-41c3-9797-de544b1a41a2",
        Details: []model.OrderDetail{
            {
                ExternalOrderGroupID: "auto-test-3aef291d-1bf0-41c3-9797-de544b1a41a2",
                ExternalOrderLineID:  "auto-test-3aef291d-1bf0-41c3-9797-de544b1a41a2",
                Type:                 "Order",
                InternalID:           "137955620",
            },
        },
    }

    jsonData, err := json.Marshal(orderCreated)
    if err != nil {
        t.Fatal(err) // Handle errors with JSON marshaling
    }
    log.Printf("jsonData %s", jsonData)
    
    // Mock expected database interactions
    db.On("StoreOrderEventData",
        tableNames[0],
        "order/created",
        "auto-test-3aef291d-1bf0-41c3-9797-de544b1a41a2",
        "2024-05-03T03:48:13.506Z",
        "BIGW",
        mock.Anything).Return(nil)

    // Setup a HTTP request for POST method
    req := httptest.NewRequest("POST", "/BIGW", bytes.NewReader(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Call the handler
    handler.WebhookEvents(w, req)

    // Check the response
    res := w.Result()
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK; got %v", res.StatusCode)
    }

    // Verify the body of the response
    body, _ := io.ReadAll(res.Body)
    expectedBody := "Success"
    if string(body) != expectedBody {
        t.Errorf("Expected body %s; got %s", expectedBody, string(body))
    }

    /// Check that the mock was called as expected
    db.AssertExpectations(t)
}

// TestDBHealthHandler tests the database health check endpoint
func TestDBHealthHandlerOk(t *testing.T) {
    db := new(MockDB)
    tableNames := []string{"EventWebhook"}
    db.On("DescribeTable", tableNames[0]).Return(nil) // Simulate a healthy database

    handler := handler.NewWebhookHandler(db,tableNames)
    req := httptest.NewRequest("GET", "/dbhealth", nil)
    w := httptest.NewRecorder()

    handler.DBHealthHandler(w, req)

    // Check response
    res := w.Result()
    defer res.Body.Close()
    assert.Equal(t, http.StatusOK, res.StatusCode, "Expected status OK")

    db.AssertExpectations(t)
}

// TestDBHealthHandlerFail tests the scenario where the database is unhealthy
func TestDBHealthHandlerFail(t *testing.T) {
    db := new(MockDB)
    tableNames := []string{"EventWebhook"}
    db.On("DescribeTable", tableNames[0]).Return(errors.New("database error")) // Simulate an unhealthy database

    h := handler.NewWebhookHandler(db,tableNames)
    handlerFunc := handler.Make(h.DBHealthHandler)

    req := httptest.NewRequest("GET", "/dbhealth", nil)
    w := httptest.NewRecorder()

    handlerFunc(w, req)

    res := w.Result()
    defer res.Body.Close()
    assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "Expected internal server error status")

    db.AssertExpectations(t)
}

func TestNewAPIError(t *testing.T) {
    err := fmt.Errorf("test error")
    apiErr := handler.NewAPIError(http.StatusBadRequest, err, "A test error occurred.")

    if apiErr.StatusCode != http.StatusBadRequest {
        t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
    }

    if apiErr.Cause != "test error" {
        t.Errorf("Expected cause 'test error', got '%s'", apiErr.Cause)
    }

    if apiErr.Message != "A test error occurred." {
        t.Errorf("Expected message 'A test error occurred.', got '%s'", apiErr.Message)
    }
}


func TestFetchByPrimaryKey(t *testing.T) {
    mockDB := new(MockDB)
    tableName := "OrderEvents"
    pk := "#PK#BIGW#DB-Update1-770014-34f0-45b3-89b4-7b22fc4a43d1"

    expectedOutput := &dynamodb.QueryOutput{
        Items: []map[string]*dynamodb.AttributeValue{
            {
                "PK": {S: aws.String(pk)},
                // Add other attributes as needed
            },
        },
    }

    mockDB.On("FetchByPrimaryKey", tableName, pk).Return(expectedOutput, nil)

    result, err := mockDB.FetchByPrimaryKey(tableName, pk)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if len(result.Items) != 1 || *result.Items[0]["PK"].S != pk {
        t.Fatalf("Expected item with PK %s, got %v", pk, result.Items)
    }

    mockDB.AssertExpectations(t)
}