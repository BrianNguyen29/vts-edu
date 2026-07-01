package resources

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// DataEnvelope is the standard success response shape.
type DataEnvelope struct {
	Data any `json:"data"`
}

// ErrorEnvelope is the standard error response shape.
type ErrorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(DataEnvelope{Data: data})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorEnvelope{}
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.RequestID = middleware.GetReqID(r.Context())
	_ = json.NewEncoder(w).Encode(resp)
}
