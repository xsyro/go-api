package requests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock structure to use in the tests
type TestStruct struct {
	Name  string `json:"name" validate:"required,nameWithSpace"`
	Email string `json:"email" validate:"required,email"`
}

func TestDecodeValidRequest(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		payload     interface{} // Data to encode into the request body
		expectError bool        // Whether an error is expected
		errorFields []string    // Fields expected to have validation errors
	}{
		{
			name:        "Valid Request",
			payload:     TestStruct{Name: "John Doe", Email: "john@example.com"},
			expectError: false,
		},
		{
			name:        "Invalid Email",
			payload:     TestStruct{Name: "John Doe", Email: "invalid"},
			expectError: true,
			errorFields: []string{"Email"},
		},
		{
			name:        "Missing Name",
			payload:     TestStruct{Email: "john@example.com"},
			expectError: true,
			errorFields: []string{"Name"},
		},
		{
			name:        "No Space in Name",
			payload:     TestStruct{Name: "JohnDoe", Email: "john@example.com"},
			expectError: true,
			errorFields: []string{"Name"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the payload into JSON
			jsonPayload, _ := json.Marshal(tc.payload)
			// Create a new HTTP request with the JSON payload
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(jsonPayload))
			req.Header.Add("Content-Type", "application/json")

			// Decode and validate the request
			_, problems, err := DecodeValidRequest[TestStruct](req)

			if tc.expectError {
				assert.Error(t, err)
				for _, field := range tc.errorFields {
					assert.Contains(t, problems, field)
				}
			} else {
				assert.NoError(t, err)
				assert.Empty(t, problems)
			}
		})
	}
}
