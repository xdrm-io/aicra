package builtin_test

import (
	"fmt"
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestBool_New(t *testing.T) {
	t.Parallel()

	inst := interface{}(builtin.NewBool())

	switch cast := inst.(type) {
	case *builtin.Bool:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.Bool{}, cast)
	}
}

func TestBool_AvailableTypes(t *testing.T) {
	t.Parallel()

	inst := builtin.NewBool()

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"bool", true},
		{"Bool", false},
		{"boolean", false},
		{" bool", false},
		{"bool ", false},
		{" bool ", false},
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

func TestBool_Values(t *testing.T) {
	t.Parallel()

	const typeName = "bool"

	checker := builtin.NewBool().Checker(typeName)
	if checker == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	tests := []struct {
		Value interface{}
		Valid bool
	}{
		{true, true},
		{false, true},
		{1, false},
		{0, false},
		{-1, false},

		// json number
		{"-1", false},
		{"0", false},
		{"1", false},

		// json string
		{"true", true},
		{"false", true},
		{[]byte("true"), true},
		{[]byte("false"), true},

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
