package builtin_test

import (
	"fmt"
	"math"
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestFloat64_New(t *testing.T) {
	t.Parallel()

	inst := interface{}(builtin.NewFloat64())

	switch cast := inst.(type) {
	case *builtin.Float64:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.Float64{}, cast)
	}
}

func TestFloat64_AvailableTypes(t *testing.T) {
	t.Parallel()

	inst := builtin.NewFloat64()

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"float", true},
		{"float64", true},
		{"Float", false},
		{"Float64", false},
		{"FLOAT", false},
		{"FLOAT64", false},
		{" float", false},
		{"float ", false},
		{" float ", false},
		{" float64", false},
		{"float64 ", false},
		{" float64 ", false},
	}

	for _, test := range tests {
		t.Run(test.Type, func(t *testing.T) {
			checker := inst.Checker(test.Type)
			if checker == nil {
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

func TestFloat64_Values(t *testing.T) {
	t.Parallel()

	const typeName = "float"

	checker := builtin.NewFloat64().Checker(typeName)
	if checker == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	tests := []struct {
		Value interface{}
		Valid bool
	}{
		{uint(0), true},
		{uint(math.MaxInt64), true},
		{uint(math.MaxUint64), true},
		{-1, true},
		{-math.MaxInt64, true},

		{float64(math.MinInt64), true},
		{float64(0), true},
		{float64(math.MaxInt64), true},
		// we cannot just compare because of how precision works
		{float64(math.MaxUint64 - 1024), true},
		{float64(math.MaxUint64 + 1), true},

		// json number
		{fmt.Sprintf("%f", -math.MaxFloat64), true},
		{"-1", true},
		{"0", true},
		{"1", true},
		{fmt.Sprintf("%d", math.MaxInt64), true},
		{fmt.Sprintf("%d", uint(math.MaxUint64)), true},
		{fmt.Sprintf("%f", float64(math.MaxFloat64)), true},

		{"string", false},
		{[]byte("bytes"), false},
		{-0.1, true},
		{0.1, true},
		{nil, false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if checker(test.Value) {
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
