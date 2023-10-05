package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"text/template"

	"github.com/xdrm-io/aicra/internal/config"
)

var (
	//go:embed endpoints.tmpl
	endpointsTmpl embed.FS
	//go:embed validators.tmpl
	validatorsTmpl embed.FS
	//go:embed mappers.tmpl
	mappersTmpl embed.FS
)

// Generator can generate go code from the configuration
type Generator struct {
	Config        config.API
	ConfigRelPath string
}

// WriteEndpoints writes endpoints go file
func (g Generator) WriteEndpoints(w io.Writer) error {
	tmpl := template.New("endpoints.tmpl")
	tmpl.Funcs(template.FuncMap{
		"getType":          getType,
		"getConfigRelPath": getConfigRelPath(g.ConfigRelPath),
	})
	tmpl, err := tmpl.ParseFS(endpointsTmpl, "endpoints.tmpl")
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, g.Config); err != nil {
		return err
	}
	return validate(buf.Bytes(), w)
}

// WriteValidators writes validators go file
func (g Generator) WriteValidators(w io.Writer) error {
	tmpl := template.New("validators.tmpl")
	tmpl.Funcs(template.FuncMap{
		"getterName": getterName,
	})
	tmpl, err := tmpl.ParseFS(validatorsTmpl, "validators.tmpl")
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, g.Config); err != nil {
		return err
	}
	return validate(buf.Bytes(), w)
}

// WriteMappers writes mappers go file
func (g Generator) WriteMappers(w io.Writer) error {
	tmpl := template.New("mappers.tmpl")
	tmpl.Funcs(template.FuncMap{
		"getType":   getType,
		"getGetter": getGetter,
		"params":    params,
	})
	tmpl, err := tmpl.ParseFS(mappersTmpl, "mappers.tmpl")
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

// getConfigRelPath returns the path of the config file relative to the generated
// endpoints file
func getConfigRelPath(path string) func() string {
	return func() string {
		return path
	}
}

// getType returns the GO type associated to a parameter according to the
// validators
func getType(validatorName string, validators map[string]config.Validator) string {
	validator, ok := validators[validatorName]
	if !ok {
		return ""
	}
	return validator.Type
}

// getGetter returns the getter name for the associated validator for a parameter
func getGetter(validatorName string, validators map[string]config.Validator) string {
	validator, ok := validators[validatorName]
	if !ok {
		return ""
	}
	return getterName(validator.Validator)
}

// validator getter name `validator<Package><Symbol>`
func getterName(symbol string) string {
	const (
		prefix = "get"
		suffix = "Validator"
	)
	parts := strings.Split(symbol, ".")
	if len(parts) == 0 {
		return ""
	}

	var capitalize = func(s string) string {
		if len(s) == 0 {
			return ""
		}
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		return strings.ToUpper(s[0:1]) + s[1:]
	}

	if len(parts) == 1 {
		return prefix + capitalize(parts[0]) + suffix
	}
	return prefix + capitalize(parts[0]) + capitalize(parts[1]) + suffix
}

// format string parameters
func params(s []string) string {
	if len(s) == 0 || len(s) == 1 && s[0] == "" {
		return "nil"
	}
	if len(s) == 1 {
		return `[]string{"` + s[0] + `"}`
	}
	return `[]string{"` + strings.Join(s, `", "`) + `"}`
}
