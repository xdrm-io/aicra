package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/validator"
)

type testCase[T any] struct {
	name     string
	typename string
	value    interface{}

	match     bool
	valid     bool
	extracted T
}

func testValidator[T any](t *testing.T, validator validator.Validator[T], tt []testCase[T]) {
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			extractor := validator.Validate(tc.typename)
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
