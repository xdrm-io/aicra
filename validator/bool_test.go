package validator_test

import (
	"fmt"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestBool_AvailableTypes(t *testing.T) {
	t.Parallel()

	dt := validator.BoolType{}

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

func TestBool_Values(t *testing.T) {
	t.Parallel()

	const typeName = "bool"

	validator := validator.BoolType{}.Validator(typeName)
	if validator == nil {
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
