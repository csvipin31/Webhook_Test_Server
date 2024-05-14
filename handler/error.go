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

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}


func validateByType(data interface{}) error {
    v := reflect.ValueOf(data)
    if v.Kind() != reflect.Ptr || v.IsNil() {
        return fmt.Errorf("data must be a non-nil pointer")
    }

    v = v.Elem()
    if v.Kind() != reflect.Struct {
        return fmt.Errorf("data must be a pointer to a struct")
    }

    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        fieldName := v.Type().Field(i).Name

        if field.IsZero() {
            return fmt.Errorf("field %s is required and cannot be zero", fieldName)
        }

        switch field.Kind() {
        case reflect.String:
            if field.Len() == 0 {
                return fmt.Errorf("field %s must be a non-empty string", fieldName)
            }
        case reflect.Struct:
            if err := validateByType(field.Addr().Interface()); err != nil {
                return err
            }
        }
    }

    return nil
}