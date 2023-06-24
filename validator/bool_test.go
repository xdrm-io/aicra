package validator_test

import (
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestBool(t *testing.T) {
	t.Parallel()

	testValidator[bool](t, validator.Bool{}, []testCase[bool]{
		{name: "type BOOL fail", typename: "BOOL", match: false},
		{name: "type Bool fail", typename: "Bool", match: false},
		{name: "type ' bool ' fail", typename: " bool ", match: false},

		{name: "true ok", typename: "bool", value: true, match: true, valid: true, extracted: true},
		{name: "false ok", typename: "bool", value: false, match: true, valid: true, extracted: false},

		{name: "true json string ok", typename: "bool", value: "true", match: true, valid: true, extracted: true},
		{name: "false json string ok", typename: "bool", value: "false", match: true, valid: true, extracted: false},

		{name: "true []byte ok", typename: "bool", value: []byte("true"), match: true, valid: true, extracted: true},
		{name: "false []byte ok", typename: "bool", value: []byte("false"), match: true, valid: true, extracted: false},

		{name: "1 invalid", typename: "bool", value: 1, match: true, valid: false},
		{name: "0 invalid", typename: "bool", value: 0, match: true, valid: false},
		{name: "-1 invalid", typename: "bool", value: -1, match: true, valid: false},

		{name: "1 json string invalid", typename: "bool", value: "1", match: true, valid: false},
		{name: "0 json string invalid", typename: "bool", value: "0", match: true, valid: false},
		{name: "-1 json string invalid", typename: "bool", value: "-1", match: true, valid: false},

		{name: "nil invalid", typename: "bool", value: nil, match: true, valid: false},
		{name: "struct invalid", typename: "bool", value: struct{}{}, match: true, valid: false},
	})
}
