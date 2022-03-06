package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestAddInputType(t *testing.T) {
	t.Parallel()

	srv := &Server{}

	srv.AddInputValidator(validator.BoolType{})
	if srv.Input == nil {
		t.Fatalf("input is nil")
	}
	if len(srv.Input) != 1 {
		t.Fatalf("expected 1 input validator; got %d", len(srv.Input))
	}
	srv.AddInputValidator(validator.IntType{})
	if len(srv.Input) != 2 {
		t.Fatalf("expected 2 input validator; got %d", len(srv.Input))
	}
}
func TestAddOutputType(t *testing.T) {
	t.Parallel()

	srv := &Server{}

	srv.AddOutputValidator("bool", reflect.TypeOf(true))
	if srv.Output == nil {
		t.Fatalf("input is nil")
	}
	if len(srv.Output) != 1 {
		t.Fatalf("expected 1 input validator; got %d", len(srv.Output))
	}
	srv.AddOutputValidator("string", reflect.TypeOf(""))
	if len(srv.Output) != 2 {
		t.Fatalf("expected 2 input validator; got %d", len(srv.Output))
	}
}

func TestLegalServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Raw   string
		Error error
	}{
		// empty
		{
			`[ { "method": "GET", "info": "a", "path": "" } ]`,
			ErrInvalidPattern,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "no-starting-slash" } ]`,
			ErrInvalidPattern,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "ending-slash/" } ]`,
			ErrInvalidPattern,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/" } ]`,
			nil,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/valid-name" } ]`,
			nil,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/valid/nested/name" } ]`,
			nil,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}" } ]`,
			ErrInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}a" } ]`,
			ErrInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}" } ]`,
			ErrUndefinedBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}/abc" } ]`,
			ErrInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}s/abc" } ]`,
			ErrInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/abc" } ]`,
			ErrUndefinedBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{b{races}s/abc" } ]`,
			ErrInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/}abc" } ]`,
			ErrInvalidPatternBraceCapture,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("service.%d", i), func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(test.Raw))

			if err == nil && test.Error != nil {
				t.Fatalf("expected an error: %q", test.Error.Error())
			}
			if err != nil && test.Error == nil {
				t.Fatalf("unexpected error: %q", err.Error())
			}

			if err != nil && test.Error != nil {
				if !errors.Is(err, test.Error) {
					t.Fatalf("expected the error %q (got %q)", test.Error.Error(), err.Error())
				}
			}
		})
	}
}
func TestAvailableMethods(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Raw         string
		ValidMethod bool
	}{
		{ // missing description
			`[ { "method": "GET", "path": "/", "info": "valid-description" }]`,
			true,
		},
		{ // missing description
			`[ { "method": "POST", "path": "/", "info": "valid-description" }]`,
			true,
		},
		{ // empty description
			`[ { "method": "PUT", "path": "/", "info": "valid-description" }]`,
			true,
		},
		{ // empty trimmed description
			`[ { "method": "DELETE", "path": "/", "info": "valid-description" }]`,
			true,
		},
		{ // valid description
			`[ { "method": "get", "path": "/", "info": "valid-description" }]`,
			false,
		},
		{ // valid description
			`[ { "method": "UNknOwN", "path": "/", "info": "valid-description" }]`,
			false,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("service.%d", i), func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(test.Raw))

			if test.ValidMethod && err != nil {
				t.Fatalf("unexpected error: %q", err.Error())
			}

			if !test.ValidMethod && !errors.Is(err, ErrUnknownMethod) {
				t.Fatalf("expected error <%s> got <%s>", ErrUnknownMethod, err)
			}
		})
	}
}
func TestParseEmpty(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(`[]`)
	srv := &Server{}
	err := srv.Parse(r)
	if err != nil {
		t.Fatalf("unexpected error (got %q)", err)
	}
}
func TestParseJsonError(t *testing.T) {
	r := strings.NewReader(`{
		"GET": {
			"info": "info
		},
	}`) // trailing ',' is invalid JSON
	srv := &Server{}
	err := srv.Parse(r)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseMissingMethodDescription(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Raw              string
		ValidDescription bool
	}{
		{ // missing description
			`[ { "method": "GET", "path": "/" }]`,
			false,
		},
		{ // missing descriptiontype
			`[ { "method": "GET", "path": "/subservice" }]`,
			false,
		},
		{ // empty description
			`[ { "method": "GET", "path": "/", "info": "" }]`,
			false,
		},
		{ // empty trimmed description
			`[ { "method": "GET", "path": "/", "info": " " }]`,
			false,
		},
		{ // valid description
			`[ { "method": "GET", "path": "/", "info": "a" }]`,
			true,
		},
		{ // valid description
			`[ { "method": "GET", "path": "/", "info": "some description" }]`,
			true,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(test.Raw))

			if test.ValidDescription && err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			if !test.ValidDescription && !errors.Is(err, ErrMissingDescription) {
				t.Fatalf("expected error <%s> got <%s>", ErrMissingDescription, err)
			}
		})
	}

}

