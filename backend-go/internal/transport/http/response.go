package httptransport

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON отправляет JSON-ответ с нужным кодом.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError отправляет ошибку в JSON.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
