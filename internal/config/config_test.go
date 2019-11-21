package config

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestLegalServiceName(t *testing.T) {
	tests := []struct {
		Raw   string
		Error error
	}{
		{
			`{
				"/": {
					"invalid/service-name": {

					}
				}
			}`,
			ErrFormat.Wrap(ErrIllegalServiceName.WrapString("invalid/service-name")),
		},
		{
			`{
				"/": {
					"invalid/service/name": {

					}
				}
			}`,
			ErrFormat.Wrap(ErrIllegalServiceName.WrapString("invalid/service/name")),
		},
		{
			`{
				"/": {
					"invalid-service-name": {

					}
				}
			}`,
			ErrFormat.Wrap(ErrIllegalServiceName.WrapString("invalid-service-name")),
		},

		{
			`{
				"/": {
					"valid.service_name": {
					}
				}
			}`,
			nil,
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
				if err.Error() != test.Error.Error() {
					t.Errorf("expected the error '%s' (got '%s')", test.Error.Error(), err.Error())
					t.FailNow()
				}
			}
		})
	}
}
func TestAvailableMethods(t *testing.T) {
	reader := strings.NewReader(`{
		"GET": { "info": "info" },
		"POST": { "info": "info" },
		"PUT": { "info": "info" },
		"DELETE": { "info": "info" }
	}`)
	srv, err := Parse(reader)
	if err != nil {
		t.Errorf("unexpected error (got '%s')", err)
		t.FailNow()
	}

	if srv.Method(http.MethodGet) == nil {
		t.Errorf("expected method GET to be available")
		t.Fail()
	}
	if srv.Method(http.MethodPost) == nil {
		t.Errorf("expected method POST to be available")
		t.Fail()
	}
	if srv.Method(http.MethodPut) == nil {
		t.Errorf("expected method PUT to be available")
		t.Fail()
	}
	if srv.Method(http.MethodDelete) == nil {
		t.Errorf("expected method DELETE to be available")
		t.Fail()
	}

	if srv.Method(http.MethodPatch) != nil {
		t.Errorf("expected method PATH to be UNavailable")
		t.Fail()
	}
}
func TestParseEmpty(t *testing.T) {
	reader := strings.NewReader(`{}`)
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
		Raw   string
		Error error
	}{
		{ // missing description
			`{
				"GET": {

				}
			}`,
			ErrFormat.Wrap(ErrMissingMethodDesc.WrapString("GET /")),
		},
		{ // empty description
			`{
				"GET": {
					"info": ""
				}
			}`,
			ErrFormat.Wrap(ErrMissingMethodDesc.WrapString("GET /")),
		},
		{ // valid description
			`{
				"GET": {
					"info": "a"
				}
			}`,
			nil,
		},
		{ // valid description
			`{
				"GET": {
					"info": "some description"
				}
			}`,
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
				if err.Error() != test.Error.Error() {
					t.Errorf("expected the error '%s' (got '%s')", test.Error.Error(), err.Error())
					t.FailNow()
				}
			}
		})
	}

}

