package validator_test

import (
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestString(t *testing.T) {
	t.Parallel()

	testValidator[string](t, validator.String{}, []testCase[string]{
		{name: "type STRING fail", typename: "STRING", match: false},
		{name: "type String fail", typename: "String", match: false},
		{name: "type ' string ' fail", typename: " string ", match: false},

		{name: "type string ok", typename: "string", match: true, valid: false},
		{name: "type string(1) ok", typename: "string(1)", match: true, valid: false},
		{name: "type string(1,3) ok", typename: "string(1,3)", match: true, valid: false},
		{name: "type string(1, 3) ok", typename: "string(1, 3)", match: true, valid: false},

		{name: "float32 invalid", typename: "string", value: float32(0), match: true, valid: false},
		{name: "float64 invalid", typename: "string", value: float64(0), match: true, valid: false},
		{name: "uint invalid", typename: "string", value: uint(0), match: true, valid: false},
		{name: "uint8 invalid", typename: "string", value: uint8(0), match: true, valid: false},
		{name: "uint16 invalid", typename: "string", value: uint16(0), match: true, valid: false},
		{name: "uint32 invalid", typename: "string", value: uint32(0), match: true, valid: false},
		{name: "uint64 invalid", typename: "string", value: uint64(0), match: true, valid: false},
		{name: "int invalid", typename: "string", value: int(0), match: true, valid: false},
		{name: "int8 invalid", typename: "string", value: int8(0), match: true, valid: false},
		{name: "int16 invalid", typename: "string", value: int16(0), match: true, valid: false},
		{name: "int32 invalid", typename: "string", value: int32(0), match: true, valid: false},
		{name: "int64 invalid", typename: "string", value: int64(0), match: true, valid: false},
		{name: "bool invalid", typename: "string", value: true, match: true, valid: false},
		{name: "nil invalid", typename: "string", value: nil, match: true, valid: false},
		{name: "struct invalid", typename: "string", value: struct{}{}, match: true, valid: false},

		{name: "simple abc ok", typename: "string", value: "abc", match: true, valid: true, extracted: "abc"},
		{name: "simple abc bytes ok", typename: "string", value: []byte("abc"), match: true, valid: true, extracted: "abc"},

		{name: "fixed(0) with 0 ok", typename: "string(0)", value: "", match: true, valid: true, extracted: ""},
		{name: "fixed(0) with 1 invalid", typename: "string(0)", value: "1", match: true, valid: false},

		{name: "fixed(16) with 16 ok", typename: "string(16)", value: "1234567890123456", match: true, valid: true, extracted: "1234567890123456"},
		{name: "fixed(16) with 15 invalid", typename: "string(16)", value: "123456789012345", match: true, valid: false},
		{name: "fixed(16) with 17 invalid", typename: "string(16)", value: "12345678901234567", match: true, valid: false},

		{name: "fixed(1000) with 1000 ok", typename: "string(1000)", value: string(make([]byte, 1000)), match: true, valid: true, extracted: string(make([]byte, 1000))},
		{name: "fixed(1000) with 999 invalid", typename: "string(1000)", value: string(make([]byte, 1000-1)), match: true, valid: false},
		{name: "fixed(1000) with 1001 invalid", typename: "string(1000)", value: string(make([]byte, 1000+1)), match: true, valid: false},

		{name: "var(0,0) with 0 ok", typename: "string(0,0)", value: "", match: true, valid: true, extracted: ""},
		{name: "var(0,0) with 1 invalid", typename: "string(0,0)", value: "1", match: true, valid: false},

		{name: "var(0,1) with 0 ok", typename: "string(0,1)", value: "", match: true, valid: true, extracted: ""},
		{name: "var(0,1) with 1 ok", typename: "string(0,1)", value: "1", match: true, valid: true, extracted: "1"},
		{name: "var(0,1) with 2 invalid", typename: "string(0,1)", value: "12", match: true, valid: false},

		{name: "var(5,16) with 4 invalid", typename: "string(5,16)", value: "1234", match: true, valid: false},
		{name: "var(5,16) with 5 ok", typename: "string(5,16)", value: "12345", match: true, valid: true, extracted: "12345"},
		{name: "var(5,16) with 6 ok", typename: "string(5,16)", value: "123456", match: true, valid: true, extracted: "123456"},
		{name: "var(5,16) with 7 ok", typename: "string(5,16)", value: "1234567", match: true, valid: true, extracted: "1234567"},
		{name: "var(5,16) with 8 ok", typename: "string(5,16)", value: "12345678", match: true, valid: true, extracted: "12345678"},
		{name: "var(5,16) with 9 ok", typename: "string(5,16)", value: "123456789", match: true, valid: true, extracted: "123456789"},
		{name: "var(5,16) with 10 ok", typename: "string(5,16)", value: "1234567890", match: true, valid: true, extracted: "1234567890"},
		{name: "var(5,16) with 11 ok", typename: "string(5,16)", value: "12345678901", match: true, valid: true, extracted: "12345678901"},
		{name: "var(5,16) with 12 ok", typename: "string(5,16)", value: "123456789012", match: true, valid: true, extracted: "123456789012"},
		{name: "var(5,16) with 13 ok", typename: "string(5,16)", value: "1234567890123", match: true, valid: true, extracted: "1234567890123"},
		{name: "var(5,16) with 14 ok", typename: "string(5,16)", value: "12345678901234", match: true, valid: true, extracted: "12345678901234"},
		{name: "var(5,16) with 15 ok", typename: "string(5,16)", value: "123456789012345", match: true, valid: true, extracted: "123456789012345"},
		{name: "var(5,16) with 16 ok", typename: "string(5,16)", value: "1234567890123456", match: true, valid: true, extracted: "1234567890123456"},
		{name: "var(5,16) with 17 invalid", typename: "string(5,16)", value: "12345678901234567", match: true, valid: false},

		{name: "var(999,1000) with 998 ok", typename: "string(999,1000)", value: string(make([]byte, 998)), match: true, valid: false},
		{name: "var(999,1000) with 999 ok", typename: "string(999,1000)", value: string(make([]byte, 999)), match: true, valid: true, extracted: string(make([]byte, 999))},
		{name: "var(999,1000) with 1000 ok", typename: "string(999,1000)", value: string(make([]byte, 1000)), match: true, valid: true, extracted: string(make([]byte, 1000))},
		{name: "var(999,1000) with 1001 ok", typename: "string(999,1000)", value: string(make([]byte, 1001)), match: true, valid: false},
	})
}
