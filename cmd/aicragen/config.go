package main

import (
	"encoding/json"
	"io"
)

// Config contains the aicra configuration
type Config struct {
	Package    string            `json:"package"`
	Imports    map[string]string `json:"imports"`
	Validators []string          `json:"validators"`
	Endpoints  []*Endpoint       `json:"endpoints"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Name        string                `json:"name"`
	Method      string                `json:"method"`
	Pattern     string                `json:"path"`
	Scope       [][]string            `json:"scope"`
	Description string                `json:"info"`
	Input       map[string]*Parameter `json:"in"`
	Output      map[string]*Parameter `json:"out"`
}

// Parameter represents an endpoint parameter definition
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Optional    bool   `json:"-"`
}

// Decode from json
func (c *Config) Decode(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	for _, endpoint := range c.Endpoints {
		for id, param := range endpoint.Input {
			if len(param.Type) > 0 && param.Type[0] == '?' {
				endpoint.Input[id].Optional = true
				endpoint.Input[id].Type = param.Type[1:]
			}
		}
	}
	return nil
}
