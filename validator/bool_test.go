package validator_test

import (
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestBool(t *testing.T) {
	t.Parallel()

	testValidator[bool](t, validator.Bool{}, []testCase[bool]{
		{name: "2 params fail", params: make([]string, 2), match: false},
		{name: "1 param fail", params: make([]string, 1), match: false},
		{name: "no param ok", match: true, valid: false},

		{name: "true ok", value: true, match: true, valid: true, extracted: true},
		{name: "false ok", value: false, match: true, valid: true, extracted: false},

		{name: "true json string ok", value: "true", match: true, valid: true, extracted: true},
		{name: "false json string ok", value: "false", match: true, valid: true, extracted: false},

		{name: "true []byte ok", value: []byte("true"), match: true, valid: true, extracted: true},
		{name: "false []byte ok", value: []byte("false"), match: true, valid: true, extracted: false},

		{name: "1 invalid", value: 1, match: true, valid: false},
		{name: "0 invalid", value: 0, match: true, valid: false},
		{name: "-1 invalid", value: -1, match: true, valid: false},

		{name: "1 json string invalid", value: "1", match: true, valid: false},
		{name: "0 json string invalid", value: "0", match: true, valid: false},
		{name: "-1 json string invalid", value: "-1", match: true, valid: false},

		{name: "nil invalid", value: nil, match: true, valid: false},
		{name: "struct invalid", value: struct{}{}, match: true, valid: false},
	})
}
