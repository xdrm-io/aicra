package validator_test

import (
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestUint(t *testing.T) {
	t.Parallel()

	testValidator[uint](t, validator.Uint{}, []testCase[uint]{
		{name: "type UINT fail", typename: "UINT", match: false},
		{name: "type Uint fail", typename: "Uint", match: false},
		{name: "type ' uint ' fail", typename: " uint ", match: false},

		{name: "type uint ok", typename: "uint", match: true, valid: false},

		{name: "float32 lost precision", typename: "uint", value: float32(1.5), match: true, valid: false},
		{name: "float64 lost precision", typename: "uint", value: float64(1.5), match: true, valid: false},

		{name: "float32 min overflow", typename: "uint", value: float32(-math.MaxFloat32), match: true, valid: false},
		{name: "float32 max overflow", typename: "uint", value: float32(math.MaxFloat32), match: true, valid: false},
		{name: "float64 min overflow", typename: "uint", value: float64(-math.MaxFloat64), match: true, valid: false},
		{name: "float64 max overflow", typename: "uint", value: float64(math.MaxFloat64), match: true, valid: false},

		{name: "int 0 ok", typename: "uint", value: int(0), match: true, valid: true, extracted: 0},
		{name: "int8 0 ok", typename: "uint", value: int8(0), match: true, valid: true, extracted: 0},
		{name: "int16 0 ok", typename: "uint", value: int16(0), match: true, valid: true, extracted: 0},
		{name: "int32 0 ok", typename: "uint", value: int32(0), match: true, valid: true, extracted: 0},
		{name: "int64 0 ok", typename: "uint", value: int64(0), match: true, valid: true, extracted: 0},

		{name: "int min overflow", typename: "uint", value: int(math.MinInt), match: true, valid: false},
		{name: "int8 min overflow", typename: "uint", value: int8(math.MinInt8), match: true, valid: false},
		{name: "int16 min overflow", typename: "uint", value: int16(math.MinInt16), match: true, valid: false},
		{name: "int32 min overflow", typename: "uint", value: int32(math.MinInt32), match: true, valid: false},
		{name: "int64 min overflow", typename: "uint", value: int64(math.MinInt64), match: true, valid: false},

		{name: "int max ok", typename: "uint", value: int(math.MaxInt), match: true, valid: true, extracted: math.MaxInt},
		{name: "int8 max ok", typename: "uint", value: int8(math.MaxInt8), match: true, valid: true, extracted: math.MaxInt8},
		{name: "int16 max ok", typename: "uint", value: int16(math.MaxInt16), match: true, valid: true, extracted: math.MaxInt16},
		{name: "int32 max ok", typename: "uint", value: int32(math.MaxInt32), match: true, valid: true, extracted: math.MaxInt32},
		{name: "int64 max ok", typename: "uint", value: int64(math.MaxInt64), match: true, valid: true, extracted: math.MaxInt64},

		{name: "uint 0 ok", typename: "uint", value: uint(0), match: true, valid: true, extracted: 0},
		{name: "uint8 0 ok", typename: "uint", value: uint8(0), match: true, valid: true, extracted: 0},
		{name: "uint16 0 ok", typename: "uint", value: uint16(0), match: true, valid: true, extracted: 0},
		{name: "uint32 0 ok", typename: "uint", value: uint32(0), match: true, valid: true, extracted: 0},
		{name: "uint64 0 ok", typename: "uint", value: uint64(0), match: true, valid: true, extracted: 0},

		{name: "uint max", typename: "uint", value: uint(math.MaxUint), match: true, valid: true, extracted: math.MaxUint},
		{name: "uint8 max ok", typename: "uint", value: uint8(math.MaxUint8), match: true, valid: true, extracted: math.MaxUint8},
		{name: "uint16 max ok", typename: "uint", value: uint16(math.MaxUint16), match: true, valid: true, extracted: math.MaxUint16},
		{name: "uint32 max ok", typename: "uint", value: uint32(math.MaxUint32), match: true, valid: true, extracted: math.MaxUint32},
		{name: "uint64 max", typename: "uint", value: uint64(math.MaxUint64), match: true, valid: true, extracted: math.MaxUint64},

		{name: "string ok", typename: "uint", value: "123", match: true, valid: true, extracted: 123},
		{name: "bytes ok", typename: "uint", value: []byte("456"), match: true, valid: true, extracted: 456},

		{name: "bool invalid", typename: "uint", value: true, match: true, valid: false},
		{name: "nil invalid", typename: "uint", value: nil, match: true, valid: false},
		{name: "struct invalid", typename: "uint", value: struct{}{}, match: true, valid: false},
	})
}
