package config

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/xdrm-io/aicra/validator"
)

var (
	captureRegex         = regexp.MustCompile(`^{([a-z_-]+)}$`)
	queryRegex           = regexp.MustCompile(`^GET@([a-z_-]+)$`)
	availableHTTPMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
)

// Service definition
type Service struct {
	Method      string                `json:"method"`
	Pattern     string                `json:"path"`
	Scope       [][]string            `json:"scope"`
	Description string                `json:"info"`
	Input       map[string]*Parameter `json:"in"`
	Output      map[string]*Parameter `json:"out"`

	// Captures contains references to URI parameters from the `Input` map.
	// The format for those parameter names is "{paramName}"
	Captures []*BraceCapture

	// Query contains references to HTTP Query parameters from the `Input` map.
	// Query parameters names are "GET@paramName", this map contains escaped
	// names, e.g. "paramName"
	Query map[string]*Parameter

	// Form references form parameters from the `Input` map (all but Captures
	// and Query).
	Form map[string]*Parameter
}

// BraceCapture links to the related URI parameter
type BraceCapture struct {
	Name  string
	Index int
	Ref   *Parameter
}

// Match returns if this service would handle this HTTP request
func (svc *Service) Match(req *http.Request) bool {
	var (
		uri        = req.RequestURI
		queryIndex = strings.IndexByte(uri, '?')
	)

	// remove query part for matching the pattern
	if queryIndex > -1 {
		uri = uri[:queryIndex]
	}

	return req.Method == svc.Method && svc.matchPattern(uri)
}

// checks if an uri matches the service's pattern
func (svc *Service) matchPattern(uri string) bool {
	var (
		uriparts = SplitURL(uri)
		parts    = SplitURL(svc.Pattern)
	)

	if len(uriparts) != len(parts) {
		return false
	}

	// root url '/'
	if len(parts) == 0 && len(uriparts) == 0 {
		return true
	}

	// check part by part
	for i, part := range parts {
		uripart := uriparts[i]

		isCapture := len(part) > 0 && part[0] == '{'

		// if no capture -> check equality
		if !isCapture {
			if part != uripart {
				return false
			}
			continue
		}

		param, exists := svc.Input[part]

		// fail if no validator
		if !exists || param.Validator == nil {
			return false
		}

		// fail if not type-valid
		if _, valid := param.Validator(uripart); !valid {
			return false
		}
	}

	return true
}

// validate the service configuration
func (svc *Service) validate(input []validator.Type, output []validator.Type) error {
	err := svc.checkMethod()
	if err != nil {
		return fmt.Errorf("field 'method': %w", err)
	}

	svc.Pattern = strings.Trim(svc.Pattern, " \t\r\n")
	err = svc.checkPattern()
	if err != nil {
		return fmt.Errorf("field 'path': %w", err)
	}

	if len(strings.Trim(svc.Description, " \t\r\n")) < 1 {
		return fmt.Errorf("field 'description': %w", ErrMissingDescription)
	}

	err = svc.checkInput(input)
	if err != nil {
		return fmt.Errorf("field 'in': %w", err)
	}

	// fail when a brace capture remains undefined
	for _, capture := range svc.Captures {
		if capture.Ref == nil {
			return fmt.Errorf("field 'in': %s: %w", capture.Name, ErrUndefinedBraceCapture)
		}
	}

	err = svc.checkOutput(output)
	if err != nil {
		return fmt.Errorf("field 'out': %w", err)
	}

	return nil
}

func (svc *Service) checkMethod() error {
	for _, available := range availableHTTPMethods {
		if svc.Method == available {
			return nil
		}
	}
	return ErrUnknownMethod
}

// checkPattern checks for the validity of the pattern definition (i.e. the uri)
//
// Note that the uri can contain capture params e.g. `/a/{b}/c/{d}`, in this
// example, input parameters with names `{b}` and `{d}` are expected.
//
// This methods sets up the service state with adding capture params that are
// expected; checkInputs() will be able to check params agains pattern captures.
func (svc *Service) checkPattern() error {
	length := len(svc.Pattern)

	// empty pattern
	if length < 1 {
		return ErrInvalidPattern
	}

	if length > 1 {
		// pattern not starting with '/' or ending with '/'
		if svc.Pattern[0] != '/' || svc.Pattern[length-1] == '/' {
			return ErrInvalidPattern
		}
	}

	// for each slash-separated chunk
	parts := SplitURL(svc.Pattern)
	for i, part := range parts {
		if len(part) < 1 {
			return ErrInvalidPattern
		}

		// if brace capture
		if matches := captureRegex.FindAllStringSubmatch(part, -1); len(matches) > 0 && len(matches[0]) > 1 {
			braceName := matches[0][1]

			// append
			if svc.Captures == nil {
				svc.Captures = make([]*BraceCapture, 0)
			}
			svc.Captures = append(svc.Captures, &BraceCapture{
				Index: i,
				Name:  braceName,
				Ref:   nil,
			})
			continue
		}

		// fail on invalid format
		if strings.ContainsAny(part, "{}") {
			return ErrInvalidPatternBraceCapture
		}
	}

	return nil
}

