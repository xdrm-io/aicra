package config_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/validator"
)

func TestAPI_UnmarshalJSON(t *testing.T) {
	var jsonErr = errors.New("json error")

	tt := []struct {
		name string
		b    []byte
		err  error
	}{
		{
			name: "invalid json",
			b:    []byte(`{`),
			err:  jsonErr,
		},
		{
			name: "invalid json receiver",
			b:    []byte(`{"package": 1}`),
			err:  jsonErr,
		},
		{
			name: "missing package",
			b:    []byte(`{}`),
			err:  config.ErrPackageMissing,
		},
		{
			name: "empty package",
			b:    []byte(`{"package":""}`),
			err:  config.ErrPackageMissing,
		},
		{
			name: "missing validators",
			b:    []byte(`{"package":"pkg"}`),
			err:  config.ErrValidatorsMissing,
		},
		{
			name: "empty validators",
			b:    []byte(`{"package":"pkg","validators":{}}`),
			err:  config.ErrValidatorsMissing,
		},
		{
			name: "missing endpoints",
			b:    []byte(`{"package":"pkg","validators":{"v":{}}}`),
			err:  config.ErrEndpointsMissing,
		},
		{
			name: "invalid endpoints",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":null}`),
			err:  config.ErrEndpointsMissing,
		},
		{
			name: "invalid import name: only number",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"1":"path"}}`),
			err:  config.ErrImportAliasCharset,
		},
		{
			name: "invalid import name: start with number",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"1abc":"path"}}`),
			err:  config.ErrImportAliasCharset,
		},
		{
			name: "invalid import name: start with special char",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"_":"path"}}`),
			err:  config.ErrImportAliasCharset,
		},
		{
			name: "invalid import name: start with special char",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"_":"path"}}`),
			err:  config.ErrImportAliasCharset,
		},
		{
			name: "valid import name: 1 char",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"a":"path"}}`),
			err:  nil,
		},
		{
			name: "valid import name: all allowed chars",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"a1Z_ab":"path"}}`),
			err:  nil,
		},

		{
			name: "invalid import name: reserved name: fmt",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"fmt":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "invalid import name: reserved name: context",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"context":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "invalid import name: reserved name: http",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"http":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "invalid import name: reserved name: aicra",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"aicra":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "invalid import name: reserved name: builtin",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"builtin":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "invalid import name: reserved name: runtime",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"runtime":"path"}}`),
			err:  config.ErrImportReserved,
		},
		{
			name: "reused import path",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"alias1":"path-1", "alias2": "path-1"}}`),
			err:  config.ErrImportTwice,
		},
		{
			name: "valid local import path",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"alias1":"local_module/some-path"}}`),
			err:  nil,
		},
		{
			name: "valid remote import path",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"alias1":"github.com/xdrm-io/aicra/validator"}}`),
			err:  nil,
		},
		{
			name: "invalid import path",
			b:    []byte(`{"package":"pkg","validators":{"v":{}},"endpoints":[],"imports":{"alias1":"github.com/xdrm-io/ai'cra/validator"}}`),
			err:  config.ErrImportPathCharset,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			var api config.API
			err := json.Unmarshal(tc.b, &api)
			if tc.err != nil {
				require.Error(t, err)
				if tc.err == jsonErr {
					e1 := &json.UnmarshalTypeError{}
					e2 := &json.SyntaxError{}
					if !errors.As(err, &e1) && !errors.As(err, &e2) {
						require.Fail(t, "expected %T or %T, got %T", e1, e2, err)
					}
					return
				}
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAPI_RuntimeCheck(t *testing.T) {
	validators := config.Validators{
		"string": validator.Wrap[string](new(validator.String)),
		"uint":   validator.Wrap[uint](new(validator.Uint)),
	}

	tt := []struct {
		name string
		a    config.Endpoint
		b    config.Endpoint
		err  error
	}{
		{
			name: "invalid endpoint",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/",
				Input: map[string]*config.Parameter{
					"param": {ValidatorName: "unknown"},
				},
			},
			b: config.Endpoint{
				Method:  "PUT",
				Pattern: "/",
			},
			err: config.ErrParamTypeUnknown,
		},

		{
			name: "diff 1-fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a"},
			b:    config.Endpoint{Method: "GET", Pattern: "/b"},
			err:  nil,
		},
		{
			name: "diff 2-fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/b"},
			b:    config.Endpoint{Method: "GET", Pattern: "/a/c"},
			err:  nil,
		},
		{
			name: "diff 1-fragment 2-fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a"},
			b:    config.Endpoint{Method: "GET", Pattern: "/a/b"},
			err:  nil,
		},
		{
			name: "collide 1-fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a"},
			b:    config.Endpoint{Method: "GET", Pattern: "/a"},
			err:  config.ErrPatternCollision,
		},
		{
			name: "same diff method",
			a:    config.Endpoint{Method: "GET", Pattern: "/a"},
			b:    config.Endpoint{Method: "PUT", Pattern: "/a"},
			err:  nil,
		},
		{
			name: "collide 2nd fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/b"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "diff 2nd fragment incompatible type",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/b"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 0, Name: "var"},
				},
			},
			err: nil,
		},
		{
			name: "middle path collision",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/b/c"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "diff middle path collision type",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/b/c"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: nil,
		},
		{
			name: "diff middle path collision with params",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/bbb/c"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{"3"}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "diff middle path skip with params",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/bbb/c"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{"2"}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: nil,
		},
		{
			name: "collide left additional fragment",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/123/c"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: nil,
		},
		{
			name: "collide uint",
			a:    config.Endpoint{Method: "GET", Pattern: "/a/123"},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "colliding end captures",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "colliding end captures diff types",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
		{
			name: "colliding middle captures",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string", ValidatorParams: []string{}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			err: config.ErrPatternCollision,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			api := &config.API{
				Endpoints: []*config.Endpoint{&tc.a, &tc.b},
			}

			err := api.RuntimeCheck(validators)
			require.ErrorIs(t, err, tc.err)
		})
		t.Run(tc.name+` inverted`, func(t *testing.T) {
			tc := tc
			t.Parallel()

			api := &config.API{
				Endpoints: []*config.Endpoint{&tc.b, &tc.a},
			}

			err := api.RuntimeCheck(validators)
			require.ErrorIs(t, err, tc.err)
		})
	}
}
