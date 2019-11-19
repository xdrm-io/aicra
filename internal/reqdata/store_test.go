package reqdata

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEmptyStore(t *testing.T) {
	store := New(nil, nil)

	if store.URI == nil {
		t.Errorf("store 'URI' list should be initialized")
		t.Fail()
	}
	if len(store.URI) != 0 {
		t.Errorf("store 'URI' list should be empty")
		t.Fail()
	}

	if store.Get == nil {
		t.Errorf("store 'Get' map should be initialized")
		t.Fail()
	}
	if store.Form == nil {
		t.Errorf("store 'Form' map should be initialized")
		t.Fail()
	}
	if store.Set == nil {
		t.Errorf("store 'Set' map should be initialized")
		t.Fail()
	}
}

func TestStoreWithUri(t *testing.T) {
	urilist := []string{"abc", "def"}
	store := New(urilist, nil)

	if len(store.URI) != len(urilist) {
		t.Errorf("store 'Set' should contain %d elements (got %d)", len(urilist), len(store.URI))
		t.Fail()
	}
	if len(store.Set) != len(urilist) {
		t.Errorf("store 'Set' should contain %d elements (got %d)", len(urilist), len(store.Set))
		t.Fail()
	}

	for i, value := range urilist {

		t.Run(fmt.Sprintf("URL#%d='%s'", i, value), func(t *testing.T) {
			key := fmt.Sprintf("URL#%d", i)
			element, isset := store.Set[key]

			if !isset {
				t.Errorf("store should contain element with key '%s'", key)
				t.Failed()
			}

			if element.Value != value {
				t.Errorf("store[%s] should return '%s' (got '%s')", key, value, element.Value)
				t.Failed()
			}
		})

	}

}

