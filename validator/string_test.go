package validator_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestString(t *testing.T) {
	t.Parallel()

	var (
		maxUint32  = fmt.Sprintf("%d", math.MaxUint32)
		overUint32 = fmt.Sprintf("%d", math.MaxUint32+1)
	)

	testValidator[string](t, validator.String{}, []testCase[string]{
		{name: "3 params fail", params: make([]string, 3), match: false},
		{name: "no param ok", match: true, valid: false},
		{name: "1 param ok", params: []string{"1"}, match: true, valid: false},
		{name: "2 params ok", params: []string{"1", "3"}, match: true, valid: false},

		{name: "1 param string fail", params: []string{"abc"}, match: false},
		{name: "1 param 0 ok", params: []string{"0"}, match: true, valid: false},
		{name: "1 param -1 fail", params: []string{"-1"}, match: false},
		{name: "1 param MaxUint32 ok", params: []string{maxUint32}, match: true, valid: false},
		{name: "1 param OverUint32 fail", params: []string{overUint32}, match: false},

		{name: "2 params string int fail", params: []string{"abc", "1"}, match: false},
		{name: "2 params int string fail", params: []string{"1", "abc"}, match: false},
		{name: "2 params 0 0 ok", params: []string{"0", "0"}, match: true, valid: false},
		{name: "2 params 0 -1 fail", params: []string{"0", "-1"}, match: false},
		{name: "2 params -1 0 fail", params: []string{"0", "-1"}, match: false},
		{name: "2 params 0 MaxUint32 ok", params: []string{"0", maxUint32}, match: true, valid: false},
		{name: "2 params 0 OverUint32 fail", params: []string{"0", overUint32}, match: false},
		{name: "2 params MaxUint32 MaxUint32 ok", params: []string{maxUint32, maxUint32}, match: true, valid: false},
		{name: "2 params OverUint32 0 fail", params: []string{overUint32, "0"}, match: false},
		{name: "2 params 3 1 fail", params: []string{"3", "1"}, match: false},

		{name: "float32 invalid", value: float32(0), match: true, valid: false},
		{name: "float64 invalid", value: float64(0), match: true, valid: false},
		{name: "uint invalid", value: uint(0), match: true, valid: false},
		{name: "uint8 invalid", value: uint8(0), match: true, valid: false},
		{name: "uint16 invalid", value: uint16(0), match: true, valid: false},
		{name: "uint32 invalid", value: uint32(0), match: true, valid: false},
		{name: "uint64 invalid", value: uint64(0), match: true, valid: false},
		{name: "int invalid", value: int(0), match: true, valid: false},
		{name: "int8 invalid", value: int8(0), match: true, valid: false},
		{name: "int16 invalid", value: int16(0), match: true, valid: false},
		{name: "int32 invalid", value: int32(0), match: true, valid: false},
		{name: "int64 invalid", value: int64(0), match: true, valid: false},
		{name: "bool invalid", value: true, match: true, valid: false},
		{name: "nil invalid", value: nil, match: true, valid: false},
		{name: "struct invalid", value: struct{}{}, match: true, valid: false},

		{name: "simple abc ok", value: "abc", match: true, valid: true, extracted: "abc"},
		{name: "simple abc bytes ok", value: []byte("abc"), match: true, valid: true, extracted: "abc"},

		{name: "fixed(0) with 0 ok", params: []string{"0"}, value: "", match: true, valid: true, extracted: ""},
		{name: "fixed(0) with 1 invalid", params: []string{"0"}, value: "1", match: true, valid: false},

		{name: "fixed(16) with 16 ok", params: []string{"16"}, value: "1234567890123456", match: true, valid: true, extracted: "1234567890123456"},
		{name: "fixed(16) with 15 invalid", params: []string{"16"}, value: "123456789012345", match: true, valid: false},
		{name: "fixed(16) with 17 invalid", params: []string{"16"}, value: "12345678901234567", match: true, valid: false},

		{name: "fixed(1000) with 1000 ok", params: []string{"1000"}, value: string(make([]byte, 1000)), match: true, valid: true, extracted: string(make([]byte, 1000))},
		{name: "fixed(1000) with 999 invalid", params: []string{"1000"}, value: string(make([]byte, 1000-1)), match: true, valid: false},
		{name: "fixed(1000) with 1001 invalid", params: []string{"1000"}, value: string(make([]byte, 1000+1)), match: true, valid: false},

		{name: "var(0,0) with 0 ok", params: []string{"0", "0"}, value: "", match: true, valid: true, extracted: ""},
		{name: "var(0,0) with 1 invalid", params: []string{"0", "0"}, value: "1", match: true, valid: false},

		{name: "var(0,1) with 0 ok", params: []string{"0", "1"}, value: "", match: true, valid: true, extracted: ""},
		{name: "var(0,1) with 1 ok", params: []string{"0", "1"}, value: "1", match: true, valid: true, extracted: "1"},
		{name: "var(0,1) with 2 invalid", params: []string{"0", "1"}, value: "12", match: true, valid: false},

		{name: "var(5,16) with 4 invalid", params: []string{"5", "16"}, value: "1234", match: true, valid: false},
		{name: "var(5,16) with 5 ok", params: []string{"5", "16"}, value: "12345", match: true, valid: true, extracted: "12345"},
		{name: "var(5,16) with 6 ok", params: []string{"5", "16"}, value: "123456", match: true, valid: true, extracted: "123456"},
		{name: "var(5,16) with 7 ok", params: []string{"5", "16"}, value: "1234567", match: true, valid: true, extracted: "1234567"},
		{name: "var(5,16) with 8 ok", params: []string{"5", "16"}, value: "12345678", match: true, valid: true, extracted: "12345678"},
		{name: "var(5,16) with 9 ok", params: []string{"5", "16"}, value: "123456789", match: true, valid: true, extracted: "123456789"},
		{name: "var(5,16) with 10 ok", params: []string{"5", "16"}, value: "1234567890", match: true, valid: true, extracted: "1234567890"},
		{name: "var(5,16) with 11 ok", params: []string{"5", "16"}, value: "12345678901", match: true, valid: true, extracted: "12345678901"},
		{name: "var(5,16) with 12 ok", params: []string{"5", "16"}, value: "123456789012", match: true, valid: true, extracted: "123456789012"},
		{name: "var(5,16) with 13 ok", params: []string{"5", "16"}, value: "1234567890123", match: true, valid: true, extracted: "1234567890123"},
		{name: "var(5,16) with 14 ok", params: []string{"5", "16"}, value: "12345678901234", match: true, valid: true, extracted: "12345678901234"},
		{name: "var(5,16) with 15 ok", params: []string{"5", "16"}, value: "123456789012345", match: true, valid: true, extracted: "123456789012345"},
		{name: "var(5,16) with 16 ok", params: []string{"5", "16"}, value: "1234567890123456", match: true, valid: true, extracted: "1234567890123456"},
		{name: "var(5,16) with 17 invalid", params: []string{"5", "16"}, value: "12345678901234567", match: true, valid: false},

		{name: "var(999,1000) with 998 ok", params: []string{"999", "1000"}, value: string(make([]byte, 998)), match: true, valid: false},
		{name: "var(999,1000) with 999 ok", params: []string{"999", "1000"}, value: string(make([]byte, 999)), match: true, valid: true, extracted: string(make([]byte, 999))},
		{name: "var(999,1000) with 1000 ok", params: []string{"999", "1000"}, value: string(make([]byte, 1000)), match: true, valid: true, extracted: string(make([]byte, 1000))},
		{name: "var(999,1000) with 1001 ok", params: []string{"999", "1000"}, value: string(make([]byte, 1001)), match: true, valid: false},
	})
}
