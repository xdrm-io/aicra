package validator_test

import (
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestFloat(t *testing.T) {
	t.Parallel()

	testValidator[float64](t, validator.Float{}, []testCase[float64]{
		{name: "type FLOAT fail", typename: "FLOAT", match: false},
		{name: "type Float fail", typename: "Float", match: false},
		{name: "type ' float ' fail", typename: " float ", match: false},
		{name: "type float32 fail", typename: "float32", match: false},

		{name: "type float64 ok", typename: "float64", match: true, valid: false},
		{name: "type float ok", typename: "float", match: true, valid: false},

		{name: "float32 ok", typename: "float", value: float32(1.5), match: true, valid: true, extracted: 1.5},
		{name: "float64 ok", typename: "float", value: float64(1.5), match: true, valid: true, extracted: 1.5},

		{name: "float32 min ok", typename: "float", value: float32(-math.MaxFloat32), match: true, valid: true, extracted: -math.MaxFloat32},
		{name: "float32 max ok", typename: "float", value: float32(math.MaxFloat32), match: true, valid: true, extracted: math.MaxFloat32},

		{name: "int 0 ok", typename: "float", value: int(0), match: true, valid: true, extracted: 0},
		{name: "int8 0 ok", typename: "float", value: int8(0), match: true, valid: true, extracted: 0},
		{name: "int16 0 ok", typename: "float", value: int16(0), match: true, valid: true, extracted: 0},
		{name: "int32 0 ok", typename: "float", value: int32(0), match: true, valid: true, extracted: 0},
		{name: "int64 0 ok", typename: "float", value: int64(0), match: true, valid: true, extracted: 0},

		{name: "int min overflow", typename: "float", value: int(math.MinInt), match: true, valid: true, extracted: math.MinInt},
		{name: "int8 min ok", typename: "float", value: int8(math.MinInt8), match: true, valid: true, extracted: math.MinInt8},
		{name: "int16 min ok", typename: "float", value: int16(math.MinInt16), match: true, valid: true, extracted: math.MinInt16},
		{name: "int32 min ok", typename: "float", value: int32(math.MinInt32), match: true, valid: true, extracted: math.MinInt32},
		{name: "int64 min overflow", typename: "float", value: int64(math.MinInt64), match: true, valid: true, extracted: math.MinInt64},

		{name: "int max overflow", typename: "float", value: int(math.MaxInt), match: true, valid: false},
		{name: "int8 max ok", typename: "float", value: int8(math.MaxInt8), match: true, valid: true, extracted: math.MaxInt8},
		{name: "int16 max ok", typename: "float", value: int16(math.MaxInt16), match: true, valid: true, extracted: math.MaxInt16},
		{name: "int32 max ok", typename: "float", value: int32(math.MaxInt32), match: true, valid: true, extracted: math.MaxInt32},
		{name: "int64 max overflow", typename: "float", value: int64(math.MaxInt64), match: true, valid: false},

		{name: "uint 0 ok", typename: "float", value: uint(0), match: true, valid: true, extracted: 0},
		{name: "uint8 0 ok", typename: "float", value: uint8(0), match: true, valid: true, extracted: 0},
		{name: "uint16 0 ok", typename: "float", value: uint16(0), match: true, valid: true, extracted: 0},
		{name: "uint32 0 ok", typename: "float", value: uint32(0), match: true, valid: true, extracted: 0},
		{name: "uint64 0 ok", typename: "float", value: uint64(0), match: true, valid: true, extracted: 0},

		{name: "uint max overflow", typename: "float", value: uint(math.MaxUint), match: true, valid: false},
		{name: "uint8 max ok", typename: "float", value: uint8(math.MaxUint8), match: true, valid: true, extracted: math.MaxUint8},
		{name: "uint16 max ok", typename: "float", value: uint16(math.MaxUint16), match: true, valid: true, extracted: math.MaxUint16},
		{name: "uint32 max ok", typename: "float", value: uint32(math.MaxUint32), match: true, valid: true, extracted: math.MaxUint32},
		{name: "uint64 max overflow", typename: "float", value: uint64(math.MaxUint64), match: true, valid: false},

		{name: "string ok", typename: "float", value: "-1.5", match: true, valid: true, extracted: -1.5},
		{name: "bytes ok", typename: "float", value: []byte("-1.5"), match: true, valid: true, extracted: -1.5},

		{name: "bool invalid", typename: "float", value: true, match: true, valid: false},
		{name: "nil invalid", typename: "float", value: nil, match: true, valid: false},
		{name: "struct invalid", typename: "float", value: struct{}{}, match: true, valid: false},
	})
}
