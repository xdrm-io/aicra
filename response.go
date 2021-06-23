package aicra

import (
	"encoding/json"
	"net/http"

	"github.com/xdrm-io/aicra/api"
)

// response for an service call
type response struct {
	Data   map[string]interface{}
	Status int
	err    api.Err
}

// newResponse creates an empty response.
func newResponse() *response {
	return &response{
		Status: http.StatusOK,
		Data:   make(map[string]interface{}),
		err:    api.ErrFailure,
	}
}

// WithError sets the response error
func (r *response) WithError(err api.Err) *response {
	r.err = err
	return r
}

// WithValue sets a response value
func (r *response) WithValue(name string, value interface{}) *response {
	r.Data[name] = value
	return r
}

// MarshalJSON generates the JSON representation of the response
//
// implements json.Marshaler
func (r *response) MarshalJSON() ([]byte, error) {
	fmt := make(map[string]interface{})
	for k, v := range r.Data {
		fmt[k] = v
	}
	fmt["error"] = r.err
	return json.Marshal(fmt)
}

// ServeHTTP writes the response representation back to the http.ResponseWriter
//
// implements http.Handler
func (res *response) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(res.err.Status)
	encoded, err := json.Marshal(res)
	if err == nil {
		w.Write(encoded)
	}
	return err
}
