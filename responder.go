package aicra

import (
	"encoding/json"
	"net/http"

	"github.com/xdrm-io/aicra/api"
)

// Responder defines how to write data and error  into the http response
type Responder func(http.ResponseWriter, map[string]interface{}, error)

// DefaultResponder used for writing data and error into http responses
func DefaultResponder(w http.ResponseWriter, data map[string]interface{}, e error) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(api.GetErrorStatus(e))

	if data == nil {
		data = make(map[string]interface{}, 1)
	}

	data["status"] = "all right"
	if e != nil {
		data["status"] = e.Error()
	}

	encoded, err := json.Marshal(data)
	if err == nil {
		w.Write(encoded)
	}
}