func (svc *Service) checkInput(validators []validator.Type) error {
	// no parameter
	if svc.Input == nil || len(svc.Input) < 1 {
		svc.Input = map[string]*Parameter{}
		return nil
	}

	// for each parameter
	for name, p := range svc.Input {
		if len(name) < 1 {
			return fmt.Errorf("%s: %w", name, ErrIllegalParamName)
		}

		// parse parameters: capture (uri), query or form and update the service
		// attributes accordingly
		ptype, err := svc.parseParam(name, p)
		if err != nil {
			return err
		}

		// Rename mandatory for capture and query
		if len(p.Rename) < 1 && (ptype == captureParam || ptype == queryParam) {
			return fmt.Errorf("%s: %w", name, ErrMandatoryRename)
		}

		// fallback to name when Rename is not provided
		if len(p.Rename) < 1 {
			p.Rename = name
		}

		err = p.validate(validators...)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		// capture parameter cannot be optional
		if p.Optional && ptype == captureParam {
			return fmt.Errorf("%s: %w", name, ErrIllegalOptionalURIParam)
		}

		err = nameConflicts(name, p, svc.Input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) checkOutput(validators []validator.Type) error {
	// no parameter
	if svc.Output == nil || len(svc.Output) < 1 {
		svc.Output = make(map[string]*Parameter, 0)
		return nil
	}

	for name, p := range svc.Output {
		if len(name) < 1 {
			return fmt.Errorf("%s: %w", name, ErrIllegalParamName)
		}

		// fallback to name when Rename is not provided
		if len(p.Rename) < 1 {
			p.Rename = name
		}

		err := p.validate(validators...)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		if p.Optional {
			return fmt.Errorf("%s: %w", name, ErrOptionalOption)
		}

		err = nameConflicts(name, p, svc.Output)
		if err != nil {
			return err
		}
	}
	return nil
}

type paramType int

const (
	captureParam paramType = iota
	queryParam
	formParam
)

// parseParam determines which param type it is from its name:
// - `{paramName}` is an capture; it captures a segment of the uri defined in
//    the pattern definition, e.g. `/some/path/with/{paramName}/somewhere`
// - `GET@paramName` is an uri query that is received from the http query format
//    in the uri, e.g. `http://domain.com/uri?paramName=paramValue&param2=value2`
// - any other name that contains valid characters is considered a Form
//   parameter; it is extracted from the http request's body as: json, multipart
//   or using the x-www-form-urlencoded format.
//
// Special notes:
// - capture params MUST be found in the pattern definition.
// - capture params MUST NOT be optional as they are in the pattern anyways.
// - capture and query params MUST be renamed because the `{param}` or
//   `GET@param` name formats cannot be translated to a valid go exported name.
//    c.f. the `dynfunc` package that creates a handler func() signature from
//    the service definitions (i.e. input and output parameters).
func (svc *Service) parseParam(name string, p *Parameter) (paramType, error) {
	var (
		captureMatches = captureRegex.FindAllStringSubmatch(name, -1)
		isCapture      = len(captureMatches) > 0 && len(captureMatches[0]) > 1
	)

	// Parameter is a capture (uri/{param})
	if isCapture {
		captureName := captureMatches[0][1]

		// fail if brace capture does not exists in pattern
		found := false
		for _, capture := range svc.Captures {
			if capture.Name == captureName {
				capture.Ref = p
				found = true
				break
			}
		}
		if !found {
			return captureParam, fmt.Errorf("%s: %w", name, ErrUnspecifiedBraceCapture)
		}
		return captureParam, nil
	}

	var (
		queryMatches = queryRegex.FindAllStringSubmatch(name, -1)
		isQuery      = len(queryMatches) > 0 && len(queryMatches[0]) > 1
	)

	// Parameter is a query (uri?param)
	if isQuery {
		queryName := queryMatches[0][1]

		// init map
		if svc.Query == nil {
			svc.Query = make(map[string]*Parameter)
		}
		svc.Query[queryName] = p

		return queryParam, nil
	}

	// Parameter is a form param
	if svc.Form == nil {
		svc.Form = make(map[string]*Parameter)
	}
	svc.Form[name] = p
	return formParam, nil
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
