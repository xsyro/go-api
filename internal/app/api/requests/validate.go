package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Single instance of Validate, to cache struct info
var validate = validator.New(validator.WithRequiredStructEnabled())

// Custom validation function for at space in `name`
func nameWithSpace(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`\w+\s+\w+`).MatchString(fl.Field().String())
}

func validateStruct(s interface{}) map[string]string {
	validate.RegisterValidation("nameWithSpace", nameWithSpace)
	errs := make(map[string]string)
	err := validate.Struct(s)
	if err != nil {
		for _, ve := range err.(validator.ValidationErrors) {
			switch ve.Tag() {
			case "required":
				errs[ve.Field()] = fmt.Sprintf("%s is required", ve.Field())
			case "email":
				errs[ve.Field()] = "Valid email required"
			case "min":
				errs[ve.Field()] = fmt.Sprintf("%s should be at least %s chars", ve.Field(), ve.Param())
			case "nameWithSpace":
				errs[ve.Field()] = fmt.Sprintf("%s must include a space", ve.Field())
			default:
				errs[ve.Field()] = fmt.Sprintf("%s is not valid", ve.Field())
			}
		}
		return errs
	}
	return nil
}

func DecodeValidRequest[T any](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := validateStruct(v); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}
