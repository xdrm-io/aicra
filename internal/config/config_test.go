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

func TestNoOpValidator(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string

		valName string
		valType reflect.Type

		typename string
		value    interface{}
		match    bool
	}{
		{
			name:     "string mismatch",
			valName:  "string",
			valType:  reflect.TypeOf(""),
			typename: "uint",
			value:    "abc",
		},
		{
			name:     "string match",
			valName:  "string",
			valType:  reflect.TypeOf(""),
			typename: "string",
			value:    "abc",
			match:    true,
		},
		{
			name:     "uint mismatch",
			valName:  "uint",
			valType:  reflect.TypeOf(uint(0)),
			typename: "string",
			value:    uint(123),
		},
		{
			name:     "uint match",
			valName:  "uint",
			valType:  reflect.TypeOf(uint(0)),
			typename: "uint",
			value:    uint(123),
			match:    true,
		},
		{
			name:     "uint match invalid",
			valName:  "uint",
			valType:  reflect.TypeOf(uint(0)),
			typename: "uint",
			value:    "abc", // validation is never used, anything is considered valid
			match:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v := noOp{name: tc.valName, goType: tc.valType}

			if v.GoType() != tc.valType {
				t.Fatalf("invalid go type\nactual: %v\nexpect: %v", v.GoType(), tc.valType)
			}

			validate := v.Validator(tc.typename)
			match := (validate != nil)
			if match != tc.match {
				t.Fatalf("invalid match\nactual: %t\nexpect: %t", match, tc.match)
			}
			if !tc.match || !match {
				return
			}
			if _, valid := validate(tc.value); !valid {
				t.Fatalf("expect to always be valid")
			}
		})
	}
}

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

