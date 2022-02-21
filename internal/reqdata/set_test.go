package reqdata

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/internal/config"
)

func getEmptyService() *config.Service {
	return &config.Service{}
}

func getServiceWithURI(capturingBraces ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
	}

	index := 0

	for _, capture := range capturingBraces {
		if len(capture) == 0 {
			index++
			continue
		}

		id := fmt.Sprintf("{%s}", capture)
		service.Input[id] = &config.Parameter{
			Rename:    capture,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}

		service.Captures = append(service.Captures, &config.BraceCapture{
			Name:  capture,
			Index: index,
			Ref:   service.Input[id],
		})
		index++
	}

	return service
}
func getServiceWithQuery(t reflect.Type, params ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
		Query: make(map[string]*config.Parameter),
	}

	for _, name := range params {
		id := fmt.Sprintf("GET@%s", name)
		service.Input[id] = &config.Parameter{
			Rename:    name,
			GoType:    t,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}

		service.Query[name] = service.Input[id]
	}

	return service
}
func getServiceWithForm(t reflect.Type, params ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
		Form:  make(map[string]*config.Parameter),
	}

	for _, name := range params {
		service.Input[name] = &config.Parameter{
			Rename:    name,
			GoType:    t,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}

		service.Form[name] = service.Input[name]
	}

	return service
}

func TestStoreWithUri(t *testing.T) {
	tt := []struct {
		name          string
		serviceParams []string
		uri           string
		err           error
		errField      string
	}{
		{
			name:          "non captured uri",
			serviceParams: []string{},
			uri:           "/non-captured/uri",
			err:           nil,
		},
		{
			name:          "missing uri param",
			serviceParams: []string{"missing"},
			uri:           "/",
			err:           ErrMissingURIParameter,
			errField:      "missing",
		},
		{
			name:          "missing uri params",
			serviceParams: []string{"gotit", "missing"},
			uri:           "/gotme",
			err:           ErrMissingURIParameter,
			errField:      "missing",
		},
		{
			name:          "2 uri params",
			serviceParams: []string{"gotit", "gotittoo"},
			uri:           "/gotme/andme",
			err:           nil,
		},
		{
			name:          "2 uri params end ignored",
			serviceParams: []string{"gotit", "gotittoo"},
			uri:           "/gotme/andme/ignored",
			err:           nil,
		},
		{
			name:          "2 uri params middle ignored",
			serviceParams: []string{"first", "", "second"},
			uri:           "/gotme/ignored/gotmetoo",
			err:           nil,
		},
		{
			name:          "3 uri params last missing",
			serviceParams: []string{"first", "", "second"},
			uri:           "/gotme/ignored",
			err:           ErrMissingURIParameter,
			errField:      "second",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			service := getServiceWithURI(tc.serviceParams...)
			store := New(service)

			req := httptest.NewRequest(http.MethodGet, "http://host.com"+tc.uri, nil)
			err := store.GetURI(*req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error <%v>, got <%v>", tc.err, err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.errField {
					t.Fatalf("error field is '%s' ; expected '%s'", cast.Field(), tc.errField)
				}
				return
			}

			if len(store.Data) != len(service.Input) {
				t.Errorf("store should contain %d elements, got %d", len(service.Input), len(store.Data))
				t.Fail()
			}

		})
	}

}