func TestParamEmptyRenameNoRename(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(`[
		{
			"method": "GET",
			"path": "/",
			"info": "valid-description",
			"in": {
				"original": { "info": "valid-desc", "type": "any", "name": "" }
			}
		}
	]`)
	srv := &Server{}
	srv.Input = append(srv.Input, validator.AnyType{})
	err := srv.Parse(r)
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}

	if len(srv.Services) < 1 {
		t.Fatalf("expected a service")
	}

	for _, param := range srv.Services[0].Input {
		if param.Rename != "original" {
			t.Fatalf("expected the parameter 'original' not to be renamed to %q", param.Rename)
		}
	}

}

func TestMissingParam(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name   string
		in     []validator.Type
		out    map[string]interface{}
		config string
		match  string // match found in the error string
	}{
		{
			name: "no input",
			in:   []validator.Type{},
			out:  map[string]interface{}{},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"in1": { "info": "info", "type": "bool" }
				},
				"out": {}
			}]`,
			match: "field 'in': in1:",
		},
		{
			name: "no input optional",
			in:   []validator.Type{},
			out:  map[string]interface{}{},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"in1": { "info": "info", "type": "?bool" }
				},
				"out": {}
			}]`,
			match: "field 'in': in1:",
		},
		{
			name: "out as in",
			in:   []validator.Type{},
			out: map[string]interface{}{
				"bool": true,
			},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"in1": { "info": "info", "type": "bool" }
				},
				"out": {}
			}]`,
			match: "field 'in': in1:",
		},
		{
			name: "out as in optional",
			in:   []validator.Type{},
			out: map[string]interface{}{
				"bool": true,
			},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"in1": { "info": "info", "type": "?bool" }
				},
				"out": {}
			}]`,
			match: "field 'in': in1:",
		},
		{
			name: "no output",
			in:   []validator.Type{},
			out:  map[string]interface{}{},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {},
				"out": {
					"out1": { "info": "info", "type": "bool" }
				}
			}]`,
			match: "field 'out': out1:",
		},
		{
			name: "in as out",
			in:   []validator.Type{validator.BoolType{}},
			out:  map[string]interface{}{},
			config: `[{
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {},
				"out": {
					"out1": { "info": "info", "type": "bool" }
				}
			}]`,
			match: "field 'out': out1:",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			for _, t := range tc.in {
				srv.AddInputValidator(t)
			}
			for k, v := range tc.out {
				srv.AddOutputValidator(k, reflect.TypeOf(v))
			}

			err := srv.Parse(strings.NewReader(tc.config))
			if !errors.Is(err, ErrUnknownParamType) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, ErrUnknownParamType)
			}
			if err == nil {
				return
			}
			if !strings.Contains(err.Error(), tc.match) {
				t.Fatalf("error %q does not contain %q", err, tc.match)
			}
		})
	}
}

func TestOptionalParam(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(`[
		{
			"method": "GET",
			"path": "/",
			"info": "valid-description",
			"in": {
				"optional": { "info": "optional-type", "type": "?bool" },
				"required": { "info": "required-type", "type": "bool" },
				"required2": { "info": "required", "type": "any" },
				"optional2": { "info": "optional", "type": "?any" }
			}
		}
	]`)
	srv := &Server{}
	srv.Input = append(srv.Input, validator.AnyType{})
	srv.Input = append(srv.Input, validator.BoolType{})
	err := srv.Parse(r)
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}

	if len(srv.Services) < 1 {
		t.Fatalf("expected a service")
	}
	for pName, param := range srv.Services[0].Input {

		if pName == "optional" || pName == "optional2" {
			if !param.Optional {
				t.Errorf("expected parameter %q to be optional", pName)
				t.Failed()
			}
		}
		if pName == "required" || pName == "required2" {
			if param.Optional {
				t.Errorf("expected parameter %q to be required", pName)
				t.Failed()
			}
		}
	}

}
func TestParseParameters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Raw   string
		Error error
	}{
		{ // invalid param name prefix
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"_param1": { }
					}
				}
			]`,
			ErrMissingParamDesc,
		},
		{ // invalid param name suffix
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1_": { }
					}
				}
			]`,
			ErrMissingParamDesc,
		},

		{ // missing param description
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { }
					}
				}
			]`,
			ErrMissingParamDesc,
		},
		{ // empty param description
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "" }
					}
				}
			]`,
			ErrMissingParamDesc,
		},

		{ // missing param type
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid" }
					}
				}
			]`,
			ErrMissingParamType,
		},
		{ // empty param type
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "" }
					}
				}
			]`,
			ErrMissingParamType,
		},
		{ // invalid type (optional mark only)
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "?" }
					}
				}
			]`,

			ErrMissingParamType,
		},
		{ // valid description + valid type
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "any" }
					}
				}
			]`,
			nil,
		},
		{ // valid description + valid OPTIONAL type
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "?any" }
					}
				}
			]`,
			nil,
		},

		{ // name conflict with rename
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "any" },
						"param2": { "info": "valid", "type": "any", "name": "param1" }
					}
				}
			]`,
			// 2 possible errors as map order is not deterministic
			ErrParamNameConflict,
		},
		{ // rename conflict with name
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "any", "name": "param2" },
						"param2": { "info": "valid", "type": "any" }
					}
				}
			]`,
			// 2 possible errors as map order is not deterministic
			ErrParamNameConflict,
		},
		{ // rename conflict with rename
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "any", "name": "conflict" },
						"param2": { "info": "valid", "type": "any", "name": "conflict" }
					}
				}
			]`,
			// 2 possible errors as map order is not deterministic
			ErrParamNameConflict,
		},

		{ // both renamed with no conflict
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "any", "name": "freename" },
						"param2": { "info": "valid", "type": "any", "name": "freename2" }
					}
				}
			]`,
			nil,
		},
		// missing rename
		{
			`[
				{
					"method": "GET",
					"path": "/{uri}",
					"info": "info",
					"in": {
						"{uri}": { "info": "valid", "type": "any" }
					}
				}
			]`,
			ErrMandatoryRename,
		},
		{
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"GET@abc": { "info": "valid", "type": "any" }
					}
				}
			]`,
			ErrMandatoryRename,
		},
		{
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"GET@abc": { "info": "valid", "type": "any", "name": "abc" }
					}
				}
			]`,
			nil,
		},

		{ // URI parameter
			`[
				{
					"method": "GET",
					"path": "/{uri}",
					"info": "info",
					"in": {
						"{uri}": { "info": "valid", "type": "any", "name": "freename" }
					}
				}
			]`,
			nil,
		},
		{ // URI parameter cannot be optional
			`[
				{
					"method": "GET",
					"path": "/{uri}",
					"info": "info",
					"in": {
						"{uri}": { "info": "valid", "type": "?any", "name": "freename" }
					}
				}
			]`,
			ErrIllegalOptionalURIParam,
		},
		{ // URI parameter not specified
			`[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"{uri}": { "info": "valid", "type": "?any", "name": "freename" }
					}
				}
			]`,
			ErrUnspecifiedBraceCapture,
		},
		{ // URI parameter not defined
			`[
				{
					"method": "GET",
					"path": "/{uri}",
					"info": "info",
					"in": { }
				}
			]`,
			ErrUndefinedBraceCapture,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			srv.Input = append(srv.Input, validator.AnyType{})
			err := srv.Parse(strings.NewReader(test.Raw))

			if err == nil && test.Error != nil {
				t.Fatalf("expected an error: %q", test.Error.Error())
			}
			if err != nil && test.Error == nil {
				t.Fatalf("unexpected error: %q", err.Error())
			}

			if err != nil && test.Error != nil {
				if !errors.Is(err, test.Error) {
					t.Fatalf("expected the error <%s> got <%s>", test.Error, err)
				}
			}
		})
	}

}

