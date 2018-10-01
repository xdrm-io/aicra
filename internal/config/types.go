package config

import (
	"git.xdrm.io/go/aicra/driver"
)

type builder struct {
	// Default tells whether or not to ignore the built-in components
	Default bool `json:"default,ommitempty"`

	// Folder is used to infer the 'Map' object
	Folder string `json:"folder,ommitempty"`

	// Map defines the association path=>file
	Map map[string]string
}

// Schema represents an AICRA configuration (not the API, the server, drivers, etc)
type Schema struct {
	// Root is root of the project structure default is "." (current directory)
	Root string `json:"root,ommitempty"`

	// Host is the hostname to listen to (default is 0.0.0.0)
	Host string `json:"host,ommitempty"`
	// Port is the port to listen to (default is 80)
	Port uint16 `json:"port,ommitempty"`

	// DriverName is the driver used to load the controllers and middlewares
	DriverName string `json:"driver"`
	Driver     driver.Driver

	// Types defines :
	// - the type folder
	// - whether to load the built-in types
	//
	// types are ommited if not set (no default)
	Types *builder `json:"types,ommitempty"`

	// Controllers defines :
	// - the controller folder
	//
	// (default is .build/controller)
	Controllers *builder `json:"controllers,ommitempty"`

	// Middlewares defines :
	// - the middleware folder
	//
	// (default is .build/middleware)
	Middlewares *builder `json:"middlewares,ommitempty"`
}
