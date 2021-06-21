package validator_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestInt_AvailableTypes(t *testing.T) {
	t.Parallel()

	dt := validator.IntType{}

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"int", true},
		{"Int", false},
		{"INT", false},
		{" int", false},
		{"int ", false},
		{" int ", false},
	}

	for _, test := range tests {
		t.Run(test.Type, func(t *testing.T) {
			validator := dt.Validator(test.Type)
			if validator == nil {
				if test.Handled {
					t.Errorf("expect %q to be handled", test.Type)
					t.Fail()
				}
				return
			}

			if !test.Handled {
				t.Errorf("expect %q NOT to be handled", test.Type)
				t.Fail()
			}
		})
	}

}

func TestInt_Values(t *testing.T) {
	t.Parallel()

	const typeName = "int"

	validator := validator.IntType{}.Validator(typeName)
	if validator == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	tests := []struct {
		Value interface{}
		Valid bool
	}{
		{-math.MaxInt64, true},
		{-1, true},
		{0, true},
		{1, true},
		{math.MaxInt64, true},

		// overflows from type conversion
		{uint(math.MaxInt64), true},
		{uint(math.MaxInt64 + 1), false},

		{float64(math.MinInt64), true},
		// we cannot just substract 1 because of how precision works
		{float64(math.MinInt64 - 1024 - 1), false},

		// WARNING : this is due to how floats are compared
		{float64(math.MaxInt64), false},
		// we cannot just add 1 because of how precision works
		{float64(math.MaxInt64 + 1024 + 2), false},

		// json number
		{fmt.Sprintf("%f", float64(math.MinInt64-1024-1)), false},
		{fmt.Sprintf("%d", math.MinInt64), true},
		{"-1", true},
		{"0", true},
		{"1", true},
		{fmt.Sprintf("%d", math.MaxInt64), true},
		{fmt.Sprintf("%f", float64(math.MaxInt64+1024+2)), false},

		{"string", false},
		{[]byte("bytes"), false},
		{-0.1, false},
		{0.1, false},
		{nil, false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if _, isValid := validator(test.Value); isValid {
				if !test.Valid {
					t.Errorf("expect value to be invalid")
					t.Fail()
				}
				return
			}
			if test.Valid {
				t.Errorf("expect value to be valid")
				t.Fail()
			}
		})
	}

}
