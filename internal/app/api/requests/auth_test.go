package requests

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidation(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("nameWithSpace", nameWithSpace)

	tests := []struct {
		name             string
		structToValidate interface{}
		expectError      bool
	}{
		{"Valid BasicAuth", BasicAuth{Email: "email@example.com", Password: "password123"}, false},
		{"Invalid BasicAuth Email", BasicAuth{Email: "notanemail", Password: "password123"}, true},
		{"Invalid BasicAuth Password", BasicAuth{Email: "email@example.com", Password: "short"}, true},
		{"Valid SignupRequest", Signup{BasicAuth: BasicAuth{Email: "email@example.com", Password: "password123"}, Name: "John Doe"}, false},
		{"Invalid SignupRequest Name", Signup{BasicAuth: BasicAuth{Email: "email@example.com", Password: "password123"}, Name: "JohnDoe"}, true},
		{"Valid RefreshRequest", Refresh{Token: "somevalidtoken"}, false},
		{"Invalid RefreshRequest", Refresh{Token: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.structToValidate)
			if (err != nil) != tt.expectError {
				t.Errorf("Validation for %s failed, expected error: %v, got: %v", tt.name, tt.expectError, err != nil)
			}
		})
	}
}
