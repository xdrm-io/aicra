package builtin_test

import (
	"fmt"
	"testing"

	"github.com/xdrm-io/aicra/validator/builtin"
)

func TestString_AvailableTypes(t *testing.T) {
	t.Parallel()

	dt := builtin.StringDataType{}

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

		{"string()", false},
		{"string(a)", false},
		{"string(-1)", false},

		{"string(,)", false},
		{"string(1,b)", false},
		{"string(a,b)", false},
		{"string(a,1)", false},
		{"string(-1,1)", false},
		{"string(1,-1)", false},
		{"string(-1,-1)", false},

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
		t.Run(test.Type, func(t *testing.T) {
			validator := dt.Validator(test.Type)

			if validator == nil {
				if test.Handled {
					t.Errorf("expect %q to be handled", test.Type)
				}
				return
			}

			if !test.Handled {
				t.Errorf("expect %q NOT to be handled", test.Type)
			}
		})
	}

}

func TestString_AnyLength(t *testing.T) {
	t.Parallel()

	const typeName = "string"

	validator := builtin.StringDataType{}.Validator(typeName)
	if validator == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	tests := []struct {
		Value interface{}
		Valid bool
	}{
		{"string", true},
		{[]byte("bytes"), true},
		{1, false},
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
func TestString_FixedLength(t *testing.T) {
	t.Parallel()

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
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			validator := builtin.StringDataType{}.Validator(test.Type)
			if validator == nil {
				t.Errorf("expect %q to be handled", test.Type)
				t.Fail()
				return
			}

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
func TestString_VariableLength(t *testing.T) {
	t.Parallel()

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
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			validator := builtin.StringDataType{}.Validator(test.Type)
			if validator == nil {
				t.Errorf("expect %q to be handled", test.Type)
				t.Fail()
				return
			}

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
