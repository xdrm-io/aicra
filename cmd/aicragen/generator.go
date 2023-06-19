package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"text/template"
)

var (
	//go:embed endpoints.tmpl
	endpointsTmpl embed.FS
)

// Generator can generate go code from the configuration
type Generator struct {
	Config
}

// WriteEndpoints writes endpoints go file
func (g Generator) WriteEndpoints(w io.Writer) error {
	tmpl, err := template.ParseFS(endpointsTmpl, "endpoints.tmpl")
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, g.Config); err != nil {
		return err
	}
	return validate(buf.Bytes(), w)
}

// validates go codes using the language parser
func validate(src []byte, w io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return fmt.Errorf("go parser: %w", err)
	}
	return printer.Fprint(w, fset, f)
}
