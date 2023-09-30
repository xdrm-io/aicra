package config_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/validator"
)

func TestParam(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name   string
		config string

		jsonErr bool
		err     error
		param   config.Parameter
	}{
		{
			name:    "invalid json",
			config:  `{""}`,
			jsonErr: true,
		},
		{
			name:   "empty",
			config: `{}`,
			err:    config.ErrParamTypeMissing,
		},
		{
			name:   "invalid type",
			config: `{ "type": "mytype()" }`,
			err:    config.ErrParamTypeInvalid,
		},
		{
			name:   "simple type",
			config: `{ "type": "mytype" }`,
			param: config.Parameter{
				Type:            "mytype",
				ValidatorName:   "mytype",
				ValidatorParams: []string{},
			},
		},
		{
			name:   "type 1 param",
			config: `{ "type": "mytype(param1)" }`,
			param: config.Parameter{
				Type:            "mytype(param1)",
				ValidatorName:   "mytype",
				ValidatorParams: []string{"param1"},
			},
		},
		{
			name:   "type 2 params",
			config: `{ "type": "mytype(param1,param2)" }`,
			param: config.Parameter{
				Type:            "mytype(param1,param2)",
				ValidatorName:   "mytype",
				ValidatorParams: []string{"param1", "param2"},
			},
		},
		{
			name:   "unexported rename",
			config: `{ "type": "mytype", "name": "lowercase" }`,
			err:    config.ErrNameUnexported,
		},
		{
			name:   "unexported rename",
			config: `{ "type": "mytype", "name": "lowercase" }`,
			err:    config.ErrParamRenameInvalid,
		},
		{
			name:   "invalid rename",
			config: `{ "type": "mytype", "name": "Invali*d" }`,
			err:    config.ErrParamRenameInvalid,
		},
		{
			name:   "optional 2 params",
			config: `{ "type": "?mytype(param1,param2)" }`,
			param: config.Parameter{
				Optional:        true,
				Type:            "mytype(param1,param2)",
				ValidatorName:   "mytype",
				ValidatorParams: []string{"param1", "param2"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			var p config.Parameter
			err := json.Unmarshal([]byte(tc.config), &p)
			var jsonErr *json.SyntaxError
			require.Equal(t, tc.jsonErr, errors.As(err, &jsonErr), "json error")
			if tc.jsonErr {
				return
			}
			require.ErrorIs(t, err, tc.err)
			if err != nil {
				return
			}
			require.EqualValues(t, tc.param, p)
		})
	}
}

type genericValidator func([]string) validator.ExtractFunc[any]

func wrapValidator[T any](v validator.Validator[T]) genericValidator {
	return func(p []string) validator.ExtractFunc[any] {
		extractor := v.Validate(p)
		if extractor == nil {
			return nil
		}
		return func(val any) (any, bool) {
			return extractor(val)
		}
	}
}

func TestParamRuntimeCheck(t *testing.T) {
	t.Parallel()

	validators := config.Validators{
		"string": validator.Wrap[string](validator.String{}),
		"uint":   validator.Wrap[uint](validator.Uint{}),
		"nil":    nil,
	}

	tt := []struct {
		name  string
		param config.Parameter

		err error
	}{
		{
			name: "validator not found",
			param: config.Parameter{
				Type: "unknown",
			},
			err: config.ErrParamTypeUnknown,
		},
		{
			name: "nil validator",
			param: config.Parameter{
				Type: "nil",
			},
			err: config.ErrParamTypeUnknown,
		},
		{
			name: "unexpected parameter",
			param: config.Parameter{
				ValidatorName:   "uint",
				ValidatorParams: []string{"unexpected"},
			},
			err: config.ErrParamTypeParamsInvalid,
		},
		{
			name: "no param ok",
			param: config.Parameter{
				ValidatorName:   "uint",
				ValidatorParams: []string{},
			},
		},
		{
			name: "nil param ok",
			param: config.Parameter{
				ValidatorName:   "uint",
				ValidatorParams: nil,
			},
		},
		{
			name: "1 param ok",
			param: config.Parameter{
				ValidatorName:   "string",
				ValidatorParams: []string{"1"},
			},
		},
		{
			name: "1 param invalid",
			param: config.Parameter{
				ValidatorName:   "string",
				ValidatorParams: []string{"NaN"},
			},
			err: config.ErrParamTypeParamsInvalid,
		},
		{
			name: "2 params ok",
			param: config.Parameter{
				ValidatorName:   "string",
				ValidatorParams: []string{"1", "3"},
			},
		},
		{
			name: "2 params invalid",
			param: config.Parameter{
				ValidatorName:   "string",
				ValidatorParams: []string{"1", "NaN"},
			},
			err: config.ErrParamTypeParamsInvalid,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			err := tc.param.RuntimeCheck(validators)
			require.ErrorIs(t, err, tc.err)
		})
	}
}
