package config

import "net/http"

var availableHTTPMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

// Service represents a service definition (from api.json)
type Service struct {
	GET    *Method `json:"GET"`
	POST   *Method `json:"POST"`
	PUT    *Method `json:"PUT"`
	DELETE *Method `json:"DELETE"`

	Children map[string]*Service `json:"/"`
}

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool
	Default     *interface{} `json:"default"`
}

// Method represents a method definition (from api.json)
type Method struct {
	Description string                `json:"info"`
	Scope       [][]string            `json:"scope"`
	Parameters  map[string]*Parameter `json:"in"`
	Download    *bool                 `json:"download"`
}