func TestExtractQuery(t *testing.T) {

	tt := []struct {
		name          string
		serviceParams []string
		query         string
		err           error
		errField      string

		paramTypes  reflect.Type
		paramNames  []string
		paramValues [][]string
	}{
		{
			name:          "none required",
			serviceParams: []string{},
			query:         "",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "1 required missing",
			serviceParams: []string{"missing"},
			query:         "",
			err:           ErrMissingRequiredParam,
			errField:      "missing",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "1 required ok",
			serviceParams: []string{"a"},
			query:         "a",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "1 required 1 empty",
			serviceParams: []string{"a"},
			query:         "a&b",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "2 required 1 missing",
			serviceParams: []string{"a", "missing"},
			query:         "a&b",
			err:           ErrMissingRequiredParam,
			errField:      "missing",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "2 required 2 empty",
			serviceParams: []string{"a", "b"},
			query:         "a&b",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "b"},
			paramValues:   [][]string{{""}, {""}},
		},
		{
			name:          "1 required empty",
			serviceParams: []string{"a"},
			err:           nil,
			query:         "a=",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "2 required 1 empty",
			serviceParams: []string{"a", "b"},
			err:           nil,
			query:         "a=&b=x",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "b"},
			paramValues:   [][]string{{""}, {"x"}},
		},
		{
			name:          "2 required 2 ok",
			serviceParams: []string{"a", "c"},
			err:           nil,
			query:         "a=b&c=d",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "c"},
			paramValues:   [][]string{{"b"}, {"d"}},
		},
		{
			name:          "2 required 1 is slice",
			serviceParams: []string{"a", "c"},
			err:           ErrInvalidType,
			errField:      "a",
			query:         "a=b&c=d&a=x",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "c"},
			paramValues:   [][]string{{"b", "x"}, {"d"}},
		},

		// expect a slice
		{
			name:          "expect slice got empty",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "expect slice got value",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value"}},
		},
		{
			name:          "expect slice got values",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value1&name=value2",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value1", "value2"}},
		},

		// expect string
		{
			name:          "expect string got empty",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "expect string got value",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value"}},
		},
		{
			name:          "expect string got values",
			serviceParams: []string{"name"},
			err:           ErrInvalidType,
			errField:      "name",
			query:         "name=value1&name=value2",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			store := New(getServiceWithQuery(tc.paramTypes, tc.serviceParams...))

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://host.com?%s", tc.query), nil)
			err := store.GetQuery(*req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error <%v>, got <%v>", tc.err, err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.errField {
					t.Fatalf("error field is '%s' ; expected '%s'", cast.Field(), tc.errField)
				}
				return
			}

			if tc.paramNames == nil || tc.paramValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no GET parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(tc.paramNames) != len(tc.paramValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(tc.paramNames), len(tc.paramValues))
			}

			for pi, pName := range tc.paramNames {
				values := tc.paramValues[pi]

				t.Run(pName, func(t *testing.T) {
					param, isset := store.Data[pName]
					if !isset {
						t.Fatalf("param does not exist")
					}

					// single value
					if tc.paramTypes.Kind() != reflect.Slice {
						cast, canCast := param.(string)
						if !canCast {
							t.Fatalf("should return a string (got '%v')", cast)
						}
						if values[0] != cast {
							t.Fatalf("should return '%s' (got '%s')", values[0], cast)
						}
						return
					}

					// multiple values, should be a slice
					cast, canCast := param.([]interface{})
					if !canCast {
						t.Fatalf("should return a []string (got '%v')", cast)
					}

					if len(cast) != len(values) {
						t.Fatalf("should return %d string(s) (got '%d')", len(values), len(cast))
					}

					for vi, value := range values {
						if value != cast[vi] {
							t.Fatalf("should return '%s' (got '%s')", value, cast[vi])
						}
					}
				})

			}
		})
	}

}
func TestStoreWithUrlEncodedFormParseError(t *testing.T) {
	// http.Request.ParseForm() fails when:
	// - http.Request.Method is one of [POST,PUT,PATCH]
	// - http.Request.Form     is not nil (created manually)
	// - http.Request.PostForm is nil (deleted manually)
	// - http.Request.Body     is nil (deleted manually)

	req := httptest.NewRequest(http.MethodPost, "http://host.com/", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// break everything
	req.Body = nil
	req.Form = make(url.Values)
	req.PostForm = nil

	// defer req.Body.Close()
	store := New(nil)
	err := store.GetForm(*req)
	if err == nil {
		t.Fatalf("expected malformed urlencoded to have FailNow being parsed (got %d elements)", len(store.Data))
	}
}
func TestExtractFormUrlEncoded(t *testing.T) {
	tt := []struct {
		name          string
		serviceParams []string
		query         string
		err           error
		errField      string

		paramTypes  reflect.Type
		paramNames  []string
		paramValues [][]string
	}{
		{
			name:          "none required",
			serviceParams: []string{},
			query:         "",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "1 required missing",
			serviceParams: []string{"missing"},
			query:         "",
			err:           ErrMissingRequiredParam,
			errField:      "missing",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "1 required ok",
			serviceParams: []string{"a"},
			query:         "a",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "1 required 1 empty",
			serviceParams: []string{"a"},
			query:         "a&b",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "2 required 1 missing",
			serviceParams: []string{"a", "missing"},
			query:         "a&b",
			err:           ErrMissingRequiredParam,
			errField:      "missing",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
		{
			name:          "2 required 2 empty",
			serviceParams: []string{"a", "b"},
			query:         "a&b",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "b"},
			paramValues:   [][]string{{""}, {""}},
		},
		{
			name:          "1 required empty",
			serviceParams: []string{"a"},
			err:           nil,
			query:         "a=",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "2 required 1 empty",
			serviceParams: []string{"a", "b"},
			err:           nil,
			query:         "a=&b=x",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "b"},
			paramValues:   [][]string{{""}, {"x"}},
		},
		{
			name:          "2 required 2 ok",
			serviceParams: []string{"a", "c"},
			err:           nil,
			query:         "a=b&c=d",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "c"},
			paramValues:   [][]string{{"b"}, {"d"}},
		},
		{
			name:          "2 required 1 is slice",
			serviceParams: []string{"a", "c"},
			err:           ErrInvalidType,
			errField:      "a",
			query:         "a=b&c=d&a=x",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "c"},
			paramValues:   [][]string{{"b", "x"}, {"d"}},
		},

		// expect a slice
		{
			name:          "expect slice got empty",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "expect slice got value",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value"}},
		},
		{
			name:          "expect slice got values",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value1&name=value2",
			paramTypes:    reflect.TypeOf([]string{}),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value1", "value2"}},
		},

		// expect string
		{
			name:          "expect string got empty",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{""}},
		},
		{
			name:          "expect string got value",
			serviceParams: []string{"name"},
			err:           nil,
			query:         "name=value",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"name"},
			paramValues:   [][]string{{"value"}},
		},
		{
			name:          "expect string got values",
			serviceParams: []string{"name"},
			err:           ErrInvalidType,
			errField:      "name",
			query:         "name=value1&name=value2",
			paramTypes:    reflect.TypeOf(""),
			paramNames:    nil,
			paramValues:   nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.query)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			defer req.Body.Close()

			store := New(getServiceWithForm(tc.paramTypes, tc.serviceParams...))
			err := store.GetForm(*req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error <%v>, got <%v>", tc.err, err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.errField {
					t.Fatalf("error field is '%s' ; expected '%s'", cast.Field(), tc.errField)
				}
				return
			}

			if tc.paramNames == nil || tc.paramValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no GET parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(tc.paramNames) != len(tc.paramValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(tc.paramNames), len(tc.paramValues))
			}

			for pi, key := range tc.paramNames {
				values := tc.paramValues[pi]

				t.Run(key, func(t *testing.T) {
					param, isset := store.Data[key]
					if !isset {
						t.Fatalf("param does not exist")
					}

					// single value
					if tc.paramTypes.Kind() != reflect.Slice {
						cast, canCast := param.(string)
						if !canCast {
							t.Fatalf("should return a string (got '%v')", cast)
						}
						if values[0] != cast {
							t.Fatalf("should return '%s' (got '%s')", values[0], cast)
						}
						return
					}

					// multiple values, should be a slice
					cast, canCast := param.([]interface{})
					if !canCast {
						t.Fatalf("should return a []string (got '%v')", cast)
					}

					if len(cast) != len(values) {
						t.Fatalf("should return %d string(s) (got '%d')", len(values), len(cast))
					}

					for vi, value := range values {
						if value != cast[vi] {
							t.Fatalf("should return '%s' (got '%s')", value, cast[vi])
						}
					}
				})

			}
		})
	}

}

