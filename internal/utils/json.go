package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, i interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(i)
}

func JsonError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	json.NewEncoder(w).Encode(response)
}
