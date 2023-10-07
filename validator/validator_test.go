package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/validator"
)

type testCase[T any] struct {
	name   string
	params []string
	value  interface{}

	match     bool
	valid     bool
	extracted T
}

func testValidator[T any](t *testing.T, validator validator.Validator[T], tt []testCase[T]) {
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			extractor := validator.Validate(tc.params)
			require.Equal(t, tc.match, extractor != nil, "match")
			if !tc.match {
				return
			}

			extracted, valid := extractor(tc.value)
			require.Equal(t, tc.valid, valid, "valid")
			if !tc.valid {
				return
			}
			require.Equal(t, tc.extracted, extracted, "extracted")
		})
	}
}

func TestValidator_Wrap(t *testing.T) {
	tt := []struct {
		name        string
		val         validator.Validator[any]
		params      []string
		validParams bool

		value     any
		extracted any
		valid     bool
	}{
		{
			name:  "string no params valid",
			val:   validator.Wrap[string](new(validator.String)),
			value: "13char string",

			validParams: true,
			extracted:   "13char string",
			valid:       true,
		},
		{
			name:        "string 2 params invalid",
			val:         validator.Wrap[string](new(validator.String)),
			params:      []string{"1", "5"},
			validParams: true,
			value:       "13char string",
			valid:       false,
		},
		{
			name:        "string 1 param valid",
			val:         validator.Wrap[string](new(validator.String)),
			params:      []string{"13"},
			validParams: true,
			value:       "13char string",
			valid:       true,
			extracted:   "13char string",
		},
		{
			name:        "invalid uint unexpected params",
			val:         validator.Wrap[uint](new(validator.Uint)),
			params:      []string{"unexpected"},
			validParams: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			extractor := tc.val.Validate(tc.params)
			if !tc.validParams {
				require.Nil(t, extractor, "extractor")
				return
			}
			require.NotNil(t, extractor, "extractor")

			extracted, valid := extractor(tc.value)
			require.Equal(t, tc.valid, valid, "valid")
			if !tc.valid {
				return
			}
			require.Equal(t, tc.extracted, extracted, "extracted")
		})
	}
}