func TestJsonParameters(t *testing.T) {
	tt := []struct {
		name          string
		serviceParams []string
		raw           string
		err           error

		paramTypes  reflect.Type
		paramNames  []string
		paramValues []interface{}
	}{
		// no need to fully check json because it is parsed with the standard library
		{
			name:          "empty body",
			serviceParams: []string{},
			raw:           "",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{},
			paramValues:   []interface{}{},
		},
		{
			name:          "empty json",
			serviceParams: []string{},
			raw:           "{}",
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{},
			paramValues:   []interface{}{},
		},
		{
			name:          "1 provided none expected",
			serviceParams: []string{},
			raw:           `{ "a": "b" }`,
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{},
			paramValues:   []interface{}{},
		},
		{
			name:          "1 required ok",
			serviceParams: []string{"a"},
			raw:           `{ "a": "b" }`,
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   []interface{}{"b"},
		},
		{
			name:          "1 required 1 ignored",
			serviceParams: []string{"a"},
			raw:           `{ "a": "b", "ignored": "d" }`,
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   []interface{}{"b"},
		},
		{
			name:          "2 required 2 ok",
			serviceParams: []string{"a", "c"},
			raw:           `{ "a": "b", "c": "d" }`,
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a", "c"},
			paramValues:   []interface{}{"b", "d"},
		},
		{
			name:          "1 required null",
			serviceParams: []string{"a"},
			raw:           `{ "a": null }`,
			err:           nil,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{"a"},
			paramValues:   []interface{}{nil},
		},
		// json parse error
		{
			name:          "invalid json",
			serviceParams: []string{},
			raw:           `{ "a": "b", }`,
			err:           ErrInvalidJSON,
			paramTypes:    reflect.TypeOf(""),
			paramNames:    []string{},
			paramValues:   []interface{}{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.raw)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/json")
			defer req.Body.Close()
			store := New(getServiceWithForm(tc.paramTypes, tc.serviceParams...))

			err := store.GetForm(*req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error <%v>, got <%v>", tc.err, err)
			}
			if err != nil {
				return
			}

			if tc.paramNames == nil || tc.paramValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no JSON parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(tc.paramNames) != len(tc.paramValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(tc.paramNames), len(tc.paramValues))
			}

			for pi, pName := range tc.paramNames {
				key := pName
				value := tc.paramValues[pi]

				t.Run(key, func(t *testing.T) {

					param, isset := store.Data[key]
					if !isset {
						t.Fatalf("store should contain element with key '%s'", key)
						return
					}

					valueType := reflect.TypeOf(value)

					paramValue := param
					paramValueType := reflect.TypeOf(param)

					if valueType != paramValueType {
						t.Fatalf("should be of type %v (got '%v')", valueType, paramValueType)
					}

					if paramValue != value {
						t.Fatalf("should return %v (got '%v')", value, paramValue)
					}

				})

			}
		})
	}

}

