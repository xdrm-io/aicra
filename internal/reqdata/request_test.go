package reqdata

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xdrm-io/aicra/internal/config"
)

func getEmptyService() *config.Service {
	return &config.Service{}
}

func getServiceWithURI(capturingBraces ...string) *config.Service {
	service := &config.Service{
		Input: make(map[string]*config.Parameter),
	}

	for i, capture := range capturingBraces {
		if len(capture) == 0 {
			continue
		}

		id := fmt.Sprintf("{%s}", capture)
		service.Input[id] = &config.Parameter{
			Rename:    capture,
			Validator: func(value interface{}) (interface{}, bool) { return value, true },
		}
		service.Captures = append(service.Captures, &config.BraceCapture{
			Name:  capture,
			Index: i,
			Ref:   service.Input[id],
		})
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

func TestRequestPoolReUse(t *testing.T) {
	// replace the map pool with our custom code
	var oldPool = mapPool
	defer func() { mapPool = oldPool }()

	var i uint32
	mapPool = sync.Pool{
		New: func() interface{} {
			atomic.AddUint32(&i, 1)
			return make(map[string]interface{}, 8)
		},
	}

	req1 := NewRequest(&config.Service{})
	if atomic.AddUint32(&i, 0) != 1 {
		t.Fatalf("NewRequest shall have allocated 1 map")
	}
	req1.Data["test"] = "value"
	req1.Release()

	time.Sleep(1 * time.Microsecond)

	req2 := NewRequest(&config.Service{})
	if atomic.AddUint32(&i, 0) != 1 {
		t.Fatalf("NewRequest shall have reused the previously allocated map")
	}
	if _, exists := req2.Data["test"]; exists {
		t.Fatalf("NewRequest shall have reset the reused map")
	}
	req2.Release()
}

func TestRequestWithUri(t *testing.T) {
	tt := []struct {
		name   string
		params []string
		uri    string
		err    error
		field  string
	}{
		{
			name:   "non captured uri",
			params: []string{},
			uri:    "/non-captured/uri",
			err:    nil,
		},
		{
			name:   "missing uri param",
			params: []string{"missing"},
			uri:    "/",
			err:    ErrMissingURIParameter,
			field:  "missing",
		},
		{
			name:   "missing uri params",
			params: []string{"gotit", "missing"},
			uri:    "/gotme",
			err:    ErrMissingURIParameter,
			field:  "missing",
		},
		{
			name:   "2 uri params",
			params: []string{"gotit", "gotittoo"},
			uri:    "/gotme/andme",
			err:    nil,
		},
		{
			name:   "2 uri params end ignored",
			params: []string{"gotit", "gotittoo"},
			uri:    "/gotme/andme/ignored",
			err:    nil,
		},
		{
			name:   "2 uri params middle ignored",
			params: []string{"first", "", "second"},
			uri:    "/gotme/ignored/gotmetoo",
			err:    nil,
		},
		{
			name:   "3 uri params last missing",
			params: []string{"first", "", "second"},
			uri:    "/gotme/ignored",
			err:    ErrMissingURIParameter,
			field:  "second",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				service = getServiceWithURI(tc.params...)
				req     = httptest.NewRequest(http.MethodGet, "http://host.com"+tc.uri, nil)
				store   = NewRequest(service)
			)

			err := store.ExtractURI(req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.field {
					t.Fatalf("invalid field\nactual: %q\nexpect: %q", cast.Field(), tc.field)
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

// checkExtracted checks extracted parameters against:
// - names: expected parameter names
// - kind: the common type kind for all parameters
// - expected: values as a string list for each value
func checkExtracted(t *testing.T, names []string, kind reflect.Kind, expected [][]string, extracted map[string]interface{}) {
	for n, name := range names {
		var (
			expectList   = expected[n]
			param, isset = extracted[name]
		)
		if !isset {
			t.Fatalf("param does not exist")
		}

		// single value
		if kind != reflect.Slice {
			cast, canCast := param.(string)
			if !canCast {
				t.Fatalf("invalid type\nactual: %T\nexpect: string", param)
			}
			if cast != expectList[0] {
				t.Fatalf("invalid value\nactual: %q\nexpect: %q", cast, expectList[0])
			}
			return
		}

		// multiple values, should be a slice
		cast, canCast := param.([]string)
		if !canCast {
			t.Fatalf("invalid type\nactual: %T\nexpect: []string", param)
		}
		// compare every string
		if len(cast) != len(expectList) {
			t.Fatalf("invalid size\nactual: %d\nexpect: %d", len(cast), len(expected))
		}
		for e, expect := range expectList {
			if expect != cast[e] {
				t.Fatalf("invalid value\nactual: %s\nexpect: %s", cast[e], expect)
			}
		}
	}
}

func TestExtractQuery(t *testing.T) {
	tt := []struct {
		name   string
		params []string
		query  string
		err    error
		field  string

		paramTypes  reflect.Type
		paramNames  []string
		paramValues [][]string
	}{
		{
			name:        "none required",
			params:      []string{},
			query:       "",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "1 required missing",
			params:      []string{"missing"},
			query:       "",
			err:         ErrMissingRequiredParam,
			field:       "missing",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "1 required ok",
			params:      []string{"a"},
			query:       "a",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "1 required 1 empty",
			params:      []string{"a"},
			query:       "a&b",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "2 required 1 missing",
			params:      []string{"a", "missing"},
			query:       "a&b",
			err:         ErrMissingRequiredParam,
			field:       "missing",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "2 required 2 empty",
			params:      []string{"a", "b"},
			query:       "a&b",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "b"},
			paramValues: [][]string{{""}, {""}},
		},
		{
			name:        "1 required empty",
			params:      []string{"a"},
			err:         nil,
			query:       "a=",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "2 required 1 empty",
			params:      []string{"a", "b"},
			err:         nil,
			query:       "a=&b=x",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "b"},
			paramValues: [][]string{{""}, {"x"}},
		},
		{
			name:        "2 required 2 ok",
			params:      []string{"a", "c"},
			err:         nil,
			query:       "a=b&c=d",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "c"},
			paramValues: [][]string{{"b"}, {"d"}},
		},
		{
			name:        "2 required 1 is slice",
			params:      []string{"a", "c"},
			err:         ErrInvalidType,
			field:       "a",
			query:       "a=b&c=d&a=x",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "c"},
			paramValues: [][]string{{"b", "x"}, {"d"}},
		},

		// expect a slice
		{
			name:        "expect slice got empty",
			params:      []string{"name"},
			err:         nil,
			query:       "name=",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "expect slice got value",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value"}},
		},
		{
			name:        "expect slice got values",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value1&name=value2",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value1", "value2"}},
		},

		// expect string
		{
			name:        "expect string got empty",
			params:      []string{"name"},
			err:         nil,
			query:       "name=",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"name"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "expect string got value",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value"}},
		},
		{
			name:        "expect string got values",
			params:      []string{"name"},
			err:         ErrInvalidType,
			field:       "name",
			query:       "name=value1&name=value2",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				req   = httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://host.com?%s", tc.query), nil)
				store = NewRequest(getServiceWithQuery(tc.paramTypes, tc.params...))
				err   = store.ExtractQuery(req)
			)

			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.field {
					t.Fatalf("invalid field\nactual: %v\nexpect: %v", cast.Field(), tc.field)
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

			checkExtracted(t, tc.paramNames, tc.paramTypes.Kind(), tc.paramValues, store.Data)
		})
	}

}