func TestStoreWithGet(t *testing.T) {
	tests := []struct {
		Query string

		InvalidNames []string
		ParamNames   []string
		ParamValues  [][]string
	}{
		{
			Query:        "",
			InvalidNames: []string{},
			ParamNames:   []string{},
			ParamValues:  [][]string{},
		},
		{
			Query:        "a",
			InvalidNames: []string{},
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
		},
		{
			Query:        "a&b",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{[]string{""}, []string{""}},
		},
		{
			Query:        "a=",
			InvalidNames: []string{},
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
		},
		{
			Query:        "a=&b=x",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{[]string{""}, []string{"x"}},
		},
		{
			Query:        "a=b&c=d",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b"}, []string{"d"}},
		},
		{
			Query:        "a=b&c=d&a=x",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b", "x"}, []string{"d"}},
		},
		{
			Query:        "a=b&_invalid=x",
			InvalidNames: []string{"_invalid"},
			ParamNames:   []string{"a", "_invalid"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{
			Query:        "a=b&invalid_=x",
			InvalidNames: []string{"invalid_"},
			ParamNames:   []string{"a", "invalid_"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{
			Query:        "a=b&GET@injection=x",
			InvalidNames: []string{"GET@injection"},
			ParamNames:   []string{"a", "GET@injection"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{ // not really useful as all after '#' should be ignored by http clients
			Query:        "a=b&URL#injection=x",
			InvalidNames: []string{"URL#injection"},
			ParamNames:   []string{"a", "URL#injection"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://host.com?%s", test.Query), nil)
			store := New(nil, req)

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Set) != 0 {
					t.Errorf("expected no GET parameters and got %d", len(store.Get))
					t.Failed()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.Failed()
			}

			for pi, pName := range test.ParamNames {
				key := fmt.Sprintf("GET@%s", pName)
				values := test.ParamValues[pi]

				isNameValid := true
				for _, invalid := range test.InvalidNames {
					if pName == invalid {
						isNameValid = false
					}
				}

				t.Run(key, func(t *testing.T) {

					param, isset := store.Set[key]
					if !isset {
						if isNameValid {
							t.Errorf("store should contain element with key '%s'", key)
							t.Failed()
						}
						return
					}

					// if should be invalid
					if isset && !isNameValid {
						t.Errorf("store should NOT contain element with key '%s' (invalid name)", key)
						t.Failed()
					}

					cast, canCast := param.Value.([]string)

					if !canCast {
						t.Errorf("should return a []string (got '%v')", cast)
						t.Failed()
					}

					if len(cast) != len(values) {
						t.Errorf("should return %d string(s) (got '%d')", len(values), len(cast))
						t.Failed()
					}

					for vi, value := range values {

						t.Run(fmt.Sprintf("value.%d", vi), func(t *testing.T) {
							if value != cast[vi] {
								t.Errorf("should return '%s' (got '%s')", value, cast[vi])
								t.Failed()
							}
						})
					}
				})

			}
		})
	}

}

func TestStoreWithUrlEncodedForm(t *testing.T) {
	tests := []struct {
		URLEncoded string

		InvalidNames []string
		ParamNames   []string
		ParamValues  [][]string
	}{
		{
			URLEncoded:   "",
			InvalidNames: []string{},
			ParamNames:   []string{},
			ParamValues:  [][]string{},
		},
		{
			URLEncoded:   "a",
			InvalidNames: []string{},
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
		},
		{
			URLEncoded:   "a&b",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{[]string{""}, []string{""}},
		},
		{
			URLEncoded:   "a=",
			InvalidNames: []string{},
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
		},
		{
			URLEncoded:   "a=&b=x",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{[]string{""}, []string{"x"}},
		},
		{
			URLEncoded:   "a=b&c=d",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b"}, []string{"d"}},
		},
		{
			URLEncoded:   "a=b&c=d&a=x",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b", "x"}, []string{"d"}},
		},
		{
			URLEncoded:   "a=b&_invalid=x",
			InvalidNames: []string{"_invalid"},
			ParamNames:   []string{"a", "_invalid"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{
			URLEncoded:   "a=b&invalid_=x",
			InvalidNames: []string{"invalid_"},
			ParamNames:   []string{"a", "invalid_"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{
			URLEncoded:   "a=b&GET@injection=x",
			InvalidNames: []string{"GET@injection"},
			ParamNames:   []string{"a", "GET@injection"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
		{
			URLEncoded:   "a=b&URL#injection=x",
			InvalidNames: []string{"URL#injection"},
			ParamNames:   []string{"a", "URL#injection"},
			ParamValues:  [][]string{[]string{"b"}, []string{""}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := bytes.NewBufferString(test.URLEncoded)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			defer req.Body.Close()
			store := New(nil, req)

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Set) != 0 {
					t.Errorf("expected no FORM parameters and got %d", len(store.Get))
					t.Failed()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.Failed()
			}

			for pi, pName := range test.ParamNames {
				key := pName
				values := test.ParamValues[pi]

				isNameValid := true
				for _, invalid := range test.InvalidNames {
					if pName == invalid {
						isNameValid = false
					}
				}

				t.Run(key, func(t *testing.T) {

					param, isset := store.Set[key]
					if !isset {
						if isNameValid {
							t.Errorf("store should contain element with key '%s'", key)
							t.Failed()
						}
						return
					}

					// if should be invalid
					if isset && !isNameValid {
						t.Errorf("store should NOT contain element with key '%s' (invalid name)", key)
						t.Failed()
					}

					cast, canCast := param.Value.([]string)

					if !canCast {
						t.Errorf("should return a []string (got '%v')", cast)
						t.Failed()
					}

					if len(cast) != len(values) {
						t.Errorf("should return %d string(s) (got '%d')", len(values), len(cast))
						t.Failed()
					}

					for vi, value := range values {

						t.Run(fmt.Sprintf("value.%d", vi), func(t *testing.T) {
							if value != cast[vi] {
								t.Errorf("should return '%s' (got '%s')", value, cast[vi])
								t.Failed()
							}
						})
					}
				})

			}
		})
	}

}

func TestJsonParameters(t *testing.T) {
	tests := []struct {
		RawJson string

		InvalidNames []string
		ParamNames   []string
		ParamValues  []interface{}
	}{
		// no need to fully check json because it is parsed with the standard library
		{
			RawJson:      "",
			InvalidNames: []string{},
			ParamNames:   []string{},
			ParamValues:  []interface{}{},
		},
		{
			RawJson:      "{}",
			InvalidNames: []string{},
			ParamNames:   []string{},
			ParamValues:  []interface{}{},
		},
		{
			RawJson:      "{ \"a\": \"b\" }",
			InvalidNames: []string{},
			ParamNames:   []string{"a"},
			ParamValues:  []interface{}{"b"},
		},
		{
			RawJson:      "{ \"a\": \"b\", \"c\": \"d\" }",
			InvalidNames: []string{},
			ParamNames:   []string{"a", "c"},
			ParamValues:  []interface{}{"b", "d"},
		},
		{
			RawJson:      "{ \"_invalid\": \"x\" }",
			InvalidNames: []string{"_invalid"},
			ParamNames:   []string{"_invalid"},
			ParamValues:  []interface{}{nil},
		},
		{
			RawJson:      "{ \"a\": \"b\", \"_invalid\": \"x\" }",
			InvalidNames: []string{"_invalid"},
			ParamNames:   []string{"a", "_invalid"},
			ParamValues:  []interface{}{"b", nil},
		},

		{
			RawJson:      "{ \"invalid_\": \"x\" }",
			InvalidNames: []string{"invalid_"},
			ParamNames:   []string{"invalid_"},
			ParamValues:  []interface{}{nil},
		},
		{
			RawJson:      "{ \"a\": \"b\", \"invalid_\": \"x\" }",
			InvalidNames: []string{"invalid_"},
			ParamNames:   []string{"a", "invalid_"},
			ParamValues:  []interface{}{"b", nil},
		},

		{
			RawJson:      "{ \"GET@injection\": \"x\" }",
			InvalidNames: []string{"GET@injection"},
			ParamNames:   []string{"GET@injection"},
			ParamValues:  []interface{}{nil},
		},
		{
			RawJson:      "{ \"a\": \"b\", \"GET@injection\": \"x\" }",
			InvalidNames: []string{"GET@injection"},
			ParamNames:   []string{"a", "GET@injection"},
			ParamValues:  []interface{}{"b", nil},
		},

		{
			RawJson:      "{ \"URL#injection\": \"x\" }",
			InvalidNames: []string{"URL#injection"},
			ParamNames:   []string{"URL#injection"},
			ParamValues:  []interface{}{nil},
		},
		{
			RawJson:      "{ \"a\": \"b\", \"URL#injection\": \"x\" }",
			InvalidNames: []string{"URL#injection"},
			ParamNames:   []string{"a", "URL#injection"},
			ParamValues:  []interface{}{"b", nil},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := bytes.NewBufferString(test.RawJson)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/json")
			defer req.Body.Close()
			store := New(nil, req)

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Set) != 0 {
					t.Errorf("expected no JSON parameters and got %d", len(store.Get))
					t.Failed()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.Failed()
			}

			for pi, pName := range test.ParamNames {
				key := pName
				value := test.ParamValues[pi]

				isNameValid := true
				for _, invalid := range test.InvalidNames {
					if pName == invalid {
						isNameValid = false
					}
				}

				t.Run(key, func(t *testing.T) {

					param, isset := store.Set[key]
					if !isset {
						if isNameValid {
							t.Errorf("store should contain element with key '%s'", key)
							t.Failed()
						}
						return
					}

					// if should be invalid
					if isset && !isNameValid {
						t.Errorf("store should NOT contain element with key '%s' (invalid name)", key)
						t.Failed()
					}

					valueType := reflect.TypeOf(value)

					paramValue := param.Value
					paramValueType := reflect.TypeOf(param.Value)

					if valueType != paramValueType {
						t.Errorf("should be of type %v (got '%v')", valueType, paramValueType)
						t.Failed()
					}

					if paramValue != value {
						t.Errorf("should return %v (got '%v')", value, paramValue)
						t.Failed()
					}

				})

			}
		})
	}

}