func TestLegalServicePath(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		conf string
		err  error
	}{
		{
			name: "empty path",
			conf: `[ { "method": "GET", "info": "a", "path": "" } ]`,
			err:  ErrInvalidPattern,
		},
		{
			name: "path no starting slash",
			conf: `[ { "method": "GET", "info": "a", "path": "no-starting-slash" } ]`,
			err:  ErrInvalidPattern,
		},
		{
			name: "path ending slash",
			conf: `[ { "method": "GET", "info": "a", "path": "ending-slash/" } ]`,
			err:  ErrInvalidPattern,
		},
		{
			name: "root path",
			conf: `[ { "method": "GET", "info": "a", "path": "/" } ]`,
			err:  nil,
		},
		{
			name: "valid path",
			conf: `[ { "method": "GET", "info": "a", "path": "/valid-name" } ]`,
			err:  nil,
		},
		{
			name: "valid nested path",
			conf: `[ { "method": "GET", "info": "a", "path": "/valid/nested/name" } ]`,
			err:  nil,
		},
		{
			name: "capture not after slash",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
		{
			name: "capture not before slash",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{braces}a" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
		{
			name: "valid ending capture",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{braces}" } ]`,
			err:  ErrUndefinedBraceCapture,
		},
		{
			name: "valid ending capture case",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{BrAcEs}" } ]`,
			err:  ErrUndefinedBraceCapture,
		},
		{
			name: "invalid middle capture before slash",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}/abc" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
		{
			name: "invalid middle capture after slash",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{braces}s/abc" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
		{
			name: "valid middle capture",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/abc" } ]`,
			err:  ErrUndefinedBraceCapture,
		},
		{
			name: "invalid middle capture with {",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{b{races}s/abc" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
		{
			name: "valid middle capture invalid } after slash",
			conf: `[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/}abc" } ]`,
			err:  ErrInvalidPatternBraceCapture,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(tc.conf))
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}
func TestAvailableMethods(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name        string
		conf        string
		validMethod bool
	}{
		{
			name:        "missing description",
			conf:        `[ { "method": "GET", "path": "/", "info": "valid-description" }]`,
			validMethod: true,
		},
		{
			name:        "missing description",
			conf:        `[ { "method": "POST", "path": "/", "info": "valid-description" }]`,
			validMethod: true,
		},
		{
			name:        "empty description",
			conf:        `[ { "method": "PUT", "path": "/", "info": "valid-description" }]`,
			validMethod: true,
		},
		{
			name:        "empty trimmed description",
			conf:        `[ { "method": "DELETE", "path": "/", "info": "valid-description" }]`,
			validMethod: true,
		},
		{
			name:        "valid description",
			conf:        `[ { "method": "get", "path": "/", "info": "valid-description" }]`,
			validMethod: false,
		},
		{
			name:        "valid description",
			conf:        `[ { "method": "UNknOwN", "path": "/", "info": "valid-description" }]`,
			validMethod: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(tc.conf))
			if tc.validMethod && err != nil {
				t.Fatalf("unexpected error: %q", err.Error())
			}
			if !tc.validMethod && !errors.Is(err, ErrUnknownMethod) {
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
		"GET": { "info": "info },
	}`) // trailing ',' is invalid JSON
	srv := &Server{}
	err := srv.Parse(r)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseMissingMethodDescription(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name      string
		conf      string
		validDesc bool
	}{
		{
			name:      "missing description",
			conf:      `[ { "method": "GET", "path": "/" }]`,
			validDesc: false,
		},
		{
			name:      "missing descriptiontype",
			conf:      `[ { "method": "GET", "path": "/subservice" }]`,
			validDesc: false,
		},
		{
			name:      "empty description",
			conf:      `[ { "method": "GET", "path": "/", "info": "" }]`,
			validDesc: false,
		},
		{
			name:      "empty trimmed description",
			conf:      `[ { "method": "GET", "path": "/", "info": " " }]`,
			validDesc: false,
		},
		{
			name:      "valid description",
			conf:      `[ { "method": "GET", "path": "/", "info": "a" }]`,
			validDesc: true,
		},
		{
			name:      "valid description",
			conf:      `[ { "method": "GET", "path": "/", "info": "some description" }]`,
			validDesc: true,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(tc.conf))
			if tc.validDesc && err != nil {
				t.Fatalf("unexpected error: %q", err)
			}
			if !tc.validDesc && !errors.Is(err, ErrMissingDescription) {
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
	srv.AddInputValidator(validator.AnyType{})
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
	srv.AddInputValidator(validator.AnyType{})
	srv.AddInputValidator(validator.BoolType{})
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
				t.Fatalf("expected parameter %q to be optional", pName)
			}
		}
		if pName == "required" || pName == "required2" {
			if param.Optional {
				t.Fatalf("expected parameter %q to be required", pName)
			}
		}
	}

}
func TestParseParameters(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		conf string
		err  error
	}{
		{
			name: "invalid param name prefix",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"_param1": { }
				}
			} ]`,
			err: ErrMissingParamDesc,
		},
		{
			name: "invalid param name suffix",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1_": { }
				}
			} ]`,
			err: ErrMissingParamDesc,
		},

		{
			name: "missing param info",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { }
				}
			} ]`,
			err: ErrMissingParamDesc,
		},
		{
			name: "empty param info",
			conf: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"in": {
						"param1": { "info": "" }
					}
				}
			]`,
			err: ErrMissingParamDesc,
		},

		{
			name: "missing param type",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid" }
				}
			} ]`,
			err: ErrMissingParamType,
		},
		{
			name: "empty param type",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "" }
				}
			} ]`,
			err: ErrMissingParamType,
		},
		{
			name: "invalid type: optional mark only",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "?" }
				}
			} ]`,
			err: ErrMissingParamType,
		},
		{
			name: "valid description and type",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "any" }
				}
			} ]`,
			err: nil,
		},
		{
			name: "valid description and optional type",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "?any" }
				}
			} ]`,
			err: nil,
		},

		{
			name: "name conflict with rename",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "any" },
					"param2": { "info": "valid", "type": "any", "name": "param1" }
				}
			} ]`,
			// 2 possible errors as map order is not deterministic
			err: ErrParamNameConflict,
		},
		{
			name: "rename conflict with name",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "any", "name": "param2" },
					"param2": { "info": "valid", "type": "any" }
				}
			} ]`,
			// 2 possible errors as map order is not deterministic
			err: ErrParamNameConflict,
		},
		{
			name: "rename conflict with rename",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "any", "name": "conflict" },
					"param2": { "info": "valid", "type": "any", "name": "conflict" }
				}
			} ]`,
			// 2 possible errors as map order is not deterministic
			err: ErrParamNameConflict,
		},

		{
			name: "both renamed with no conflict",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"param1": { "info": "valid", "type": "any", "name": "freename" },
					"param2": { "info": "valid", "type": "any", "name": "freename2" }
				}
			} ]`,
			err: nil,
		},
		{
			name: "missing uri rename",
			conf: `[ {
				"method": "GET",
				"path": "/{uri}",
				"info": "info",
				"in": {
					"{uri}": { "info": "valid", "type": "any" }
				}
			} ]`,
			err: ErrMandatoryRename,
		},
		{
			name: "valid uri with rename",
			conf: `[ {
				"method": "GET",
				"path": "/{uri}",
				"info": "info",
				"in": {
					"{uri}": { "info": "valid", "type": "any", "name": "freename" }
				}
			} ]`,
			err: nil,
		},
		{
			name: "missing query rename",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"GET@abc": { "info": "valid", "type": "any" }
				}
			} ]`,
			err: ErrMandatoryRename,
		},
		{
			name: "valid query with rename",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"GET@abc": { "info": "valid", "type": "any", "name": "abc" }
				}
			} ]`,
			err: nil,
		},
		{
			name: "valid query case with rename",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"GET@AbC": { "info": "valid", "type": "any", "name": "abc" }
				}
			} ]`,
			err: nil,
		},

		{
			name: "optional uri",
			conf: `[ {
				"method": "GET",
				"path": "/{uri}",
				"info": "info",
				"in": {
					"{uri}": { "info": "valid", "type": "?any", "name": "freename" }
				}
			} ]`,
			err: ErrIllegalOptionalURIParam,
		},
		{
			name: "uri missing in path",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {
					"{uri}": { "info": "valid", "type": "?any", "name": "freename" }
				}
			} ]`,
			err: ErrUnspecifiedBraceCapture,
		},
		{
			name: "missing uri from path",
			conf: `[ {
				"method": "GET",
				"path": "/{uri}",
				"info": "info",
				"in": { }
			} ]`,
			err: ErrUndefinedBraceCapture,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			srv.AddInputValidator(validator.AnyType{})
			err := srv.Parse(strings.NewReader(tc.conf))
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
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

func TestServiceCollisionPanic(t *testing.T) {
	t.Parallel()

	type service struct {
		method string
		path   string
		params map[string]string // name: type
	}

	tt := []struct {
		name       string
		srv1, srv2 service
		panics     bool
	}{
		{
			name: "same 1-part",
			srv1: service{method: "GET", path: "/a"},
			srv2: service{method: "GET", path: "/a"},
		},
		{
			name:   "collide 2nd part",
			srv1:   service{method: "GET", path: "/a/b"},
			srv2:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			panics: true,
		},
		{
			name:   "collide 2nd part invert",
			srv1:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			srv2:   service{method: "GET", path: "/a/b"},
			panics: true,
		},
		{
			name:   "middle path collision",
			srv1:   service{method: "GET", path: "/a/b/c"},
			srv2:   service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "string"}},
			panics: true,
		},
		{
			name:   "middle path collision inverted",
			srv1:   service{method: "GET", path: "/a/{var}/c", params: map[string]string{"{var}": "string"}},
			srv2:   service{method: "GET", path: "/a/b/c"},
			panics: true,
		},
		{
			name:   "colliding end captures",
			srv1:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			srv2:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			panics: false, // exit before we can panic
		},
		{
			name:   "colliding end captures diff types",
			srv1:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "uint"}},
			srv2:   service{method: "GET", path: "/a/{var}", params: map[string]string{"{var}": "string"}},
			panics: false, // exit before we can panic
		},
	}

	for _, tc := range tt {
		t.Run("missing param:"+tc.name, func(t *testing.T) {
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

			// parse first but ignore error
			srv.Parse(strings.NewReader(raw))

			// remove parameters
			for _, svc := range srv.Services {
				svc.Input = map[string]*Parameter{}
			}

			// recheck for collissions
			defer func() {
				r := recover()
				if r == nil && tc.panics {
					t.Fatalf("expected a panic")
				}
				if r != nil && !tc.panics {
					t.Fatalf("unexpected panic")
				}
			}()
			srv.collide()
		})

		t.Run("nil validator:"+tc.name, func(t *testing.T) {
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

			// parse first but ignore error
			srv.Parse(strings.NewReader(raw))

			// remove param validators
			for _, svc := range srv.Services {
				for i := range svc.Input {
					svc.Input[i].Validator = nil
				}
			}

			// recheck for collissions
			defer func() {
				r := recover()
				if r == nil && tc.panics {
					t.Fatalf("expected a panic")
				}
				if r != nil && !tc.panics {
					t.Fatalf("unexpected panic")
				}
			}()
			srv.collide()
		})
	}
}

