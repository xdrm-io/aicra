package config

import (
	"net/http"

	"git.xdrm.io/go/aicra/config/datatype"
)

var availableHTTPMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

// Server represents a full server configuration
type Server struct {
	Types    []datatype.DataType
	Services []*Service
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

	Captures []*BraceCapture
}

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

// BraceCapture links to the related URI parameter
type BraceCapture struct {
	Name  string
	Index int
	Ref   *Parameter
}

// links to the related URI parameter and hold a value
type braceCaptureValue struct {
	BraceCapture
	Value interface{}
}
