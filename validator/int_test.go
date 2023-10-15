package validator_test

import (
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestInt(t *testing.T) {
	t.Parallel()

	testValidator[int](t, validator.Int{}, []testCase[int]{
		{name: "2 params fail", params: make([]string, 2), match: false},
		{name: "1 param fail", params: make([]string, 1), match: false},
		{name: "no param ok", match: true, valid: false},

		{name: "float32 lost precision", value: float32(1.5), match: true, valid: false},
		{name: "float64 lost precision", value: float64(1.5), match: true, valid: false},

		{name: "float32 min overflow", value: float32(-math.MaxFloat32), match: true, valid: false},
		{name: "float32 max overflow", value: float32(math.MaxFloat32), match: true, valid: false},
		{name: "float64 min overflow", value: float64(-math.MaxFloat64), match: true, valid: false},
		{name: "float64 max overflow", value: float64(math.MaxFloat64), match: true, valid: false},

		{name: "int 0 ok", value: int(0), match: true, valid: true, extracted: 0},
		{name: "int8 0 ok", value: int8(0), match: true, valid: true, extracted: 0},
		{name: "int16 0 ok", value: int16(0), match: true, valid: true, extracted: 0},
		{name: "int32 0 ok", value: int32(0), match: true, valid: true, extracted: 0},
		{name: "int64 0 ok", value: int64(0), match: true, valid: true, extracted: 0},

		{name: "int min ok", value: int(math.MinInt), match: true, valid: true, extracted: math.MinInt},
		{name: "int8 min ok", value: int8(math.MinInt8), match: true, valid: true, extracted: math.MinInt8},
		{name: "int16 min ok", value: int16(math.MinInt16), match: true, valid: true, extracted: math.MinInt16},
		{name: "int32 min ok", value: int32(math.MinInt32), match: true, valid: true, extracted: math.MinInt32},
		{name: "int64 min ok", value: int64(math.MinInt64), match: true, valid: true, extracted: math.MinInt64},

		{name: "int max ok", value: int(math.MaxInt), match: true, valid: true, extracted: math.MaxInt},
		{name: "int8 max ok", value: int8(math.MaxInt8), match: true, valid: true, extracted: math.MaxInt8},
		{name: "int16 max ok", value: int16(math.MaxInt16), match: true, valid: true, extracted: math.MaxInt16},
		{name: "int32 max ok", value: int32(math.MaxInt32), match: true, valid: true, extracted: math.MaxInt32},
		{name: "int64 max ok", value: int64(math.MaxInt64), match: true, valid: true, extracted: math.MaxInt64},

		{name: "uint 0 ok", value: uint(0), match: true, valid: true, extracted: 0},
		{name: "uint8 0 ok", value: uint8(0), match: true, valid: true, extracted: 0},
		{name: "uint16 0 ok", value: uint16(0), match: true, valid: true, extracted: 0},
		{name: "uint32 0 ok", value: uint32(0), match: true, valid: true, extracted: 0},
		{name: "uint64 0 ok", value: uint64(0), match: true, valid: true, extracted: 0},

		{name: "uint max overflow", value: uint(math.MaxUint), match: true, valid: false},
		{name: "uint8 max ok", value: uint8(math.MaxUint8), match: true, valid: true, extracted: math.MaxUint8},
		{name: "uint16 max ok", value: uint16(math.MaxUint16), match: true, valid: true, extracted: math.MaxUint16},
		{name: "uint32 max ok", value: uint32(math.MaxUint32), match: true, valid: true, extracted: math.MaxUint32},
		{name: "uint64 max overflow", value: uint64(math.MaxUint64), match: true, valid: false},

		{name: "string ok", value: "123", match: true, valid: true, extracted: 123},
		{name: "bytes ok", value: []byte("-456"), match: true, valid: true, extracted: -456},

		{name: "bool invalid", value: true, match: true, valid: false},
		{name: "nil invalid", value: nil, match: true, valid: false},
		{name: "struct invalid", value: struct{}{}, match: true, valid: false},
	})
}