func TestMatchSimple(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		conf  string
		uri   string
		match bool
	}{
		{
			name: "incomplete uri",
			conf: `[ {
				"method": "GET",
				"path": "/a",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/",
			match: false,
		},
		{
			name: "additional parts",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/a",
			match: false,
		},
		{
			name: "root match",
			conf: `[ {
				"method": "GET",
				"path": "/",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/",
			match: true,
		},
		{
			name: "1part uri match",
			conf: `[ {
				"method": "GET",
				"path": "/a",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/a",
			match: true,
		},
		{
			name: "1part uri match ending slash",
			conf: `[ {
				"method": "GET",
				"path": "/a",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/a/",
			match: true,
		},
		{
			name: "1part uri match ending slashes",
			conf: `[ {
				"method": "GET",
				"path": "/a",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/a///////",
			match: true,
		},
		{
			name: "1part ignored query",
			conf: `[ {
				"method": "GET",
				"path": "/a",
				"info": "info",
				"in": {}
			} ]`,
			uri:   "/a?param=value",
			match: true,
		},
		{
			name: "2part mismatching bool",
			conf: `[ {
				"method": "GET",
				"path": "/a/{id}",
				"info": "info",
				"in": {
					"{id}": { "info": "info", "type": "bool", "name": "id"
					}
				}
			} ]`,
			uri:   "/a/12/",
			match: false,
		},
		{
			name: "2part matching int",
			conf: `[ {
				"method": "GET",
				"path": "/a/{id}",
				"info": "info",
				"in": {
					"{id}": { "info": "info", "type": "int", "name": "id" }
				}
			} ]`,
			uri:   "/a/12/",
			match: true,
		},
		{
			name: "2part matching bool",
			conf: `[ {
				"method": "GET",
				"path": "/a/{valid}",
				"info": "info",
				"in": {
					"{valid}": { "info": "info", "type": "bool", "name": "valid" }
				}
			} ]`,
			uri:   "/a/true/",
			match: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			srv.AddInputValidator(validator.AnyType{})
			srv.AddInputValidator(validator.IntType{})
			srv.AddInputValidator(validator.BoolType{})
			err := srv.Parse(strings.NewReader(tc.conf))

			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			if len(srv.Services) != 1 {
				t.Fatalf("expected to have 1 service, got %d", len(srv.Services))
			}

			req := httptest.NewRequest(http.MethodGet, tc.uri, nil)
			match := srv.Services[0].Match(req)
			if tc.match && !match {
				t.Fatalf("expected %q to match", tc.uri)
			}
			if !tc.match && match {
				t.Fatalf("expected %q NOT to match", tc.uri)
			}
		})
	}

}

func TestFindPriority(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		conf  string
		uri   string
		match bool
		info  string
	}{
		{
			name: "mismatch",
			conf: `[
				{ "method": "GET", "path": "/a", "info": "s1" },
				{ "method": "GET", "path": "/",  "info": "s2" }
				]`,
			uri:   "/b",
			match: false,
		},
		{
			name: "match root last",
			conf: `[
				{ "method": "GET", "path": "/a", "info": "s1" },
				{ "method": "GET", "path": "/",  "info": "s2" }
				]`,
			uri:   "/",
			match: true,
			info:  "s2",
		},
		{
			name: "match root first",
			conf: `[
				{ "method": "GET", "path": "/",  "info": "s2" },
				{ "method": "GET", "path": "/a", "info": "s1" }
			]`,
			uri:   "/",
			match: true,
			info:  "s2",
		},
		{
			name: "match non-root last",
			conf: `[
				{ "method": "GET", "path": "/a", "info": "s1" },
				{ "method": "GET", "path": "/",  "info": "s2" }
			]`,
			uri:   "/a",
			match: true,
			info:  "s1",
		},
		{
			name: "match longest",
			conf: `[
				{ "method": "GET", "path": "/a/b/c",  "info": "s1" },
				{ "method": "GET", "path": "/a/b",    "info": "s2" }
			]`,
			uri:   "/a/b/c",
			match: true,
			info:  "s1",
		},
		{
			name: "match shortest",
			conf: `[
				{ "method": "GET", "path": "/a/b/c",  "info": "s1" },
				{ "method": "GET", "path": "/a/b",    "info": "s2" }
			]`,
			uri:   "/a/b/",
			match: true,
			info:  "s2",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{}
			srv.AddInputValidator(validator.AnyType{})
			srv.AddInputValidator(validator.IntType{})
			srv.AddInputValidator(validator.BoolType{})
			err := srv.Parse(strings.NewReader(tc.conf))

			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			req := httptest.NewRequest(http.MethodGet, tc.uri, nil)
			svc := srv.Find(req)
			if svc == nil && tc.match {
				t.Fatalf("expected to find a service")
			}
			if svc != nil && !tc.match {
				t.Fatalf("expected to find no service, got %q", svc.Description)
			}
			if !tc.match {
				return
			}
			if svc.Description != tc.info {
				t.Fatalf("invalid description\nactual: %q\nexpect: %q", svc.Description, tc.info)
			}
		})
	}
}

func TestScopeVars(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string

		scope [][]string
		vars  map[string]string

		// expected scopes according to the vars
		expect [][]string
	}{
		{
			name:   "convert empty first to empty",
			scope:  [][]string{{}},
			expect: [][]string{},
		},
		{
			name:  "missing brace syntax",
			scope: [][]string{{"a", "b"}, {"c", "d"}},
			vars: map[string]string{
				"a": "x1",
				"b": "x2",
			},
			expect: [][]string{{"a", "b"}, {"c", "d"}},
		},
		{
			name:  "brace syntax no matching var",
			scope: [][]string{{"[a]", "b"}, {"c", "d"}},
			vars: map[string]string{
				"b": "x1",
			},
			expect: [][]string{{"[a]", "b"}, {"c", "d"}},
		},
		{
			name:  "replace same level",
			scope: [][]string{{"[a]", "[b]"}, {"c", "d"}},
			vars: map[string]string{
				"a": "x1",
				"b": "x2",
			},
			expect: [][]string{{"[x1]", "[x2]"}, {"c", "d"}},
		},
		{
			name:  "replace different levels",
			scope: [][]string{{"[a]", "b"}, {"[c]", "d"}},
			vars: map[string]string{
				"a": "x1",
				"c": "x2",
			},
			expect: [][]string{{"[x1]", "b"}, {"[x2]", "d"}},
		},
		{
			name:  "replace multiple per scope",
			scope: [][]string{{"a", "a[b]c[d]e"}, {"f", "g[h][d][b]"}},
			vars: map[string]string{
				"b": "x1",
				"d": "x2",
				"h": "x3",
			},
			expect: [][]string{{"a", "a[x1]c[x2]e"}, {"f", "g[x3][x2][x1]"}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{
				Scope:    tc.scope,
				Captures: make([]*BraceCapture, 0, len(tc.vars)),
			}

			for name := range tc.vars {
				svc.Captures = append(svc.Captures, &BraceCapture{
					Ref: &Parameter{Rename: name},
				})
			}

			svc.cleanScope()

			// copy scope
			updated := make([][]string, len(svc.Scope))
			for a, list := range svc.Scope {
				updated[a] = make([]string, len(list))
				for b, perm := range list {
					updated[a][b] = perm
				}
			}

			// update using ScopeVars
			for _, sv := range svc.ScopeVars {
				value := tc.vars[sv.CaptureName]
				a, b := sv.Position[0], sv.Position[1]
				updated[a][b] = strings.ReplaceAll(
					updated[a][b],
					fmt.Sprintf("[%s]", sv.CaptureName),
					fmt.Sprintf("[%v]", value),
				)
			}

			// compare
			if len(updated) != len(tc.expect) {
				t.Fatalf("invalid size\nactual: %d\nexpect: %d", len(updated), len(tc.expect))
			}

			for a, expectList := range tc.expect {
				if len(updated[a]) != len(expectList) {
					t.Fatalf("invalid [%d] size\nactual: %d\nexpect: %d", a, len(updated[a]), len(expectList))
				}
				for b, expect := range expectList {
					if updated[a][b] != expect {
						t.Fatalf("invalid [%d][%d]\nactual: %q\nexpect: %q", a, b, updated[a][b], expect)
					}
				}
			}

		})
	}
}