func TestExtractFormUrlEncoded(t *testing.T) {
	tt := []struct {
		name   string
		params []string
		query  string
		err    error
		field  string

		paramTypes  reflect.Type
		paramNames  []string
		paramValues [][]string
	}{
		{
			name:        "none required",
			params:      []string{},
			query:       "",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "1 required missing",
			params:      []string{"missing"},
			query:       "",
			err:         ErrMissingRequiredParam,
			field:       "missing",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "1 required ok",
			params:      []string{"a"},
			query:       "a",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "1 required 1 empty",
			params:      []string{"a"},
			query:       "a&b",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "2 required 1 missing",
			params:      []string{"a", "missing"},
			query:       "a&b",
			err:         ErrMissingRequiredParam,
			field:       "missing",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
		{
			name:        "2 required 2 empty",
			params:      []string{"a", "b"},
			query:       "a&b",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "b"},
			paramValues: [][]string{{""}, {""}},
		},
		{
			name:        "1 required empty",
			params:      []string{"a"},
			err:         nil,
			query:       "a=",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "2 required 1 empty",
			params:      []string{"a", "b"},
			err:         nil,
			query:       "a=&b=x",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "b"},
			paramValues: [][]string{{""}, {"x"}},
		},
		{
			name:        "2 required 2 ok",
			params:      []string{"a", "c"},
			err:         nil,
			query:       "a=b&c=d",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "c"},
			paramValues: [][]string{{"b"}, {"d"}},
		},
		{
			name:        "2 required 1 is slice",
			params:      []string{"a", "c"},
			err:         ErrInvalidType,
			field:       "a",
			query:       "a=b&c=d&a=x",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "c"},
			paramValues: [][]string{{"b", "x"}, {"d"}},
		},

		// expect a slice
		{
			name:        "expect slice got empty",
			params:      []string{"name"},
			err:         nil,
			query:       "name=",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "expect slice got value",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value"}},
		},
		{
			name:        "expect slice got values",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value1&name=value2",
			paramTypes:  reflect.TypeOf([]string{}),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value1", "value2"}},
		},

		// expect string
		{
			name:        "expect string got empty",
			params:      []string{"name"},
			err:         nil,
			query:       "name=",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"name"},
			paramValues: [][]string{{""}},
		},
		{
			name:        "expect string got value",
			params:      []string{"name"},
			err:         nil,
			query:       "name=value",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"name"},
			paramValues: [][]string{{"value"}},
		},
		{
			name:        "expect string got values",
			params:      []string{"name"},
			err:         ErrInvalidType,
			field:       "name",
			query:       "name=value1&name=value2",
			paramTypes:  reflect.TypeOf(""),
			paramNames:  nil,
			paramValues: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.query)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			defer req.Body.Close()

			store := NewRequest(getServiceWithForm(tc.paramTypes, tc.params...))
			err := store.ExtractForm(req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
			if err != nil {
				cast, ok := err.(*Err)
				if !ok {
					t.Fatalf("error should be of type *Err")
				}
				if cast.Field() != tc.field {
					t.Fatalf("invalid field\nactual: %v\nexpect: %v", cast.Field(), tc.field)
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

			checkExtracted(t, tc.paramNames, tc.paramTypes.Kind(), tc.paramValues, store.Data)
		})
	}

}

