package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.xdrm.io/go/aicra/datatype/builtin"
)

func TestLegalServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Raw   string
		Error error
	}{
		// empty
		{
			`[ { "method": "GET", "info": "a", "path": "" } ]`,
			errInvalidPattern,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "no-starting-slash" } ]`,
			errInvalidPattern,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "ending-slash/" } ]`,
			errInvalidPattern,
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
			errInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}a" } ]`,
			errInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}" } ]`,
			errUndefinedBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}/abc" } ]`,
			errInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}s/abc" } ]`,
			errInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/abc" } ]`,
			errUndefinedBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{b{races}s/abc" } ]`,
			errInvalidPatternBraceCapture,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/}abc" } ]`,
			errInvalidPatternBraceCapture,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("service.%d", i), func(t *testing.T) {
			srv := &Server{}
			err := srv.Parse(strings.NewReader(test.Raw))

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
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if !test.ValidMethod && !errors.Is(err, errUnknownMethod) {
				t.Errorf("expected error <%s> got <%s>", errUnknownMethod, err)
				t.FailNow()
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
		t.Errorf("unexpected error (got '%s')", err)
		t.FailNow()
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
		t.Errorf("expected error")
		t.FailNow()
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
			srv := &Server{}
			err := srv.Parse(strings.NewReader(test.Raw))

			if test.ValidDescription && err != nil {
				t.Errorf("unexpected error: '%s'", err)
				t.FailNow()
			}

			if !test.ValidDescription && !errors.Is(err, errMissingDescription) {
				t.Errorf("expected error <%s> got <%s>", errMissingDescription, err)
				t.FailNow()
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
	srv.Types = append(srv.Types, builtin.AnyDataType{})
	err := srv.Parse(r)
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv.Services) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}

	for _, param := range srv.Services[0].Input {
		if param.Rename != "original" {
			t.Errorf("expected the parameter 'original' not to be renamed to '%s'", param.Rename)
			t.FailNow()
		}
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
	srv.Types = append(srv.Types, builtin.AnyDataType{})
	srv.Types = append(srv.Types, builtin.BoolDataType{})
	err := srv.Parse(r)
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv.Services) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}
	for pName, param := range srv.Services[0].Input {

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
			errMissingParamDesc,
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
			errMissingParamDesc,
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
			errMissingParamDesc,
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
			errMissingParamDesc,
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
			errMissingParamType,
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
			errMissingParamType,
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

			errMissingParamType,
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
			errParamNameConflict,
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
			errParamNameConflict,
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
			errParamNameConflict,
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
			errMandatoryRename,
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
			errMandatoryRename,
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
			errIllegalOptionalURIParam,
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
			errUnspecifiedBraceCapture,
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
			errUndefinedBraceCapture,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			srv.Types = append(srv.Types, builtin.AnyDataType{})
			err := srv.Parse(strings.NewReader(test.Raw))

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

func TestServiceCollision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Config string
		Error  error
	}{
		{
			`[
				{ "method": "GET", "path": "/a",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/b",
					"info": "info", "in": {}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a",
					"info": "info", "in": {}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/b",
					"info": "info", "in": {}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/b",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a",
					"info": "info", "in": {}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/b",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "string", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/b",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/b/d",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}/d",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "string", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "string", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123/d",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "string", "name": "c" }
					}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}/d",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "string", "name": "c" }
					}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123/d",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}/d",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			nil,
		},
		{
			`[
				{ "method": "GET", "path": "/a/123/d",
					"info": "info", "in": {}
				},
				{ "method": "GET", "path": "/a/{c}/d",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/{b}",
					"info": "info", "in": {
						"{b}": { "info":"info", "type": "uint", "name": "b" }
					}
				},
				{ "method": "GET", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			errPatternCollision,
		},
		{
			`[
				{ "method": "GET", "path": "/a/{b}",
					"info": "info", "in": {
						"{b}": { "info":"info", "type": "uint", "name": "b" }
					}
				},
				{ "method": "PUT", "path": "/a/{c}",
					"info": "info", "in": {
						"{c}": { "info":"info", "type": "uint", "name": "c" }
					}
				}
			]`,
			nil, // different methods
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv := &Server{}
			srv.Types = append(srv.Types, builtin.StringDataType{})
			srv.Types = append(srv.Types, builtin.UintDataType{})
			err := srv.Parse(strings.NewReader(test.Config))

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
			srv.Types = append(srv.Types, builtin.AnyDataType{})
			srv.Types = append(srv.Types, builtin.IntDataType{})
			srv.Types = append(srv.Types, builtin.BoolDataType{})
			err := srv.Parse(strings.NewReader(test.Config))

			if err != nil {
				t.Errorf("unexpected error: '%s'", err)
				t.FailNow()
			}

			if len(srv.Services) != 1 {
				t.Errorf("expected to have 1 service, got %d", len(srv.Services))
				t.FailNow()
			}

			req := httptest.NewRequest(http.MethodGet, test.URL, nil)

			match := srv.Services[0].Match(req)
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
