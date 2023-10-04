package config_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/validator"
)

func TestAPI_UnmarshalJSON(t *testing.T) {
	tt := []struct {
		name string
		b    []byte

		jsonSyntaxErr bool
		jsonTypeErr   bool
		err           error
	}{
		{
			name:          "invalid json",
			b:             []byte(`{`),
			jsonSyntaxErr: true,
		},
		{
			name:        "invalid json receiver",
			b:           []byte(`{"package": 1}`),
			jsonTypeErr: true,
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
			if tc.jsonSyntaxErr {
				require.Error(t, err)
				e := &json.SyntaxError{}
				require.ErrorAs(t, err, &e, "json syntax error")
				return
			}
			if tc.jsonTypeErr {
				require.Error(t, err)
				e := &json.UnmarshalTypeError{}
				require.ErrorAs(t, err, &e, "json unmarshal type error")
				return
			}

			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAPI_Find(t *testing.T) {
	t.Parallel()

	validators := config.Validators{
		"string": validator.Wrap[string](new(validator.String)),
		"uint":   validator.Wrap[uint](new(validator.Uint)),
	}

	tt := []struct {
		name      string
		endpoints []config.Endpoint
		method    string
		fragments []string
		match     bool
	}{
		{
			name:   "match GET /",
			method: "GET", fragments: []string{},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/"},
			},
			match: true,
		},
		{
			name:   "match GET /a",
			method: "GET", fragments: []string{"a"},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a", Fragments: []string{"a"}},
			},
			match: true,
		},

		{
			name:   "mismatch GET /a/b missing fragment",
			method: "GET", fragments: []string{"a"},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b", Fragments: []string{"a", "b"}},
			},
			match: false,
		},
		{
			name:   "mismatch GET /a/b additional fragment",
			method: "GET", fragments: []string{"a", "b", "c"},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b", Fragments: []string{"a", "b"}},
			},
			match: false,
		},
		{
			name:   "mismatch GET /a/b vs /a/c",
			method: "GET", fragments: []string{"a", "c"},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b", Fragments: []string{"a", "b"}},
			},
			match: false,
		},

		{
			name:   "mismatch GET /a/b vs /a/c",
			method: "GET", fragments: []string{"a", "c"},
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b", Fragments: []string{"a", "b"}},
			},
			match: false,
		},

		{
			name:   "match GET /a/{uint}/c",
			method: "GET", fragments: []string{"a", "123", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "uint"},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: true,
		},
		{
			name:   "mismatch GET /a/{uint}/c",
			method: "GET", fragments: []string{"a", "abc", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "uint"},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},
		{
			name:   "match GET /a/{string:3}/c",
			method: "GET", fragments: []string{"a", "abc", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "string", ValidatorParams: []string{"3"}},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: true,
		},
		{
			name:   "mismatch GET /a/{string:2}/c",
			method: "GET", fragments: []string{"a", "abc", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "string", ValidatorParams: []string{"2"}},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},

		{
			name:   "err missing param",
			method: "GET", fragments: []string{"a", "123", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input:     map[string]*config.Parameter{},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},
		{
			name:   "err nil param",
			method: "GET", fragments: []string{"a", "123", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": nil,
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},
		{
			name:   "err missing validator",
			method: "GET", fragments: []string{"a", "123", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "UNKNOWN"},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},
		{
			name:   "err nil validator unexpected params",
			method: "GET", fragments: []string{"a", "123", "c"},
			endpoints: []config.Endpoint{
				{
					Method:    "GET",
					Pattern:   "/a/{var}/c",
					Fragments: []string{"a", "{var}", "c"},
					Input: map[string]*config.Parameter{
						"{var}": {ValidatorName: "uint", ValidatorParams: []string{"unexpected"}},
					},
					Captures: []*config.BraceCapture{
						{Index: 1, Name: "var"},
					},
				},
			},
			match: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			// t.Parallel()

			api := &config.API{Endpoints: make([]*config.Endpoint, 0, len(tc.endpoints))}
			for _, e := range tc.endpoints {
				api.Endpoints = append(api.Endpoints, &e)
			}

			endpoint := api.Find(tc.method, tc.fragments, validators)
			if tc.match {
				require.NotNil(t, endpoint)
				return
			}
			require.Nil(t, endpoint)
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
			name: "capture not found",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input:   map[string]*config.Parameter{},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/b",
			},
			err: config.ErrBraceCaptureUndefined,
		},
		{
			name: "invalid validator params",
			a: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint", ValidatorParams: []string{"unexpected"}},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/1",
			},
			err: config.ErrParamTypeParamsInvalid,
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
					"{var}": {ValidatorName: "string"},
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
					"{var}": {ValidatorName: "uint"},
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
					"{var}": {ValidatorName: "string"},
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
					"{var}": {ValidatorName: "uint"},
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
					"{var}": {ValidatorName: "uint"},
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
					"{var}": {ValidatorName: "uint"},
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
					"{var}": {ValidatorName: "uint"},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "uint"},
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
					"{var}": {ValidatorName: "uint"},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string"},
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
					"{var}": {ValidatorName: "uint"},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "var"},
				},
			},
			b: config.Endpoint{
				Method:  "GET",
				Pattern: "/a/{var}/c",
				Input: map[string]*config.Parameter{
					"{var}": {ValidatorName: "string"},
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
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
		t.Run(tc.name+` inverted`, func(t *testing.T) {
			tc := tc
			t.Parallel()

			api := &config.API{
				Endpoints: []*config.Endpoint{&tc.b, &tc.a},
			}

			err := api.RuntimeCheck(validators)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
