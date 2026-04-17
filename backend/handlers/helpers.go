package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"test-app/backend/types"
)

// writeJSON writes a typed value as a JSON response.
func writeJSON[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a standard error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, types.ErrorResponse{Error: msg})
}

// writeMessage writes a standard success message response.
func writeMessage(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, types.MessageResponse{Message: msg})
}

// decodeJSON reads and decodes a JSON request body into the given type.
func decodeJSON[T any](r *http.Request) (T, error) {
	var v T
	err := json.NewDecoder(r.Body).Decode(&v)
	return v, err
}

// parseIDParam extracts a uint ID from a chi URL parameter.
func parseIDParam(r *http.Request, param string) (uint, error) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// isNotFound checks if the error is a GORM record-not-found error.
func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
