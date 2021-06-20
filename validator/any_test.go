package validator_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xdrm-io/aicra/validator"
)

func TestAny_ReflectType(t *testing.T) {
	t.Parallel()

	var (
		dt       = validator.AnyType{}
		expected = reflect.TypeOf(interface{}(nil))
	)
	if dt.GoType() != expected {
		t.Fatalf("invalid GoType() %v ; expected %v", dt.GoType(), expected)
	}
}

func TestAny_AvailableTypes(t *testing.T) {
	t.Parallel()

	dt := validator.AnyType{}

	tests := []struct {
		Type    string
		Handled bool
	}{
		{"any", true},
		{" any", false},
		{"any ", false},
		{" any ", false},
		{"Any", false},
		{"ANY", false},
		{"anything-else", false},
	}

	for _, test := range tests {
		validator := dt.Validator(test.Type)

		if validator == nil {
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

func TestAny_AlwaysTrue(t *testing.T) {
	t.Parallel()

	const typeName = "any"

	validator := validator.AnyType{}.Validator(typeName)
	if validator == nil {
		t.Errorf("expect %q to be handled", typeName)
		t.Fail()
	}

	values := []interface{}{
		1,
		0.1,
		nil,
		"string",
		[]byte("bytes"),
	}

	for i, value := range values {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if _, isValid := validator(value); !isValid {
				t.Errorf("expect value to be valid")
				t.Fail()
			}
		})
	}

}
