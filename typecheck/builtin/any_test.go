package builtin_test

import (
	"fmt"
	"testing"

	"git.xdrm.io/go/aicra/typecheck/builtin"
)

func TestAny_New(t *testing.T) {
	t.Parallel()

	inst := interface{}(builtin.NewAny())

	switch cast := inst.(type) {
	case *builtin.Any:
		return
	default:
		t.Errorf("expect %T ; got %T", &builtin.Any{}, cast)
	}
}

func TestAny_AvailableTypes(t *testing.T) {
	t.Parallel()

	inst := builtin.NewAny()

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

func TestAny_AlwaysTrue(t *testing.T) {
	t.Parallel()

	const typeName = "any"

	checker := builtin.NewAny().Checker(typeName)
	if checker == nil {
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
			if !checker(value) {
				t.Errorf("expect value to be valid")
				t.Fail()
			}
		})
	}

}
