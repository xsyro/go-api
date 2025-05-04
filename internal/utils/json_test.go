package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsyro/goapi/internal/utils"
)

func TestWriteJson(t *testing.T) {
	p := struct {
		Test string `json:"test"`
	}{
		Test: "value",
	}

	rr := httptest.NewRecorder()
	err := utils.WriteJson(rr, p)
	assert.NoError(t, err)
	assert.Contains(t, rr.Body.String(), `"test":"value"`)
}

func TestJsonError(t *testing.T) {
	code := http.StatusBadRequest // 400
	message := "This is an error message"
	rr := httptest.NewRecorder()

	utils.JsonError(rr, code, message)
	assert.Equal(t, code, rr.Code, "Expected status code to match")

	expectedBody := `{"error":"This is an error message"}`
	assert.JSONEq(t, expectedBody, rr.Body.String(), "Expected body to contain correct error message")
}
