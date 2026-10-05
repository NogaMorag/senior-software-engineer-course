package platform

import (
	"encoding/json"
	"net/http"
)

// WriteJSON sends v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status is already sent; an encoding error here can only be a
	// client that went away, and there is nothing useful left to tell it.
	_ = json.NewEncoder(w).Encode(v)
}
