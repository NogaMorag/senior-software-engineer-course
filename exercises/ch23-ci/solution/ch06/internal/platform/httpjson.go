package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// HTTPError is an error that knows which status code and message the client
// should get. Anything else that reaches RespondError is a 500.
type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("%d: %s", e.Status, e.Message) }

// DecodeJSON reads exactly one JSON value of type T from the request body. The
// body is limited to maxBytes, unknown fields are rejected, and so is anything
// after the value. Errors are *HTTPError with status 400 or 413.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			return v, &HTTPError{http.StatusRequestEntityTooLarge, fmt.Sprintf("body is larger than %d bytes", tooBig.Limit)}
		case errors.Is(err, io.EOF):
			return v, &HTTPError{http.StatusBadRequest, "body is empty"}
		default:
			return v, &HTTPError{http.StatusBadRequest, "invalid JSON: " + err.Error()}
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return v, &HTTPError{http.StatusBadRequest, "body must contain a single JSON value"}
	}
	return v, nil
}

// WriteJSON sends v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status is already sent; an encoding error here can only be a
	// client that went away, and there is nothing useful left to tell it.
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError sends {"error": msg} with the given status.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// RespondError sends err to the client: an *HTTPError as it is, anything else
// as a logged 500 whose details stay in the log, not in the response.
func RespondError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	var he *HTTPError
	if errors.As(err, &he) {
		WriteError(w, he.Status, he.Message)
		return
	}
	log.ErrorContext(r.Context(), "request failed", slog.String("error", err.Error()))
	WriteError(w, http.StatusInternalServerError, "internal error")
}
