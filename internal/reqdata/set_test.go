package reqdata

import (
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
func getServiceWithQuery(params ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
		Query: make(map[string]*config.Parameter),
	}

	for _, name := range params {
		id := fmt.Sprintf("GET@%s", name)
		service.Input[id] = &config.Parameter{
			Rename:    name,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}

		service.Query[name] = service.Input[id]
	}

	return service
}
func getServiceWithForm(params ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
		Form:  make(map[string]*config.Parameter),
	}

	for _, name := range params {
		service.Input[name] = &config.Parameter{
			Rename:    name,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}

		service.Form[name] = service.Input[name]
	}

	return service
}

func TestStoreWithUri(t *testing.T) {
	tests := []struct {
		ServiceParams []string
		URI           string
		Err           error
	}{
		{
			[]string{},
			"/non-captured/uri",
			nil,
		},
		{
			[]string{"missing"},
			"/",
			ErrMissingURIParameter,
		},
		{
			[]string{"gotit", "missing"},
			"/gotme",
			ErrMissingURIParameter,
		},
		{
			[]string{"gotit", "gotittoo"},
			"/gotme/andme",
			nil,
		},
		{
			[]string{"gotit", "gotittoo"},
			"/gotme/andme/ignored",
			nil,
		},
		{
			[]string{"first", "", "second"},
			"/gotme/ignored/gotmetoo",
			nil,
		},
		{
			[]string{"first", "", "second"},
			"/gotme/ignored",
			ErrMissingURIParameter,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test.%d", i), func(t *testing.T) {
			service := getServiceWithURI(test.ServiceParams...)
			store := New(service)

			req := httptest.NewRequest(http.MethodGet, "http://host.com"+test.URI, nil)
			err := store.GetURI(*req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Fatalf("expected error <%s>, got <%s>", test.Err, err)
					}
					return
				}
				t.Fatalf("unexpected error <%s>", err)
			}

			if len(store.Data) != len(service.Input) {
				t.Errorf("store should contain %d elements, got %d", len(service.Input), len(store.Data))
				t.Fail()
			}

		})
	}

}