func serializeParams(t *testing.T, params map[string]string) []byte {
	var parameters = map[string]*Parameter{}
	for k, t := range params {
		parameters[k] = &Parameter{
			Description: "info",
			Type:        t,
			GoType:      validator.StringType{}.GoType(),
			Validator:   validator.StringType{}.Validator(t),
			Rename:      strings.TrimSuffix(strings.TrimPrefix(k, "{"), "}"),
		}
		if t == "bool" {
			parameters[k].GoType = validator.BoolType{}.GoType()
			parameters[k].Validator = validator.BoolType{}.Validator(t)
		}
	}

	raw, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("cannot serialize: %s", err)
	}
	return raw
}

func TestServiceCollision(t *testing.T) {
	t.Parallel()

	type service struct {
		method string
		path   string
		params map[string]string // name: type
	}

	tt := []struct {
		name       string
		srv1, srv2 service
		err        error
	}{
		{
			name: "diff 1-part",
			srv1: service{method: "GET", path: "/a"},
			srv2: service{method: "GET", path: "/b"},
			err:  nil,
		},
		{
			name: "diff 2-part",
			srv1: service{method: "GET", path: "/a/b"},
			srv2: service{method: "GET", path: "/a/c"},
			err:  nil,
		},
		{
			name: "diff 1-part 2-part",
			srv1: service{method: "GET", path: "/a"},
			srv2: service{method: "GET", path: "/a/b"},
			err:  nil,
		},
		{
			name: "same 1-part",
			srv1: service{method: "GET", path: "/a"},
			srv2: service{method: "GET", path: "/a"},
			err:  ErrPatternCollision,
		},
		{
			name: "same diff method",
			srv1: service{method: "GET", path: "/a"},
			srv2: service{method: "POST", path: "/a"},
			err:  nil,
		},
		{
			name: "collide 2nd part",
			srv1: service{method: "GET", path: "/a/b"},
			srv2: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			err:  ErrPatternCollision,
		},
		{
			name: "collide 2nd part invert",
			srv1: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			srv2: service{method: "GET", path: "/a/b"},
			err:  ErrPatternCollision,
		},
		{
			name: "diff 2nd part incompatible type",
			srv1: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			srv2: service{method: "GET", path: "/a/b"},
			err:  nil,
		},
		{
			name: "middle path collision",
			srv1: service{method: "GET", path: "/a/b/c"},
			srv2: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "string"}},
			err:  ErrPatternCollision,
		},
		{
			name: "middle path collision inverted",
			srv1: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "string"}},
			srv2: service{method: "GET", path: "/a/b/c"},
			err:  ErrPatternCollision,
		},
		{
			name: "diff middle path collision type",
			srv1: service{method: "GET", path: "/a/b/c"},
			srv2: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "uint"}},
			err:  nil,
		},
		{
			name: "diff middle path collision type inverted",
			srv1: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "uint"}},
			srv2: service{method: "GET", path: "/a/b/c"},
			err:  nil,
		},
		{
			name: "collide left additional part",
			srv1: service{method: "GET", path: "/a/123/c"},
			srv2: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			err:  nil,
		},
		{
			name: "collide left additional part inverted",
			srv1: service{method: "GET", path: "/a/123"},
			srv2: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "uint"}},
			err:  nil,
		},
		{
			name: "collide uint",
			srv1: service{method: "GET", path: "/a/123"},
			srv2: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			err:  ErrPatternCollision,
		},
		{
			name: "colliding end captures",
			srv1: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			srv2: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			err:  ErrPatternCollision,
		},
		{
			name: "colliding end captures diff types",
			srv1: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			srv2: service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			err:  ErrPatternCollision,
		},
		{
			name: "colliding middle captures",
			srv1: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "uint"}},
			srv2: service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "uint"}},
			err:  ErrPatternCollision,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			srv.AddInputValidator(validator.StringType{})
			srv.AddInputValidator(validator.UintType{})

			var (
				params1 = serializeParams(t, tc.srv1.params)
				params2 = serializeParams(t, tc.srv2.params)
			)

			raw := fmt.Sprintf(`[
					{ "method": "%s", "path": "%s", "info": "info", "in": %s },
					{ "method": "%s", "path": "%s", "info": "info", "in": %s }
				]`,
				tc.srv1.method, tc.srv1.path, params1,
				tc.srv2.method, tc.srv2.path, params2,
			)

			err := srv.Parse(strings.NewReader(raw))
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}

