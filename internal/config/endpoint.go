package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	captureRe    = regexp.MustCompile(`^{([A-Za-z_-]+)}$`)
	queryRe      = regexp.MustCompile(`^GET@([A-Za-z_-]+)$`)
	formRe       = regexp.MustCompile(`^[A-Za-z0-9 \.\(\)\$\+_-]+$`)
	availMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
)

// Endpoint definition
type Endpoint struct {
	Name        string                `json:"name"`
	Method      string                `json:"method"`
	Pattern     string                `json:"path"`
	Scope       [][]string            `json:"scope"`
	Description string                `json:"info"`
	Input       map[string]*Parameter `json:"in"`
	Output      map[string]*Parameter `json:"out"`

	// Captures contains references to URI parameters from the `Input` map.
	// The format for those parameter names is "{paramName}"
	Captures []*BraceCapture `json:"-"`

	// Pattern uri fragments
	fragments []string `json:"-"`
}

// BraceCapture links to the related URI parameter
type BraceCapture struct {
	Name    string
	Index   int
	Defined bool
}

// UnmarshalJSON with custom validation
func (e *Endpoint) UnmarshalJSON(b []byte) error {
	type receiver Endpoint
	var r receiver
	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("%s %s: %w", r.Method, r.Pattern, err)
	}

	e.Name = r.Name
	e.Method = r.Method
	e.Pattern = r.Pattern
	e.Scope = r.Scope
	e.Description = r.Description
	e.Input = r.Input
	e.Output = r.Output

	if err := e.validate(); err != nil {
		return fmt.Errorf("%s %s: %w", r.Method, r.Pattern, err)
	}
	return nil
}

var nameRe = regexp.MustCompile(`^[A-Z][A-Za-z0-9_-]*$`)

// validate the service configuration
func (e *Endpoint) validate() error {
	err := e.checkMethod()
	if err != nil {
		return fmt.Errorf("field 'method': %w", err)
	}

	e.Pattern = strings.Trim(e.Pattern, " \t\r\n")
	err = e.checkPattern()
	if err != nil {
		return fmt.Errorf("field 'path': %w", err)
	}

	// starts with an uppercase letter
	if len(e.Name) < 1 {
		return fmt.Errorf("field 'name': %w", ErrNameMissing)
	}
	if !unicode.IsUpper(rune(e.Name[0])) {
		return fmt.Errorf("field 'name': %w", ErrNameUnexported)
	}
	if !nameRe.MatchString(e.Name) {
		return fmt.Errorf("field 'name': %w", ErrNameInvalid)
	}

	if len(e.Description) < 1 {
		return fmt.Errorf("field 'description': %w", ErrDescMissing)
	}

	err = e.checkInput()
	if err != nil {
		return fmt.Errorf("field 'in': %w", err)
	}

	// fail when a brace capture remains undefined
	for _, capture := range e.Captures {
		if !capture.Defined {
			return fmt.Errorf("field 'in': %s: %w", capture.Name, ErrBraceCaptureUndefined)
		}
	}
	err = e.checkOutput()
	if err != nil {
		return fmt.Errorf("field 'out': %w", err)
	}
	return nil
}

// Match returns if this service would handle this HTTP request
func (e *Endpoint) Match(req *http.Request) bool {
	return req.Method == e.Method && e.matchPattern(req.URL.Path)
}

// checks if an uri matches the service's pattern
func (e *Endpoint) matchPattern(uri string) bool {
	var fragments = URIFragments(uri)
	if len(fragments) != len(e.fragments) {
		return false
	}

	// root url '/'
	if len(e.fragments) == 0 && len(fragments) == 0 {
		return true
	}

	// check part by part
	for i, fragment := range e.fragments {
		part := fragments[i]
		isCapture := len(fragment) > 0 && fragment[0] == '{'

		// if no capture -> check equality
		if !isCapture {
			if fragment != part {
				return false
			}
			continue
		}
		if param, exists := e.Input[fragment]; !exists || param == nil {
			return false
		}
	}

	return true
}

// Validate the endpoint configuration with code-generated validators
func (e *Endpoint) Validate(avail Validators) error {
	for name, param := range e.Input {
		if err := param.Validate(avail); err != nil {
			return fmt.Errorf("field 'in': '%s': %w", name, err)
		}
	}
	return nil
}

func (e *Endpoint) checkMethod() error {
	for _, avail := range availMethods {
		if e.Method == avail {
			return nil
		}
	}
	return ErrMethodUnknown
}

