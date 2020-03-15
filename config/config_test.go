package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.xdrm.io/go/aicra/config/datatype/builtin"
)

func TestLegalServiceName(t *testing.T) {
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
			nil,
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
			nil,
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
			_, err := Parse(strings.NewReader(test.Raw))

			if err == nil && test.Error != nil {
				t.Errorf("expected an error: '%s'", test.Error.Error())
				t.FailNow()
			}
			if err != nil && test.Error == nil {
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if err != nil && test.Error != nil {
				if !errors.Is(err, test.Error) {
					t.Errorf("expected the error '%s' (got '%s')", test.Error.Error(), err.Error())
					t.FailNow()
				}
			}
		})
	}
}
func TestAvailableMethods(t *testing.T) {
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
			_, err := Parse(strings.NewReader(test.Raw))

			if test.ValidMethod && err != nil {
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if !test.ValidMethod && !errors.Is(err, ErrUnknownMethod) {
				t.Errorf("expected error <%s> got <%s>", ErrUnknownMethod, err)
				t.FailNow()
			}
		})
	}
}
func TestParseEmpty(t *testing.T) {
	reader := strings.NewReader(`[]`)
	_, err := Parse(reader)
	if err != nil {
		t.Errorf("unexpected error (got '%s')", err)
		t.FailNow()
	}
}
func TestParseJsonError(t *testing.T) {
	reader := strings.NewReader(`{
		"GET": {
			"info": "info
		},
	}`) // trailing ',' is invalid JSON
	_, err := Parse(reader)
	if err == nil {
		t.Errorf("expected error")
		t.FailNow()
	}
}

func TestParseMissingMethodDescription(t *testing.T) {
	tests := []struct {
		Raw              string
		ValidDescription bool
	}{
		{ // missing description
			`[ { "method": "GET", "path": "/" }]`,
			false,
		},
		{ // missing description
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
			_, err := Parse(strings.NewReader(test.Raw))

			if test.ValidDescription && err != nil {
				t.Errorf("unexpected error: '%s'", err)
				t.FailNow()
			}

			if !test.ValidDescription && !errors.Is(err, ErrMissingDescription) {
				t.Errorf("expected error <%s> got <%s>", ErrMissingDescription, err)
				t.FailNow()
			}
		})
	}

}

func TestParamEmptyRenameNoRename(t *testing.T) {
	reader := strings.NewReader(`[
		{
			"method": "GET",
			"path": "/",
			"info": "valid-description",
			"in": {
				"original": { "info": "valid-desc", "type": "any", "name": "" }
			}
		}
	]`)
	srv, err := Parse(reader, builtin.AnyDataType{})
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv.services) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}

	for _, param := range srv.services[0].Input {
		if param.Rename != "original" {
			t.Errorf("expected the parameter 'original' not to be renamed to '%s'", param.Rename)
			t.FailNow()
		}
	}

}
func TestOptionalParam(t *testing.T) {
	reader := strings.NewReader(`[
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
	srv, err := Parse(reader, builtin.AnyDataType{}, builtin.BoolDataType{})
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv.services) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}
	for pName, param := range srv.services[0].Input {

		if pName == "optional" || pName == "optional2" {
			if !param.Optional {
				t.Errorf("expected parameter '%s' to be optional", pName)
				t.Failed()
			}
		}
		if pName == "required" || pName == "required2" {
			if param.Optional {
				t.Errorf("expected parameter '%s' to be required", pName)
				t.Failed()
			}
		}
	}

}
func TestParseParameters(t *testing.T) {
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
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			_, err := Parse(strings.NewReader(test.Raw), builtin.AnyDataType{})

			if err == nil && test.Error != nil {
				t.Errorf("expected an error: '%s'", test.Error.Error())
				t.FailNow()
			}
			if err != nil && test.Error == nil {
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if err != nil && test.Error != nil {
				if !errors.Is(err, test.Error) {
					t.Errorf("expected the error <%s> got <%s>", test.Error, err)
					t.FailNow()
				}
			}
		})
	}

}

func TestMatchSimple(t *testing.T) {
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
					"path": "/a/{id}",
					"info": "info",
					"in": {
						"{id}": {
							"info": "info",
							"type": "bool"
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
							"type": "int"
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
							"type": "bool"
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
							"type": "bool"
						}
					}
			} ]`,
			"/a/true/",
			true,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv, err := Parse(strings.NewReader(test.Config), builtin.AnyDataType{}, builtin.IntDataType{}, builtin.BoolDataType{})

			if err != nil {
				t.Errorf("unexpected error: '%s'", err)
				t.FailNow()
			}

			if len(srv.services) != 1 {
				t.Errorf("expected to have 1 service, got %d", len(srv.services))
				t.FailNow()
			}

			req := httptest.NewRequest(http.MethodGet, test.URL, nil)

			match := srv.services[0].Match(req)
			if test.Match && !match {
				t.Errorf("expected '%s' to match", test.URL)
				t.FailNow()
			}
			if !test.Match && match {
				t.Errorf("expected '%s' NOT to match", test.URL)
				t.FailNow()
			}
		})
	}

}
