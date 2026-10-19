// Package httpapi is the HTTP layer: router, handlers, middleware and the
// JSON/error helpers (specs/backend/plan.md §4).
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Error codes (contract §1.2).
const (
	CodeNotFound         = "not_found"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInternal         = "internal_error"
	CodeUnavailable      = "unavailable"
)

const contentTypeJSON = "application/json; charset=utf-8"

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeJSON writes v as a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		slog.Error("encode response", "err", err)
		writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error")
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// writeError is the only way errors leave the service. msg must never
// contain internal details.
func writeError(w http.ResponseWriter, status int, code, msg string) {
	var b errorBody
	b.Error.Code, b.Error.Message = code, msg
	body, _ := json.Marshal(b) // cannot fail: two strings
	h := w.Header()
	h.Set("Content-Type", contentTypeJSON)
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
