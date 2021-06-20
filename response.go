package aicra

import (
	"encoding/json"
	"net/http"

	"github.com/xdrm-io/aicra/api"
)

// response for an service call
type response struct {
	Data    map[string]interface{}
	Status  int
	Headers http.Header
	err     api.Err
}

// newResponse creates an empty response.
func newResponse() *response {
	return &response{
		Status:  http.StatusOK,
		Data:    make(map[string]interface{}),
		err:     api.ErrFailure,
		Headers: make(http.Header),
	}
}

// WithError sets the response error
func (res *response) WithError(err api.Err) *response {
	res.err = err
	return res
}

// SetValue sets a response value
func (res *response) SetValue(name string, value interface{}) {
	res.Data[name] = value
}

// MarshalJSON generates the JSON representation of the response
//
// implements json.Marshaler
func (res *response) MarshalJSON() ([]byte, error) {
	fmt := make(map[string]interface{})
	for k, v := range res.Data {
		fmt[k] = v
	}
	fmt["error"] = res.err
	return json.Marshal(fmt)
}

// ServeHTTP writes the response representation back to the http.ResponseWriter
//
// implements http.Handler
func (res *response) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(res.err.Status)
	encoded, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Write(encoded)
	return nil
}
