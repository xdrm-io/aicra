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

	"git.xdrm.io/go/aicra/internal/config"
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
			err := store.ExtractURI(req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Errorf("expected error <%s>, got <%s>", test.Err, err)
						t.FailNow()
					}
					return
				}
				t.Errorf("unexpected error <%s>", err)
				t.FailNow()
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
			ParamValues:  [][]string{[]string{""}},
		},
		{
			ServiceParam: []string{"a"},
			Query:        "a&b",
			Err:          nil,
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
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
			ParamValues:  [][]string{[]string{""}, []string{""}},
		},
		{
			ServiceParam: []string{"a"},
			Err:          nil,
			Query:        "a=",
			ParamNames:   []string{"a"},
			ParamValues:  [][]string{[]string{""}},
		},
		{
			ServiceParam: []string{"a", "b"},
			Err:          nil,
			Query:        "a=&b=x",
			ParamNames:   []string{"a", "b"},
			ParamValues:  [][]string{[]string{""}, []string{"x"}},
		},
		{
			ServiceParam: []string{"a", "c"},
			Err:          nil,
			Query:        "a=b&c=d",
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b"}, []string{"d"}},
		},
		{
			ServiceParam: []string{"a", "c"},
			Err:          nil,
			Query:        "a=b&c=d&a=x",
			ParamNames:   []string{"a", "c"},
			ParamValues:  [][]string{[]string{"b", "x"}, []string{"d"}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {

			store := New(getServiceWithQuery(test.ServiceParam...))

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://host.com?%s", test.Query), nil)
			err := store.ExtractQuery(req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Errorf("expected error <%s>, got <%s>", test.Err, err)
						t.FailNow()
					}
					return
				}
				t.Errorf("unexpected error <%s>", err)
				t.FailNow()
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Errorf("expected no GET parameters and got %d", len(store.Data))
					t.FailNow()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.FailNow()
			}

			for pi, pName := range test.ParamNames {
				values := test.ParamValues[pi]

				t.Run(pName, func(t *testing.T) {
					param, isset := store.Data[pName]
					if !isset {
						t.Errorf("param does not exist")
						t.FailNow()
					}

					cast, canCast := param.Value.([]string)
					if !canCast {
						t.Errorf("should return a []string (got '%v')", cast)
						t.FailNow()
					}

					if len(cast) != len(values) {
						t.Errorf("should return %d string(s) (got '%d')", len(values), len(cast))
						t.FailNow()
					}

					for vi, value := range values {

						t.Run(fmt.Sprintf("value.%d", vi), func(t *testing.T) {
							if value != cast[vi] {
								t.Errorf("should return '%s' (got '%s')", value, cast[vi])
								t.FailNow()
							}
						})
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
	err := store.ExtractForm(req)
	if err == nil {
		t.Errorf("expected malformed urlencoded to have FailNow being parsed (got %d elements)", len(store.Data))
		t.FailNow()

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
			ParamValues:   [][]string{[]string{""}},
		},
		{
			ServiceParams: []string{"a"},
			URLEncoded:    "a&b",
			Err:           nil,
			ParamNames:    []string{"a"},
			ParamValues:   [][]string{[]string{""}},
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
			ParamValues:   [][]string{[]string{""}, []string{""}},
		},
		{
			ServiceParams: []string{"a"},
			Err:           nil,
			URLEncoded:    "a=",
			ParamNames:    []string{"a"},
			ParamValues:   [][]string{[]string{""}},
		},
		{
			ServiceParams: []string{"a", "b"},
			Err:           nil,
			URLEncoded:    "a=&b=x",
			ParamNames:    []string{"a", "b"},
			ParamValues:   [][]string{[]string{""}, []string{"x"}},
		},
		{
			ServiceParams: []string{"a", "c"},
			Err:           nil,
			URLEncoded:    "a=b&c=d",
			ParamNames:    []string{"a", "c"},
			ParamValues:   [][]string{[]string{"b"}, []string{"d"}},
		},
		{
			ServiceParams: []string{"a", "c"},
			Err:           nil,
			URLEncoded:    "a=b&c=d&a=x",
			ParamNames:    []string{"a", "c"},
			ParamValues:   [][]string{[]string{"b", "x"}, []string{"d"}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("request.%d", i), func(t *testing.T) {
			body := strings.NewReader(test.URLEncoded)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			defer req.Body.Close()

			store := New(getServiceWithForm(test.ServiceParams...))
			err := store.ExtractForm(req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Errorf("expected error <%s>, got <%s>", test.Err, err)
						t.FailNow()
					}
					return
				}
				t.Errorf("unexpected error <%s>", err)
				t.FailNow()
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Errorf("expected no GET parameters and got %d", len(store.Data))
					t.FailNow()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.FailNow()
			}

			for pi, key := range test.ParamNames {
				values := test.ParamValues[pi]

				t.Run(key, func(t *testing.T) {
					param, isset := store.Data[key]
					if !isset {
						t.Errorf("param does not exist")
						t.FailNow()
					}

					cast, canCast := param.Value.([]string)
					if !canCast {
						t.Errorf("should return a []string (got '%v')", cast)
						t.FailNow()
					}

					if len(cast) != len(values) {
						t.Errorf("should return %d string(s) (got '%d')", len(values), len(cast))
						t.FailNow()
					}

					for vi, value := range values {

						t.Run(fmt.Sprintf("value.%d", vi), func(t *testing.T) {
							if value != cast[vi] {
								t.Errorf("should return '%s' (got '%s')", value, cast[vi])
								t.FailNow()
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

			err := store.ExtractForm(req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Errorf("expected error <%s>, got <%s>", test.Err, err)
						t.FailNow()
					}
					return
				}
				t.Errorf("unexpected error <%s>", err)
				t.FailNow()
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Errorf("expected no JSON parameters and got %d", len(store.Data))
					t.FailNow()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.FailNow()
			}

			for pi, pName := range test.ParamNames {
				key := pName
				value := test.ParamValues[pi]

				t.Run(key, func(t *testing.T) {

					param, isset := store.Data[key]
					if !isset {
						t.Errorf("store should contain element with key '%s'", key)
						t.FailNow()
						return
					}

					valueType := reflect.TypeOf(value)

					paramValue := param.Value
					paramValueType := reflect.TypeOf(param.Value)

					if valueType != paramValueType {
						t.Errorf("should be of type %v (got '%v')", valueType, paramValueType)
						t.FailNow()
					}

					if paramValue != value {
						t.Errorf("should return %v (got '%v')", value, paramValue)
						t.FailNow()
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

			err := store.ExtractForm(req)
			if err != nil {
				if test.Err != nil {
					if !errors.Is(err, test.Err) {
						t.Errorf("expected error <%s>, got <%s>", test.Err, err)
						t.FailNow()
					}
					return
				}
				t.Errorf("unexpected error <%s>", err)
				t.FailNow()
			}

			if test.ParamNames == nil || test.ParamValues == nil {
				if len(store.Data) != 0 {
					t.Errorf("expected no JSON parameters and got %d", len(store.Data))
					t.FailNow()
				}

				// no param to check
				return
			}

			if len(test.ParamNames) != len(test.ParamValues) {
				t.Errorf("invalid test: names and values differ in size (%d vs %d)", len(test.ParamNames), len(test.ParamValues))
				t.FailNow()
			}

			for pi, key := range test.ParamNames {
				value := test.ParamValues[pi]

				t.Run(key, func(t *testing.T) {

					param, isset := store.Data[key]
					if !isset {
						t.Errorf("store should contain element with key '%s'", key)
						t.FailNow()
						return
					}

					valueType := reflect.TypeOf(value)

					paramValue := param.Value
					paramValueType := reflect.TypeOf(param.Value)

					if valueType != paramValueType {
						t.Errorf("should be of type %v (got '%v')", valueType, paramValueType)
						t.FailNow()
					}

					if paramValue != value {
						t.Errorf("should return %v (got '%v')", value, paramValue)
						t.FailNow()
					}

				})

			}
		})
	}

}