func TestMatchSimple(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Config string
		URL    string
		Match  bool
	}{
		{ // false positive -1
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/",
			false,
		},
		{ // false positive +1
			`[ {
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {}
			} ]`,
			"/a",
			false,
		},
		{ // root url
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/",
			false,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/",
			false,
		},
		{
			`[ {
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {}
			} ]`,
			"/",
			true,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/a",
			true,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/a/",
			true,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a",
					"info": "info",
					"in": {}
			} ]`,
			"/a?param=value",
			true,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a/{id}",
					"info": "info",
					"in": {
						"{id}": {
							"info": "info",
							"type": "bool",
							"name": "id"
						}
					}
			} ]`,
			"/a/12/",
			false,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a/{id}",
					"info": "info",
					"in": {
						"{id}": {
							"info": "info",
							"type": "int",
							"name": "id"
						}
					}
			} ]`,
			"/a/12/",
			true,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a/{valid}",
					"info": "info",
					"in": {
						"{valid}": {
							"info": "info",
							"type": "bool",
							"name": "valid"
						}
					}
			} ]`,
			"/a/12/",
			false,
		},
		{
			`[ {
					"method": "GET",
					"path": "/a/{valid}",
					"info": "info",
					"in": {
						"{valid}": {
							"info": "info",
							"type": "bool",
							"name": "valid"
						}
					}
			} ]`,
			"/a/true/",
			true,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			srv.Input = append(srv.Input, validator.AnyType{})
			srv.Input = append(srv.Input, validator.IntType{})
			srv.Input = append(srv.Input, validator.BoolType{})
			err := srv.Parse(strings.NewReader(test.Config))

			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			if len(srv.Services) != 1 {
				t.Fatalf("expected to have 1 service, got %d", len(srv.Services))
			}

			req := httptest.NewRequest(http.MethodGet, test.URL, nil)

			match := srv.Services[0].Match(req)
			if test.Match && !match {
				t.Fatalf("expected %q to match", test.URL)
			}
			if !test.Match && match {
				t.Fatalf("expected %q NOT to match", test.URL)
			}
		})
	}

}

func TestFindPriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Config       string
		URL          string
		MatchingDesc string
	}{
		{
			`[
				{ "method": "GET", "path": "/a", "info": "s1" },
				{ "method": "GET", "path": "/",  "info": "s2" }
			]`,
			"/",
			"s2",
		},
		{
			`[
				{ "method": "GET", "path": "/",  "info": "s2" },
				{ "method": "GET", "path": "/a", "info": "s1" }
			]`,
			"/",
			"s2",
		},
		{
			`[
				{ "method": "GET", "path": "/a", "info": "s1" },
				{ "method": "GET", "path": "/",  "info": "s2" }
			]`,
			"/a",
			"s1",
		},
		{
			`[
				{ "method": "GET", "path": "/a/b/c",  "info": "s1" },
				{ "method": "GET", "path": "/a/b",    "info": "s2" }
			]`,
			"/a/b/c",
			"s1",
		},
		{
			`[
				{ "method": "GET", "path": "/a/b/c",  "info": "s1" },
				{ "method": "GET", "path": "/a/b",    "info": "s2" }
			]`,
			"/a/b/",
			"s2",
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			srv.Input = append(srv.Input, validator.AnyType{})
			srv.Input = append(srv.Input, validator.IntType{})
			srv.Input = append(srv.Input, validator.BoolType{})
			err := srv.Parse(strings.NewReader(test.Config))

			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

				t.Fatalf("expected to find a service")
			}
				t.Fatalf("invalid description\nactual: %q\nexpect: %q", svc.Description, tc.info)
			}
		})
	}
}
