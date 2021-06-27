package aicra

import (
	"encoding/json"
	"net/http"

	"github.com/xdrm-io/aicra/api"
)

// Responder defines how to write data and error  into the http response
type Responder func(http.ResponseWriter, map[string]interface{}, api.Err)

// DefaultResponder used for writing data and error into http responses
func DefaultResponder(w http.ResponseWriter, data map[string]interface{}, apiErr api.Err) {
	w.WriteHeader(apiErr.Status)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	if data == nil {
		data = map[string]interface{}{}
	}

	data["error"] = apiErr

	encoded, err := json.Marshal(data)
	if err == nil {
		w.Write(encoded)
	}
}
