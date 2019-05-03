package builtin_test

import (
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestString_New(t *testing.T) {
	inst := interface{}(builtin.NewString())

	switch cast := inst.(type) {
	case *builtin.String:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.String{}, cast)
	}
}

func TestString_AvailableTypes(t *testing.T) {

	inst := builtin.NewString()

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"string", true},
		{"String", false},
		{"STRING", false},
		{" string", false},
		{"string ", false},
		{" string ", false},

		{"string(1)", true},
		{"string( 1)", false},
		{"string(1 )", false},
		{"string( 1 )", false},

		{"string(1,2)", true},
		{"string(1, 2)", true},
		{"string(1,  2)", false},
		{"string( 1,2)", false},
		{"string(1,2 )", false},
		{"string( 1,2 )", false},
		{"string( 1, 2)", false},
		{"string(1, 2 )", false},
		{"string( 1, 2 )", false},
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

func TestString_AnyLength(t *testing.T) {
	const typeName = "string"

	checker := builtin.NewString().Checker(typeName)
	if checker == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	tests := []struct {
		Value interface{}
		Valid bool
	}{
		{"string", true},
		{[]byte("bytes"), false},
		{1, false},
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
func TestString_FixedLength(t *testing.T) {

	tests := []struct {
		Type  string
		Value interface{}
		Valid bool
	}{
		{"string(0)", "", true},
		{"string(0)", "", true},
		{"string(0)", "1", false},

		{"string(16)", "1234567890123456", true},
		{"string(16)", "123456789012345", false},
		{"string(16)", "12345678901234567", false},

		{"string(1000)", string(make([]byte, 1000)), true},
		{"string(1000)", string(make([]byte, 1000-1)), false},
		{"string(1000)", string(make([]byte, 1000+1)), false},
	}

	for i, test := range tests {
		checker := builtin.NewString().Checker(test.Type)
		if checker == nil {
			t.Errorf("%d: expect %q to be handled", i, test.Type)
			t.Fail()
			continue
		}

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
func TestString_VariableLength(t *testing.T) {

	tests := []struct {
		Type  string
		Value interface{}
		Valid bool
	}{
		{"string(0,0)", "", true},
		{"string(0,0)", "1", false},

		{"string(0,1)", "", true},
		{"string(0,1)", "1", true},
		{"string(0,1)", "12", false},

		{"string(5,16)", "1234", false},
		{"string(5,16)", "12345", true},
		{"string(5,16)", "123456", true},
		{"string(5,16)", "1234567", true},
		{"string(5,16)", "12345678", true},
		{"string(5,16)", "123456789", true},
		{"string(5,16)", "1234567890", true},
		{"string(5,16)", "12345678901", true},
		{"string(5,16)", "123456789012", true},
		{"string(5,16)", "1234567890123", true},
		{"string(5,16)", "12345678901234", true},
		{"string(5,16)", "123456789012345", true},
		{"string(5,16)", "1234567890123456", true},
		{"string(5,16)", "12345678901234567", false},

		{"string(999,1000)", string(make([]byte, 998)), false},
		{"string(999,1000)", string(make([]byte, 999)), true},
		{"string(999,1000)", string(make([]byte, 1000)), true},
		{"string(999,1000)", string(make([]byte, 1001)), false},
	}

	for i, test := range tests {
		checker := builtin.NewString().Checker(test.Type)
		if checker == nil {
			t.Errorf("%d: expect %q to be handled", i, test.Type)
			t.Fail()
			continue
		}

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
