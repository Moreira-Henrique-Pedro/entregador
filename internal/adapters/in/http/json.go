package http

import (
	"encoding/json"
	"net/http"
)

const maxRequestBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

// decodeJSON writes a 400 and returns false when the body is not a valid T.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	var request T
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body: " + err.Error()})
		return nil, false
	}
	return &request, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
