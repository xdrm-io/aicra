package builtin_test

import (
	"fmt"
	"math"
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestUint_New(t *testing.T) {
	t.Parallel()

	inst := interface{}(builtin.NewUint())

	switch cast := inst.(type) {
	case *builtin.Uint:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.Uint{}, cast)
	}
}

func TestUint_AvailableTypes(t *testing.T) {
	t.Parallel()

	inst := builtin.NewUint()

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"uint", true},
		{"Uint", false},
		{"UINT", false},
		{" uint", false},
		{"uint ", false},
		{" uint ", false},
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

func TestUint_Values(t *testing.T) {
	t.Parallel()

	const typeName = "uint"

	checker := builtin.NewUint().Checker(typeName)
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
		{-1, false},
		{-math.MaxInt64, false},

		{float64(math.MinInt64), false},
		{float64(0), true},
		{float64(math.MaxInt64), true},
		// we cannot just compare because of how precision works
		{float64(math.MaxUint64 - 1024), true},
		{float64(math.MaxUint64 + 1), false},

		// json number
		{fmt.Sprintf("%d", math.MinInt64), false},
		{"-1", false},
		{"0", true},
		{"1", true},
		{fmt.Sprintf("%d", math.MaxInt64), true},
		{fmt.Sprintf("%d", uint(math.MaxUint64)), true},
		// strane offset because of how precision works
		{fmt.Sprintf("%f", float64(math.MaxUint64+1024*3)), false},

		{"string", false},
		{[]byte("bytes"), false},
		{-0.1, false},
		{0.1, false},
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
