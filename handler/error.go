package handler

import (
	"encoding/json"
	"fmt"
	"log" 
	"net/http"
	"reflect"
)

type APIError struct {
	StatusCode int    `json:"status_code"`
	Cause      string `json:"error"`
	Message    string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("api error: %d - %s", e.StatusCode, e.Cause)
}

func NewAPIError(statusCode int, err error, message string) APIError {
	return APIError{
		StatusCode: statusCode,
		Cause:      err.Error(),
		Message:    message,
	}
}

func InvalidRequestData(errors map[string]string, message string) APIError {
	errorDetails, _ := json.Marshal(errors) 
	return APIError{
		StatusCode: http.StatusUnprocessableEntity,
		Cause:      "Validation Error",
		Message:    message + " "+ string(errorDetails) ,
	}
}

func InvalidJson() APIError {
	return NewAPIError(http.StatusBadRequest, fmt.Errorf("invalid JSON Request Data"), "Please check JSON format")
}

type APIfunc func(w http.ResponseWriter, r *http.Request) error

func Make(h APIfunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			log.Printf("HTTP API Error: %v, Path: %s", err, r.URL.Path)
			switch err := err.(type) {
			case APIError:
				writeJSON(w, err.StatusCode, err)
			default:
				writeJSON(w, http.StatusInternalServerError, APIError{StatusCode: http.StatusInternalServerError, Cause: "Internal Server Error", Message: "An unexpected error has occurred."})
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

func sendAPIError(w http.ResponseWriter, apiErr APIError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(apiErr.StatusCode)
    json.NewEncoder(w).Encode(apiErr)
}

func validateByType(data interface{}) error {
    v := reflect.ValueOf(data).Elem()  // Ensure data is a pointer to a struct

    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        typeOfField := field.Type()

        // Check for zero value based on type dynamically
        if field.IsZero() {
            return fmt.Errorf("field %s is required and cannot be zero", v.Type().Field(i).Name)
        }

        // You can add more type-specific validations here if necessary
        switch typeOfField.Kind() {
        case reflect.String:
            // Example of additional validation for strings
            if field.Len() == 0 {
                return fmt.Errorf("field %s must be a non-empty string", v.Type().Field(i).Name)
            }
        case reflect.Int, reflect.Int32, reflect.Int64:
            // Example validation for integers if needed
        case reflect.Struct:
            // Recursive validation for nested structs
            err := validateByType(field.Addr().Interface())
            if err != nil {
                return err
            }
        }
    }

    return nil
}