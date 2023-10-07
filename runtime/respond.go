package runtime

import (
	"encoding/json"
	"net/http"

	"github.com/xdrm-io/aicra/api"
)

// Respond using the user provided responder
func Respond(w http.ResponseWriter, data map[string]any, err error) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(api.GetErrorStatus(err))

	if data == nil {
		if err == nil {
			w.Write([]byte(`{"status":"all right"}`))
			return
		}
		w.Write([]byte(`{"status":"` + err.Error() + `"}`))
		return
	}

	data["status"] = "all right"
	if err != nil {
		data["status"] = err.Error()
	}

	encoded, err := json.Marshal(data)
	if err == nil {
		w.Write(encoded)
	}
}