func TestMultipartParameters(t *testing.T) {
	fileContent := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	tt := []struct {
		serviceParams []string
		rawMultipart  string
		err           error

		// we must receive []byte when we send a file. In the opposite, we wait
		// for a string when we send a literal parameter
		lastParamBytes bool

		paramNames  []string
		paramValues []interface{}
	}{
		{
			serviceParams: []string{},
			rawMultipart:  ``,
			err:           nil,
			paramNames:    []string{},
			paramValues:   []interface{}{},
		},
		{
			serviceParams: []string{},
			rawMultipart: `--xxx
			`,
			err:         ErrInvalidMultipart,
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			serviceParams: []string{},
			rawMultipart: `--xxx
--xxx--`,
			err:         ErrInvalidMultipart,
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			serviceParams: []string{},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			serviceParams: []string{"a"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"b"},
		},
		{
			serviceParams: []string{"a", "c"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="c"

d
--xxx--`,
			err:         nil,
			paramNames:  []string{"a", "c"},
			paramValues: []interface{}{"b", "d"},
		},
		{
			serviceParams: []string{"a"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="ignored"

x
--xxx--`,
			err:         nil,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"b"},
		},
		{
			serviceParams:  []string{"a", "file"},
			lastParamBytes: true,
			rawMultipart: fmt.Sprintf(`--xxx
Content-Disposition: form-data; name="a"

a-content
--xxx
Content-Disposition: form-data; name="file"
Content-Type: application/zip

%s
--xxx--`, fileContent),
			err:         nil,
			paramNames:  []string{"a", "file"},
			paramValues: []interface{}{"a-content", fileContent},
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := strings.NewReader(tc.rawMultipart)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "multipart/form-data; boundary=xxx")
			defer req.Body.Close()

			service := &config.Service{
				Input: make(map[string]*config.Parameter),
				Form:  make(map[string]*config.Parameter),
			}

			for i, name := range tc.serviceParams {
				isLast := (i == len(tc.serviceParams)-1)

				gotype := reflect.TypeOf("")
				if isLast && tc.lastParamBytes {
					gotype = reflect.TypeOf([]byte{})
				}
				service.Input[name] = &config.Parameter{
					Rename:    name,
					GoType:    gotype,
					Validator: func(value interface{}) (interface{}, bool) { return value, true },
				}
				service.Form[name] = service.Input[name]
			}

			store := New(getServiceWithForm(reflect.TypeOf(""), tc.serviceParams...))

			err := store.GetForm(*req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error <%v>, got <%v>", tc.err, err)
			}
			if err != nil {
				return
			}

			if tc.paramNames == nil || tc.paramValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no JSON parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(tc.paramNames) != len(tc.paramValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(tc.paramNames), len(tc.paramValues))
			}

			for pi, key := range tc.paramNames {
				isLast := (pi == len(tc.paramNames)-1)
				expect := tc.paramValues[pi]

				t.Run(key, func(t *testing.T) {
					param, exists := store.Data[key]
					if !exists {
						t.Fatalf("store should contain element with key '%s'", key)
						return
					}

					// expect a []byte
					if isLast && tc.lastParamBytes {
						expectVal, ok := expect.([]byte)
						if !ok {
							t.Fatalf("expected value is not a []byte")
						}
						actualVal, ok := param.([]byte)
						if !ok {
							t.Fatalf("actual value is not a []byte")
						}

						if bytes.Compare(actualVal, expectVal) != 0 {
							t.Fatalf("invalid bytes\nactual: %v\nexpect: %v", actualVal, expectVal)
						}
						return
					}

					// expect a string
					expectVal, ok := expect.(string)
					if !ok {
						t.Fatalf("expected value is not a string")
					}
					actualVal, ok := param.(string)
					if !ok {
						t.Fatalf("actual value is not a string")
					}

					if actualVal != expectVal {
						t.Fatalf("invalid string\nactual: %s\nexpect: %s", actualVal, expectVal)
					}

				})

			}
		})
	}

}
