package handlers

import(
	"net/http"
	"time"
	"encoding/json"
)

type ErrorResponse struct {
    Timestamp string `json:"timestamp"`
    Error     string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    resp := ErrorResponse{
        Timestamp: time.Now().Format(time.RFC3339),
        Error:     err.Error(),
    }
    json.NewEncoder(w).Encode(resp)
}
