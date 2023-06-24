package validator_test

import (
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestInt(t *testing.T) {
	t.Parallel()

	testValidator[int](t, validator.Int{}, []testCase[int]{
		{name: "type INT fail", typename: "INT", match: false},
		{name: "type Int fail", typename: "Int", match: false},
		{name: "type ' int ' fail", typename: " int ", match: false},

		{name: "type int ok", typename: "int", match: true, valid: false},

		{name: "float32 lost precision", typename: "int", value: float32(1.5), match: true, valid: false},
		{name: "float64 lost precision", typename: "int", value: float64(1.5), match: true, valid: false},

		{name: "float32 min overflow", typename: "int", value: float32(-math.MaxFloat32), match: true, valid: false},
		{name: "float32 max overflow", typename: "int", value: float32(math.MaxFloat32), match: true, valid: false},
		{name: "float64 min overflow", typename: "int", value: float64(-math.MaxFloat64), match: true, valid: false},
		{name: "float64 max overflow", typename: "int", value: float64(math.MaxFloat64), match: true, valid: false},

		{name: "int 0 ok", typename: "int", value: int(0), match: true, valid: true, extracted: 0},
		{name: "int8 0 ok", typename: "int", value: int8(0), match: true, valid: true, extracted: 0},
		{name: "int16 0 ok", typename: "int", value: int16(0), match: true, valid: true, extracted: 0},
		{name: "int32 0 ok", typename: "int", value: int32(0), match: true, valid: true, extracted: 0},
		{name: "int64 0 ok", typename: "int", value: int64(0), match: true, valid: true, extracted: 0},

		{name: "int min ok", typename: "int", value: int(math.MinInt), match: true, valid: true, extracted: math.MinInt},
		{name: "int8 min ok", typename: "int", value: int8(math.MinInt8), match: true, valid: true, extracted: math.MinInt8},
		{name: "int16 min ok", typename: "int", value: int16(math.MinInt16), match: true, valid: true, extracted: math.MinInt16},
		{name: "int32 min ok", typename: "int", value: int32(math.MinInt32), match: true, valid: true, extracted: math.MinInt32},
		{name: "int64 min ok", typename: "int", value: int64(math.MinInt64), match: true, valid: true, extracted: math.MinInt64},

		{name: "int max ok", typename: "int", value: int(math.MaxInt), match: true, valid: true, extracted: math.MaxInt},
		{name: "int8 max ok", typename: "int", value: int8(math.MaxInt8), match: true, valid: true, extracted: math.MaxInt8},
		{name: "int16 max ok", typename: "int", value: int16(math.MaxInt16), match: true, valid: true, extracted: math.MaxInt16},
		{name: "int32 max ok", typename: "int", value: int32(math.MaxInt32), match: true, valid: true, extracted: math.MaxInt32},
		{name: "int64 max ok", typename: "int", value: int64(math.MaxInt64), match: true, valid: true, extracted: math.MaxInt64},

		{name: "uint 0 ok", typename: "int", value: uint(0), match: true, valid: true, extracted: 0},
		{name: "uint8 0 ok", typename: "int", value: uint8(0), match: true, valid: true, extracted: 0},
		{name: "uint16 0 ok", typename: "int", value: uint16(0), match: true, valid: true, extracted: 0},
		{name: "uint32 0 ok", typename: "int", value: uint32(0), match: true, valid: true, extracted: 0},
		{name: "uint64 0 ok", typename: "int", value: uint64(0), match: true, valid: true, extracted: 0},

		{name: "uint max overflow", typename: "int", value: uint(math.MaxUint), match: true, valid: false},
		{name: "uint8 max ok", typename: "int", value: uint8(math.MaxUint8), match: true, valid: true, extracted: math.MaxUint8},
		{name: "uint16 max ok", typename: "int", value: uint16(math.MaxUint16), match: true, valid: true, extracted: math.MaxUint16},
		{name: "uint32 max ok", typename: "int", value: uint32(math.MaxUint32), match: true, valid: true, extracted: math.MaxUint32},
		{name: "uint64 max overflow", typename: "int", value: uint64(math.MaxUint64), match: true, valid: false},

		{name: "string ok", typename: "int", value: "123", match: true, valid: true, extracted: 123},
		{name: "bytes ok", typename: "int", value: []byte("-456"), match: true, valid: true, extracted: -456},

		{name: "bool invalid", typename: "int", value: true, match: true, valid: false},
		{name: "nil invalid", typename: "int", value: nil, match: true, valid: false},
		{name: "struct invalid", typename: "int", value: struct{}{}, match: true, valid: false},
	})
}