// checkPattern checks for the validity of the pattern definition (i.e. the uri)
//
// Note that the uri can contain capture params e.g. `/a/{b}/c/{d}`, in this
// example, input parameters with names `{b}` and `{d}` are expected.
//
// This methods sets up the service state with adding capture params that are
// expected; checkInputs() will be able to check params against pattern captures
func (e *Endpoint) checkPattern() error {
	length := len(e.Pattern)

	// empty pattern
	if length < 1 {
		return ErrPatternInvalid
	}

	if length > 1 {
		// pattern not starting with '/' or ending with '/'
		if e.Pattern[0] != '/' || e.Pattern[length-1] == '/' {
			return ErrPatternInvalid
		}
	}

	// for each slash-separated chunk
	e.fragments = URIFragments(e.Pattern)
	for i, fragment := range e.fragments {
		if len(fragment) < 1 {
			return ErrPatternInvalid
		}

		// if brace capture
		if matches := captureRe.FindStringSubmatch(fragment); len(matches) > 1 {
			braceName := matches[1]

			// append
			if e.Captures == nil {
				e.Captures = make([]*BraceCapture, 0)
			}
			e.Captures = append(e.Captures, &BraceCapture{
				Index: i,
				Name:  braceName,
			})
			continue
		}

		// fail on invalid format
		if strings.ContainsAny(fragment, "{}") {
			return ErrPatternInvalidBraceCapture
		}
	}

	return nil
}

func (e *Endpoint) checkInput() error {
	// no parameter
	if e.Input == nil || len(e.Input) < 1 {
		e.Input = map[string]*Parameter{}
		return nil
	}

	// for each parameter
	for name, p := range e.Input {
		if len(name) < 1 {
			return fmt.Errorf("%s: %w", name, ErrParamNameIllegal)
		}

		// parse parameters: capture (uri), query or form and update the service
		// attributes accordingly
		err := e.parseParam(name, p)
		if err != nil {
			return err
		}

		// Rename mandatory for capture and query
		if p.Rename == "" && (p.Kind == KindURI || p.Kind == KindQuery) {
			return fmt.Errorf("%s: %w", name, ErrRenameMandatory)
		}

		// fallback to name when Rename is not provided
		if p.Rename == "" {
			p.Rename = name
		}

		// capture parameter cannot be optional
		if p.Optional && p.Kind == KindURI {
			return fmt.Errorf("%s: %w", name, ErrParamOptionalIllegalURI)
		}

		err = nameConflicts(name, p, e.Input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Endpoint) checkOutput() error {
	// no parameter
	if e.Output == nil || len(e.Output) < 1 {
		e.Output = make(map[string]*Parameter, 0)
		return nil
	}

	for name, p := range e.Output {
		if name == "" {
			return fmt.Errorf("%s: %w", name, ErrParamNameIllegal)
		}

		// fallback to name when Rename is not provided
		if p.Rename == "" {
			p.Rename = name
		}

		if p.Optional {
			return fmt.Errorf("%s: %w", name, ErrOutputOptional)
		}

		if err := nameConflicts(name, p, e.Output); err != nil {
			return err
		}
	}
	return nil
}

// parseParam determines which param type it is from its name:
//   - `{paramName}` is an capture; it captures a segment of the uri defined in
//     the pattern definition, e.g. `/some/path/with/{paramName}/somewhere`
//   - `GET@paramName` is an uri query that is received from the http query format
//     in the uri, e.g. `http://domain.com/uri?paramName=paramValue&param2=value2`
//   - any other name that contains valid characters is considered a Form
//     parameter; it is extracted from the http request's body as: json, multipart
//     or using the x-www-form-urlencoded format.
//
// Special notes:
//   - capture params MUST be found in the pattern definition.
//   - capture params MUST NOT be optional as they are in the pattern anyways.
//   - capture and query params MUST be renamed because the `{param}` or
//     `GET@param` name formats cannot be translated to a valid go exported name.
//     c.f. the `dynfunc` package that creates a handler func() signature from
//     the service definitions (i.e. input and output parameters).
func (e *Endpoint) parseParam(name string, p *Parameter) error {
	// Parameter is a capture (uri/{param})
	if match := captureRe.FindStringSubmatch(name); len(match) > 1 {
		p.Kind = KindURI

		// fail if brace capture does not exists in pattern
		var found bool
		for i, capture := range e.Captures {
			if capture.Name == match[1] {
				capture.Defined = true
				p.ExtractName = strconv.FormatInt(int64(i), 10)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: %w", name, ErrBraceCaptureUnspecified)
		}
		return nil
	}

	if match := queryRe.FindStringSubmatch(name); len(match) > 1 {
		p.Kind = KindQuery
		p.ExtractName = match[1]
		return nil
	}

	if match := formRe.MatchString(name); !match {
		return ErrParamNameIllegal
	}
	p.Kind = KindForm
	p.ExtractName = name
	return nil
}

// nameConflicts returns whether ar given parameter has its name or Rename field
// in conflict with an existing parameter
func nameConflicts(name string, param *Parameter, others map[string]*Parameter) error {
	for otherName, other := range others {
		// ignore self
		if otherName == name {
			continue
		}

		// 1. same rename field
		// 2. original name matches a renamed field
		// 3. renamed field matches an original name
		if param.Rename == other.Rename || name == other.Rename || otherName == param.Rename {
			return fmt.Errorf("%s: %w", otherName, ErrParamNameConflict)
		}
	}
	return nil
}