func TestJsonParameters(t *testing.T) {
	tt := []struct {
		name   string
		params []string
		raw    string
		err    error

		paramTypes  reflect.Type
		paramNames  []string
		paramValues []interface{}
	}{
		// no need to fully check json because it is parsed with the standard library
		{
			name:        "empty body",
			params:      []string{},
			raw:         "",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			name:        "empty json",
			params:      []string{},
			raw:         "{}",
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			name:        "1 provided none expected",
			params:      []string{},
			raw:         `{ "a": "b" }`,
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
		{
			name:        "1 required ok",
			params:      []string{"a"},
			raw:         `{ "a": "b" }`,
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: []interface{}{"b"},
		},
		{
			name:        "1 required 1 ignored",
			params:      []string{"a"},
			raw:         `{ "a": "b", "ignored": "d" }`,
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: []interface{}{"b"},
		},
		{
			name:        "2 required 2 ok",
			params:      []string{"a", "c"},
			raw:         `{ "a": "b", "c": "d" }`,
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a", "c"},
			paramValues: []interface{}{"b", "d"},
		},
		{
			name:        "1 required null",
			params:      []string{"a"},
			raw:         `{ "a": null }`,
			err:         nil,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{"a"},
			paramValues: []interface{}{nil},
		},
		// json parse error
		{
			name:        "invalid json",
			params:      []string{},
			raw:         `{ "a": "b", }`,
			err:         ErrInvalidJSON,
			paramTypes:  reflect.TypeOf(""),
			paramNames:  []string{},
			paramValues: []interface{}{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.raw)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "application/json")
			defer req.Body.Close()
			store := NewRequest(getServiceWithForm(tc.paramTypes, tc.params...))

			err := store.ExtractForm(req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
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
						t.Fatalf("store should contain element with key %q", key)
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
		name         string
		params       []string
		rawMultipart string
		err          error

		// we must receive []byte when we send a file. In the opposite, we wait
		// for a string when we send a literal parameter
		lastParamBytes bool

		paramNames  []string
		paramValues [][]string
	}{
		{
			name:         "empty body",
			params:       []string{},
			rawMultipart: ``,
			err:          nil,
			paramNames:   []string{},
			paramValues:  [][]string{},
		},
		{
			name:   "only boundary",
			params: []string{},
			rawMultipart: `--xxx
`,
			err:         nil,
			paramNames:  []string{},
			paramValues: [][]string{},
		},
		{
			name:   "only boundaries",
			params: []string{},
			rawMultipart: `--xxx
--xxx--`,
			err:         ErrInvalidMultipart,
			paramNames:  []string{},
			paramValues: [][]string{},
		},
		{
			name:   "1 ignored part",
			params: []string{},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			paramNames:  []string{},
			paramValues: [][]string{},
		},
		{
			name:   "1 part",
			params: []string{"a"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx--`,
			paramNames:  []string{"a"},
			paramValues: [][]string{{"b"}},
		},
		{
			name:   "2 parts",
			params: []string{"a", "c"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="c"

d
--xxx--`,
			err:         nil,
			paramNames:  []string{"a", "c"},
			paramValues: [][]string{{"b"}, {"d"}},
		},
		{
			name:   "1 part 1 ignored",
			params: []string{"a"},
			rawMultipart: `--xxx
Content-Disposition: form-data; name="a"

b
--xxx
Content-Disposition: form-data; name="ignored"

x
--xxx--`,
			err:         nil,
			paramNames:  []string{"a"},
			paramValues: [][]string{{"b"}},
		},
		{
			name:           "1 param 1 file",
			params:         []string{"a", "file"},
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
			paramValues: [][]string{{"a-content"}, {base64.StdEncoding.EncodeToString(fileContent)}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.rawMultipart)
			req := httptest.NewRequest(http.MethodPost, "http://host.com", body)
			req.Header.Add("Content-Type", "multipart/form-data; boundary=xxx")
			defer req.Body.Close()

			service := &config.Service{
				Input: make(map[string]*config.Parameter),
				Form:  make(map[string]*config.Parameter),
			}

			for i, name := range tc.params {
				isLast := (i == len(tc.params)-1)

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

			store := NewRequest(getServiceWithForm(reflect.TypeOf(""), tc.params...))

			err := store.ExtractForm(req)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
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

			// check all but the last parameter
			var (
				names  = tc.paramNames
				values = tc.paramValues
			)
			if tc.lastParamBytes {
				names = names[:len(names)-1]
				values = values[:len(values)-1]
			}

			checkExtracted(t, names, reflect.String, values, store.Data)

			if !tc.lastParamBytes {
				return
			}
			var (
				name          = tc.paramNames[len(tc.paramNames)-1]
				expect        = tc.paramValues[len(tc.paramValues)-1][0]
				param, exists = store.Data[name]
			)
			if !exists {
				t.Fatalf("store should contain element with key %q", name)
				return
			}
			// expect a []byte
			expectVal, _ := base64.StdEncoding.DecodeString(expect)
			actualVal := param.([]byte)
			if bytes.Compare(actualVal, expectVal) != 0 {
				t.Fatalf("invalid bytes\nactual: %v\nexpect: %v", actualVal, expectVal)
			}
		})
	}

}