func TestParseParameters(t *testing.T) {
	tests := []struct {
		Raw              string
		Error            error
		ErrorAlternative error
	}{
		{ // invalid param name prefix
			`{
				"GET": {
					"info": "info",
					"in": {
						"_param1": {}
					}
				}
			}`,
			ErrFormat.Wrap(ErrIllegalParamName.WrapString("GET / {_param1}")),
			nil,
		},
		{ // invalid param name suffix
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1_": {}
					}
				}
			}`,
			ErrFormat.Wrap(ErrIllegalParamName.WrapString("GET / {param1_}")),
			nil,
		},

		{ // missing param description
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": {}
					}
				}
			}`,
			ErrFormat.Wrap(ErrMissingParamDesc.WrapString("GET / {param1}")),
			nil,
		},
		{ // empty param description
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": {
							"info": ""
						}
					}
				}
			}`,
			ErrFormat.Wrap(ErrMissingParamDesc.WrapString("GET / {param1}")),
			nil,
		},

		{ // missing param type
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": {
							"info": "valid"
						}
					}
				}
			}`,
			ErrFormat.Wrap(ErrMissingParamType.WrapString("GET / {param1}")),
			nil,
		},
		{ // empty param type
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": {
							"info": "valid",
							"type": ""
						}
					}
				}
			}`,
			ErrFormat.Wrap(ErrMissingParamType.WrapString("GET / {param1}")),
			nil,
		},
		{ // valid description + valid type
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": {
							"info": "valid",
							"type": "valid"
						}
					}
				}
			}`,
			nil,
			nil,
		},

		{ // name conflict with rename
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "valid" },
						"param2": { "info": "valid", "type": "valid", "name": "param1" }

					}
				}
			}`,
			// 2 possible errors as map order is not deterministic
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param1}")),
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param2}")),
		},
		{ // rename conflict with name
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "valid", "name": "param2" },
						"param2": { "info": "valid", "type": "valid" }

					}
				}
			}`,
			// 2 possible errors as map order is not deterministic
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param1}")),
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param2}")),
		},
		{ // rename conflict with rename
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "valid", "name": "conflict" },
						"param2": { "info": "valid", "type": "valid", "name": "conflict" }

					}
				}
			}`,
			// 2 possible errors as map order is not deterministic
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param1}")),
			ErrFormat.Wrap(ErrParamNameConflict.WrapString("GET / {param2}")),
		},

		{ // both renamed with no conflict
			`{
				"GET": {
					"info": "info",
					"in": {
						"param1": { "info": "valid", "type": "valid", "name": "freename" },
						"param2": { "info": "valid", "type": "valid", "name": "freename2" }

					}
				}
			}`,
			nil,
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
				if err.Error() != test.Error.Error() && err.Error() != test.ErrorAlternative.Error() {
					t.Errorf("expected the error '%s' (got '%s')", test.Error.Error(), err.Error())
					t.FailNow()
				}
			}
		})
	}

}

func TestBrowseSimple(t *testing.T) {
	tests := []struct {
		Raw         string
		Path        []string
		BrowseDepth int
		ValidDepth  bool
	}{
		{ // false positive -1
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						}
					}
				}
			}`,
			[]string{"parent", "subdir"},
			1,
			false,
		},
		{ // false positive +1
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						}
					}
				}
			}`,
			[]string{"parent", "subdir"},
			3,
			false,
		},

		{
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						}
					}
				}
			}`,
			[]string{"parent", "subdir"},
			2,
			true,
		},
		{ // unknown path
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						}
					}
				}
			}`,
			[]string{"x", "y"},
			2,
			false,
		},
		{ // unknown path
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						}
					}
				}
			}`,
			[]string{"parent", "y"},
			1,
			true,
		},
		{ // Warning: this case is important to understand the precedence of service paths over
			// the value of some variables. Here if we send a string parameter in the GET method that
			// unfortunately is equal to 'subdir', it will call the sub-service /parent/subdir' instead
			// of the service /parent with its parameter set to the value 'subdir'.
			`{
				"/" : {
					"parent": {
						"/": {
							"subdir": {}
						},
						"GET": {
							"info": "valid-desc",
							"in": {
								"some-value": {
									"info": "valid-desc",
									"type": "valid-type"
								}
							}
						}
					}
				}
			}`,
			[]string{"parent", "subdir"},
			2,
			true,
		},
	}

	for i, test := range tests {

		t.Run(fmt.Sprintf("method.%d", i), func(t *testing.T) {
			srv, err := Parse(strings.NewReader(test.Raw))

			if err != nil {
				t.Errorf("unexpected error: '%s'", err)
				t.FailNow()
			}

			_, depth := srv.Browse(test.Path)
			if test.ValidDepth {
				if depth != test.BrowseDepth {
					t.Errorf("expected a depth of %d (got %d)", test.BrowseDepth, depth)
					t.FailNow()
				}
			} else {
				if depth == test.BrowseDepth {
					t.Errorf("expected a depth NOT %d (got %d)", test.BrowseDepth, depth)
					t.FailNow()
				}

			}
		})
	}

}
