package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"webhook_test_server/handler"
	"webhook_test_server/model"

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

func (m *MockDB) StoreData(tableName, pKey string, data interface{}) error {
    args := m.Called(tableName, pKey, data)
    return args.Error(0)
}

func (m *MockDB) DescribeTable(tableName string) error {
    args := m.Called(tableName)
    return args.Error(0)
}

func (m *MockDB) InitializeTables(tableName string) error {
    args := m.Called(tableName)
    return args.Error(0)
}

func (m *MockDB) StoreEventData(tableName, eventType, eventId, lastUpdated, merchantId string, eventData interface{}, opts model.EventOptions) error {
    args := m.Called(tableName, eventType, eventId, lastUpdated, merchantId, eventData,opts)
    return args.Error(0)
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

     // Setting up the expected call with mock for CreateTableIfNotExists
    db.On("CreateTableIfNotExists", "My_Table").Return(nil)

    // Setting up the expected call with mock
    db.On("StoreData", 
        "My_Table", 
        "PK#MerchantId:45", 
        mock.AnythingOfType("model.UserMessageData")).Return(nil)


    handler := handler.NewWebhookHandler(db)

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
func TestWebhookEvents(t *testing.T) {
    db := new(MockDB)

    // Initialize the handler
    handler := handler.NewWebhookHandler(db)

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
        "EventWebhook",
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
    handler.HandleEventWebhook(w, req)

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

    // Initialize the handler
    handler := handler.NewWebhookHandler(db)

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
    db.On("StoreEventData",
        "EventWebhook",
        "order/created",
        "48b4a0d1-2a95-4308-9a45-00c65b6e70e4",
        "2024-05-03T03:48:13.506Z",
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



// TestDBHealthHandler tests the database health check endpoint
func TestDBHealthHandlerOk(t *testing.T) {
    db := new(MockDB)
    db.On("DescribeTable", "My_Table").Return(nil) // Simulate a healthy database

    handler := handler.NewWebhookHandler(db)
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
    db.On("DescribeTable", "My_Table").Return(errors.New("database error")) // Simulate an unhealthy database

    handler := handler.NewWebhookHandler(db)
    req := httptest.NewRequest("GET", "/dbhealth", nil)
    w := httptest.NewRecorder()

    handler.DBHealthHandler(w, req)

    res := w.Result()
    defer res.Body.Close()
    assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "Expected internal server error status")

    db.AssertExpectations(t)
}