func TestExtractQuery(t *testing.T) {

	tests := []struct {
		ServiceParam []string
		Query        string
		Err          error

		ParamNames  []string
		ParamValues [][]string
	}{
		{
			ServiceParam: []string{},
			Query:        "",
			Err:          nil,
			ParamNames:   nil,
			ParamValues:  nil,
		},
		{
			ServiceParam: []string{"missing"},
			Query:        "",
			Err:          ErrMissingRequiredParam,
			ParamNames:   nil,
			ParamValues:  nil,
		},
		{
			ServiceParam: []string{"a"},
			Query:        "a",
			Err:          nil,
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{{""}},
		},
		{
			ServiceParam: []string{"a"},
			Query:        "a&b",
			Err:          nil,
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{{""}},
		},
		{
			ServiceParam: []string{"a", "missing"},
			Query:        "a&b",
			Err:          ErrMissingRequiredParam,
			ParamNames:   nil,
			ParamValues:  nil,
		},
		{
			ServiceParam: []string{"a", "b"},
			Query:        "a&b",
			Err:          nil,
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{{""}, {""}},
		},
		{
			ServiceParam: []string{"a"},
			Err:          nil,
			Query:        "a=",
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{{""}},
		},
		{
			ServiceParam: []string{"a", "b"},
			Err:          nil,
			Query:        "a=&b=x",
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{{""}, {"x"}},
		},
		{
			ServiceParam: []string{"a", "c"},
			Err:          nil,
			Query:        "a=b&c=d",
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{{"b"}, {"d"}},
		},
		{
			ServiceParam: []string{"a", "c"},
			Err:          nil,
			Query:        "a=b&c=d&a=x",
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{{"b", "x"}, {"d"}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request[%d]", i), func(t *testing.T) {

			store := New(getServiceWithQuery(test.ServiceParam...))

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://host.com?%s", test.Query), nil)
			err := store.GetQuery(*req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Fatalf("expected error <%s>, got <%s>", test.Err, err)
					}
					return
				}
				t.Fatalf("unexpected error <%s>", err)
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no GET parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
			}

			for pi, pName := range test.ParamNames {
				values := test.ParamValues[pi]

				t.Run(pName, func(t *testing.T) {
					param, isset := store.Data[pName]
					if !isset {
						t.Fatalf("param does not exist")
					}

					// single value, should return a single element
					if len(values) == 1 {
						cast, canCast := param.(string)
						if !canCast {
							t.Fatalf("should return a string (got '%v')", cast)
						}
						if values[0] != cast {
							t.Fatalf("should return '%s' (got '%s')", values[0], cast)
						}
						return
					}

					// multiple values, should return a slice
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
	tests := []struct {
		ServiceParams []string
		URLEncoded    string
		Err           error

		ParamNames  []string
		ParamValues [][]string
	}{
		{
			ServiceParams: []string{},
			URLEncoded:    "",
			Err:           nil,
			ParamNames:    nil,
			ParamValues:   nil,
		},
		{
			ServiceParams: []string{"missing"},
			URLEncoded:    "",
			Err:           ErrMissingRequiredParam,
			ParamNames:    nil,
			ParamValues:   nil,
		},
		{
			ServiceParams: []string{"a"},
			URLEncoded:    "a",
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   [][]string{{""}},
		},
		{
			ServiceParams: []string{"a"},
			URLEncoded:    "a&b",
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   [][]string{{""}},
		},
		{
			ServiceParams: []string{"a", "missing"},
			URLEncoded:    "a&b",
			Err:           ErrMissingRequiredParam,
			ParamNames:    nil,
			ParamValues:   nil,
		},
		{
			ServiceParams: []string{"a", "b"},
			URLEncoded:    "a&b",
			Err:           nil,
			ParamNames:    []string{"a", "b"},
			ParamValues:   [][]string{{""}, {""}},
		},
		{
			ServiceParams: []string{"a"},
			Err:           nil,
			URLEncoded:    "a=",
			ParamNames:    []string{"a"},
			ParamValues:   [][]string{{""}},
		},
		{
			ServiceParams: []string{"a", "b"},
			Err:           nil,
			URLEncoded:    "a=&b=x",
			ParamNames:    []string{"a", "b"},
			ParamValues:   [][]string{{""}, {"x"}},
		},
		{
			ServiceParams: []string{"a", "c"},
			Err:           nil,
			URLEncoded:    "a=b&c=d",
			ParamNames:    []string{"a", "c"},
			ParamValues:   [][]string{{"b"}, {"d"}},
		},
		{
			ServiceParams: []string{"a", "c"},
			Err:           nil,
			URLEncoded:    "a=b&c=d&a=x",
			ParamNames:    []string{"a", "c"},
			ParamValues:   [][]string{{"b", "x"}, {"d"}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := strings.NewReader(test.URLEncoded)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			defer req.Body.Close()

			store := New(getServiceWithForm(test.ServiceParams...))
			err := store.GetForm(*req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Fatalf("expected error <%s>, got <%s>", test.Err, err)
					}
					return
				}
				t.Fatalf("unexpected error <%s>", err)
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no GET parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
			}

			for pi, key := range test.ParamNames {
				values := test.ParamValues[pi]

				t.Run(key, func(t *testing.T) {
					param, isset := store.Data[key]
					if !isset {
						t.Fatalf("param does not exist")
					}

					// single value, should return a single element
					if len(values) == 1 {
						cast, canCast := param.(string)
						if !canCast {
							t.Fatalf("should return a string (got '%v')", cast)
						}
						if values[0] != cast {
							t.Fatalf("should return '%s' (got '%s')", values[0], cast)
						}
						return
					}

					// multiple values, should return a slice
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
	tests := []struct {
		ServiceParams []string
		Raw           string
		Err           error

		ParamNames  []string
		ParamValues []interface{}
	}{
		// no need to fully check json because it is parsed with the standard library
		{
			ServiceParams: []string{},
			Raw:           "",
			Err:           nil,
			ParamNames:    []string{},
			ParamValues:   []interface{}{},
		},
		{
			ServiceParams: []string{},
			Raw:           "{}",
			Err:           nil,
			ParamNames:    []string{},
			ParamValues:   []interface{}{},
		},
		{
			ServiceParams: []string{},
			Raw:           `{ "a": "b" }`,
			Err:           nil,
			ParamNames:    []string{},
			ParamValues:   []interface{}{},
		},
		{
			ServiceParams: []string{"a"},
			Raw:           `{ "a": "b" }`,
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   []interface{}{"b"},
		},
		{
			ServiceParams: []string{"a"},
			Raw:           `{ "a": "b", "ignored": "d" }`,
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   []interface{}{"b"},
		},
		{
			ServiceParams: []string{"a", "c"},
			Raw:           `{ "a": "b", "c": "d" }`,
			Err:           nil,
			ParamNames:    []string{"a", "c"},
			ParamValues:   []interface{}{"b", "d"},
		},
		{
			ServiceParams: []string{"a"},
			Raw:           `{ "a": null }`,
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   []interface{}{nil},
		},
		// json parse error
		{
			ServiceParams: []string{},
			Raw:           `{ "a": "b", }`,
			Err:           ErrInvalidJSON,
			ParamNames:    []string{},
			ParamValues:   []interface{}{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := strings.NewReader(test.Raw)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/json")
			defer req.Body.Close()
			store := New(getServiceWithForm(test.ServiceParams...))

			err := store.GetForm(*req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Fatalf("expected error <%s>, got <%s>", test.Err, err)
					}
					return
				}
				t.Fatalf("unexpected error <%s>", err)
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no JSON parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
			}

			for pi, pName := range test.ParamNames {
				key := pName
				value := test.ParamValues[pi]

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
	tests := []struct {
		ServiceParams []string
		RawMultipart  string
		Err           error

		ParamNames  []string
		ParamValues []interface{}
	}{
		// no need to fully check json because it is parsed with the standard library
		{
			ServiceParams: []string{},
			RawMultipart:  ``,
			Err:           nil,
			ParamNames:    []string{},
			ParamValues:   []interface{}{},
		},
		{
			ServiceParams: []string{},
			RawMultipart: `--xxx
			`,
			Err:         ErrInvalidMultipart,
			ParamNames:  []string{},
			ParamValues: []interface{}{},
		},
		{
			ServiceParams: []string{},
			RawMultipart: `--xxx
--xxx--`,
			Err:         ErrInvalidMultipart,
			ParamNames:  []string{},
			ParamValues: []interface{}{},
		},
		{
			ServiceParams: []string{},
			RawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			ParamNames:  []string{},
			ParamValues: []interface{}{},
		},
		{
			ServiceParams: []string{"a"},
			RawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			ParamNames:  []string{"a"},
			ParamValues: []interface{}{"b"},
		},
		{
			ServiceParams: []string{"a", "c"},
			RawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="c"

d
--xxx--`,
			Err:         nil,
			ParamNames:  []string{"a", "c"},
			ParamValues: []interface{}{"b", "d"},
		},
		{
			ServiceParams: []string{"a"},
			RawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="ignored"

x
--xxx--`,
			Err:         nil,
			ParamNames:  []string{"a"},
			ParamValues: []interface{}{"b"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := strings.NewReader(test.RawMultipart)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "multipart/form-data; boundary=xxx")
			defer req.Body.Close()
			store := New(getServiceWithForm(test.ServiceParams...))

			err := store.GetForm(*req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Fatalf("expected error <%s>, got <%s>", test.Err, err)
					}
					return
				}
				t.Fatalf("unexpected error <%s>", err)
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Fatalf("expected no JSON parameters and got %d", len(store.Data))
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Fatalf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
			}

			for pi, key := range test.ParamNames {
				value := test.ParamValues[pi]

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
