package config

import (
	"errors"
	"fmt"
	"strings"
	"testing"
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
			ErrInvalidPatternBracePosition,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}a" } ]`,
			ErrInvalidPatternBracePosition,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}" } ]`,
			nil,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/s{braces}/abc" } ]`,
			ErrInvalidPatternBracePosition,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}s/abc" } ]`,
			ErrInvalidPatternBracePosition,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/abc" } ]`,
			nil,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{b{races}s/abc" } ]`,
			ErrInvalidPatternOpeningBrace,
		},
		{
			`[ { "method": "GET", "info": "a", "path": "/invalid/{braces}/}abc" } ]`,
			ErrInvalidPatternClosingBrace,
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
				"original": { "info": "valid-desc", "type": "valid-type", "name": "" }
			}
		}
	]`)
	srv, err := Parse(reader)
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}

	for _, param := range (srv)[0].Input {
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
				"optional": { "info": "valid-desc", "type": "?optional-type" },
				"required": { "info": "valid-desc", "type": "required-type" },
				"required2": { "info": "valid-desc", "type": "a" },
				"optional2": { "info": "valid-desc", "type": "?a" }
			}
		}
	]`)
	srv, err := Parse(reader)
	if err != nil {
		t.Errorf("unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(srv) < 1 {
		t.Errorf("expected a service")
		t.FailNow()
	}
	for pName, param := range (srv)[0].Input {

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
			ErrIllegalParamName,
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
			ErrIllegalParamName,
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
						"param1": { "info": "valid", "type": "a" }
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
						"param1": { "info": "valid", "type": "?valid" }
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
						"param1": { "info": "valid", "type": "valid" },
						"param2": { "info": "valid", "type": "valid", "name": "param1" }
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
						"param1": { "info": "valid", "type": "valid", "name": "param2" },
						"param2": { "info": "valid", "type": "valid" }
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
						"param1": { "info": "valid", "type": "valid", "name": "conflict" },
						"param2": { "info": "valid", "type": "valid", "name": "conflict" }
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
						"param1": { "info": "valid", "type": "valid", "name": "freename" },
						"param2": { "info": "valid", "type": "valid", "name": "freename2" }
					}
				}
			]`,
			nil,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
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
					t.Errorf("expected the error <%s> got <%s>", test.Error, err)
					t.FailNow()
				}
			}
		})
	}

}

// todo: rewrite with new api format
// func TestMatchSimple(t *testing.T) {
// 	tests := []struct {
// 		Raw         string
// 		Path        []string
// 		BrowseDepth int
// 		ValidDepth  bool
// 	}{
// 		{ // false positive -1
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"parent", "subdir"},
// 			1,
// 			false,
// 		},
// 		{ // false positive +1
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"parent", "subdir"},
// 			3,
// 			false,
// 		},

// 		{
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"parent", "subdir"},
// 			2,
// 			true,
// 		},
// 		{ // unknown path
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"x", "y"},
// 			2,
// 			false,
// 		},
// 		{ // unknown path
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"parent", "y"},
// 			1,
// 			true,
// 		},
// 		{ // Warning: this case is important to understand the precedence of service paths over
// 			// the value of some variables. Here if we send a string parameter in the GET method that
// 			// unfortunately is equal to 'subdir', it will call the sub-service /parent/subdir' instead
// 			// of the service /parent with its parameter set to the value 'subdir'.
// 			`{
// 				"/" : {
// 					"parent": {
// 						"/": {
// 							"subdir": {}
// 						},
// 						"GET": {
// 							"info": "valid-desc",
// 							"in": {
// 								"some-value": {
// 									"info": "valid-desc",
// 									"type": "valid-type"
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}`,
// 			[]string{"parent", "subdir"},
// 			2,
// 			true,
// 		},
// 	}

// 	for i, test := range tests {

// 		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
// 			srv, err := Parse(strings.NewReader(test.Raw))

// 			if err != nil {
// 				t.Errorf("unexpected error: '%s'", err)
// 				t.FailNow()
// 			}

// 			_, depth := srv.Match(test.Path)
// 			if test.ValidDepth {
// 				if depth != test.BrowseDepth {
// 					t.Errorf("expected a depth of %d (got %d)", test.BrowseDepth, depth)
// 					t.FailNow()
// 				}
// 			} else {
// 				if depth == test.BrowseDepth {
// 					t.Errorf("expected a depth NOT %d (got %d)", test.BrowseDepth, depth)
// 					t.FailNow()
// 				}

// 			}
// 		})
// 	}

// }
