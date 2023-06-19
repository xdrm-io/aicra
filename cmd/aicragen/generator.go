package main

import (
	"embed"
	"io"
	"text/template"
)

var (
	//go:embed mappers.tmpl
	mappersTmpl embed.FS
	//go:embed endpoints.tmpl
	endpointsTmpl embed.FS
	//go:embed validators.tmpl
	validatorsTmpl embed.FS
)

// Generator can generate go code from the configuration
type Generator struct {
	Config
}

// WriteValidators writes validators go file
func (g Generator) WriteValidators(w io.Writer) error {
	tmpl, err := template.ParseFS(validatorsTmpl, "validators.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, g.Config)
}

// WriteMappers writes mappers go file
func (g Generator) WriteMappers(w io.Writer) error {
	tmpl, err := template.ParseFS(mappersTmpl, "mappers.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, g.Config)
}

// WriteEndpoints writes endpoints go file
func (g Generator) WriteEndpoints(w io.Writer) error {
	tmpl, err := template.ParseFS(endpointsTmpl, "endpoints.tmpl")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, g.Config)
}
