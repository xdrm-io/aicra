package validator_test

import (
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestAny(t *testing.T) {
	t.Parallel()

	type custom struct {
		A uint
		B bool
	}

	testValidator[any](t, validator.Any{}, []testCase[any]{
		{name: "type ANY fail", typename: "ANY", match: false},
		{name: "type Any fail", typename: "Any", match: false},
		{name: "type ' any ' fail", typename: " any ", match: false},

		{name: "true ok", typename: "any", value: true, match: true, valid: true, extracted: true},
		{name: "false ok", typename: "any", value: false, match: true, valid: true, extracted: false},

		{name: "1 ok", typename: "any", value: 1, match: true, valid: true, extracted: 1},
		{name: "0 ok", typename: "any", value: 0, match: true, valid: true, extracted: 0},
		{name: "-1 ok", typename: "any", value: -1, match: true, valid: true, extracted: -1},
		{name: "1.23 ok", typename: "any", value: 1.23, match: true, valid: true, extracted: 1.23},
		{name: "string ok", typename: "any", value: "string", match: true, valid: true, extracted: "string"},
		{name: "bytes ok", typename: "any", value: []byte{1, 2, 3}, match: true, valid: true, extracted: []byte{1, 2, 3}},

		{name: "nil ok", typename: "any", value: nil, match: true, valid: true, extracted: nil},
		{name: "struct ok", typename: "any", value: struct{}{}, match: true, valid: true, extracted: struct{}{}},
		{name: "custom struct ok", typename: "any", value: custom{A: 123, B: true}, match: true, valid: true, extracted: custom{A: 123, B: true}},
	})
}
