package config

import (
	"net/http"

	"git.xdrm.io/go/aicra/config/datatype"
)

var availableHTTPMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	// Optional is set to true when the type is prefixed with '?'
	Optional bool

	// Validator is inferred from @Type
	Validator datatype.Validator
}

// Service represents a service definition (from api.json)
type Service struct {
	Method      string                `json:"method"`
	Pattern     string                `json:"path"`
	Scope       [][]string            `json:"scope"`
	Description string                `json:"info"`
	Input       map[string]*Parameter `json:"in"`
	// Download    *bool                 `json:"download"`
	// Output map[string]*Parameter `json:"out"`
}

// Server represents a full server configuration
type Server struct {
	types    []datatype.DataType
	services []*Service
}
