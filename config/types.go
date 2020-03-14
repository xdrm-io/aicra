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
	Optional    bool

	// validator is set from the @Type
	validator datatype.Validator
}

// Service represents a service definition (from api.json)
type Service struct {
	Method      string                `json:"method"`
	Pattern     string                `json:"path"`
	Scope       [][]string            `json:"scope"`
	Description string                `json:"info"`
	Download    *bool                 `json:"download"`
	Input       map[string]*Parameter `json:"in"`
	// Output map[string]*Parameter `json:"out"`
}

// Services contains every service that represents a server configuration
type Services []*Service
