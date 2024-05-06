package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
        mock.AnythingOfType("*model.UserMessageData")).Return(nil)


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
