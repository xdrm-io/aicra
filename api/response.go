package api

import (
	"encoding/json"
	"net/http"
)

// ResponseData defines format for response parameters to return
type ResponseData map[string]interface{}

// Response represents an API response to be sent
type Response struct {
	Data    ResponseData
	Status  int
	Headers http.Header
	err     Error
}

// EmptyResponse creates an empty response.
func EmptyResponse() *Response {
	return &Response{
		Status:  http.StatusOK,
		Data:    make(ResponseData),
		err:     ErrorFailure,
		Headers: make(http.Header),
	}
}

// WithError sets the error
func (res *Response) WithError(err Error) *Response {
	res.err = err
	return res
}

func (res *Response) Error() string {
	return res.err.Error()
}

// SetData adds/overrides a new response field
func (res *Response) SetData(name string, value interface{}) {
	res.Data[name] = value
}

// MarshalJSON implements the 'json.Marshaler' interface and is used
// to generate the JSON representation of the response
func (res *Response) MarshalJSON() ([]byte, error) {
	fmt := make(map[string]interface{})
	for k, v := range res.Data {
		fmt[k] = v
	}
	fmt["error"] = res.err
	return json.Marshal(fmt)
}

func (res *Response) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(res.err.Status())
	encoded, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Write(encoded)
	return nil
}
