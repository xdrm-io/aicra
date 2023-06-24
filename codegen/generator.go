package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"text/template"

	"github.com/xdrm-io/aicra/internal/config"
)

//go:embed endpoints.tmpl
var endpointsTmpl embed.FS

// Generator can generate go code from the configuration
type Generator struct {
	Config config.API
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

// validates go codes using go parser and printer
func validate(src []byte, w io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return fmt.Errorf("go parser: %w", err)
	}
	cfg := printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 4,
	}
	return cfg.Fprint(w, fset, f)
}
