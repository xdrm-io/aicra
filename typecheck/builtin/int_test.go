package builtin_test

import (
	"math"
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestInt_New(t *testing.T) {
	inst := interface{}(builtin.NewInt())

	switch cast := inst.(type) {
	case *builtin.Int:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.Int{}, cast)
	}
}

func TestInt_AvailableTypes(t *testing.T) {

	inst := builtin.NewInt()

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
		checker := inst.Checker(test.Type)

		if checker == nil {
			if test.Handled {
				t.Errorf("expect %q to be handled", test.Type)
			}
			continue
		}

		if !test.Handled {
			t.Errorf("expect %q NOT to be handled", test.Type)
		}
	}

}

func TestInt_Values(t *testing.T) {
	const typeName = "int"

	checker := builtin.NewInt().Checker(typeName)
	if checker == nil {
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

		{"string", false},
		{[]byte("bytes"), false},
		{-0.1, false},
		{0.1, false},
		{nil, false},
	}

	for i, test := range tests {
		if checker(test.Value) {
			if !test.Valid {
				t.Errorf("%d: expect value to be invalid", i)
				t.Fail()
			}
			continue
		}
		if test.Valid {
			t.Errorf("%d: expect value to be valid", i)
			t.Fail()
		}
	}

}
