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
		{name: "2 params fail", params: make([]string, 2), match: false},
		{name: "1 param fail", params: make([]string, 1), match: false},
		{name: "no param ok", match: true, valid: true},

		{name: "true ok", value: true, match: true, valid: true, extracted: true},
		{name: "false ok", value: false, match: true, valid: true, extracted: false},

		{name: "1 ok", value: 1, match: true, valid: true, extracted: 1},
		{name: "0 ok", value: 0, match: true, valid: true, extracted: 0},
		{name: "-1 ok", value: -1, match: true, valid: true, extracted: -1},
		{name: "1.23 ok", value: 1.23, match: true, valid: true, extracted: 1.23},
		{name: "string ok", value: "string", match: true, valid: true, extracted: "string"},
		{name: "bytes ok", value: []byte{1, 2, 3}, match: true, valid: true, extracted: []byte{1, 2, 3}},

		{name: "nil ok", value: nil, match: true, valid: true, extracted: nil},
		{name: "struct ok", value: struct{}{}, match: true, valid: true, extracted: struct{}{}},
		{name: "custom struct ok", value: custom{A: 123, B: true}, match: true, valid: true, extracted: custom{A: 123, B: true}},
	})
}
