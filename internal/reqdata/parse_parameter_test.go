package reqdata

import (
	"fmt"
	"math"
	"testing"
)

func TestSimpleString(t *testing.T) {
	t.Parallel()

	p := parseParameter("some-string")

	cast, canCast := p.(string)
	if !canCast {
		t.Fatalf("expected parameter to be a string")
	}

	if cast != "some-string" {
		t.Fatalf("expected parameter to equal 'some-string', got %q", cast)
	}
}

func TestSimpleFloat(t *testing.T) {
	t.Parallel()

	tcases := []float64{12.3456789, -12.3456789, 0.0000001, -0.0000001}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			p := parseParameter(tcase)

			cast, canCast := p.(float64)
			if !canCast {
				t.Fatalf("expected parameter to be a float64")
			}

			if math.Abs(cast-tcase) > 0.00000001 {
				t.Fatalf("expected parameter to equal '%f', got '%f'", tcase, cast)
			}
		})
	}
}

func TestSimpleBool(t *testing.T) {
	t.Parallel()

	tcases := []bool{true, false}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			p := parseParameter(tcase)

			cast, canCast := p.(bool)
			if !canCast {
				t.Fatalf("expected parameter to be a bool")
			}

			if cast != tcase {
				t.Fatalf("expected parameter to equal '%t', got '%t'", tcase, cast)
			}
		})
	}
}

func TestJsonStringSlice(t *testing.T) {
	t.Parallel()

	p := parseParameter(`["str1", "str2"]`)

	slice, canCast := p.([]interface{})
	if !canCast {
		t.Fatalf("expected parameter to be a []interface{}")
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 values, got %d", len(slice))
	}

	results := []string{"str1", "str2"}

	for i, res := range results {

		cast, canCast := slice[i].(string)
		if !canCast {
			t.Errorf("expected parameter %d to be a []string", i)
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be %q, got %q", res, cast)
			continue
		}

	}

}

func TestStringSlice(t *testing.T) {
	t.Parallel()

	p := parseParameter([]string{"str1", "str2"})

	slice, canCast := p.([]interface{})
	if !canCast {
		t.Fatalf("expected parameter to be a []interface{}")
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 values, got %d", len(slice))
	}

	results := []string{"str1", "str2"}

	for i, res := range results {

		cast, canCast := slice[i].(string)
		if !canCast {
			t.Errorf("expected parameter %d to be a []string", i)
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be %q, got %q", res, cast)
			continue
		}

	}

}

func TestJsonPrimitiveBool(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		Raw       string
		BoolValue bool
	}{
		{"true", true},
		{"false", false},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			p := parseParameter(tcase.Raw)

			cast, canCast := p.(bool)
			if !canCast {
				t.Fatalf("expected parameter to be a bool")
			}

			if cast != tcase.BoolValue {
				t.Fatalf("expected a value of %t, got %t", tcase.BoolValue, cast)
			}
		})
	}

}

func TestJsonPrimitiveFloat(t *testing.T) {
	t.Parallel()

	tcases := []struct {
		Raw        string
		FloatValue float64
	}{
		{"1", 1},
		{"-1", -1},

		{"0.001", 0.001},
		{"-0.001", -0.001},

		{"1.9992", 1.9992},
		{"-1.9992", -1.9992},

		{"19992", 19992},
		{"-19992", -19992},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			p := parseParameter(tcase.Raw)

			cast, canCast := p.(float64)
			if !canCast {
				t.Fatalf("expected parameter to be a float64")
			}

			if math.Abs(cast-tcase.FloatValue) > 0.00001 {
				t.Fatalf("expected a value of %f, got %f", tcase.FloatValue, cast)
			}
		})
	}

}

func TestJsonBoolSlice(t *testing.T) {
	t.Parallel()

	p := parseParameter([]string{"true", "false"})

	slice, canCast := p.([]interface{})
	if !canCast {
		t.Fatalf("expected parameter to be a []interface{}")
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 values, got %d", len(slice))
	}

	results := []bool{true, false}

	for i, res := range results {

		cast, canCast := slice[i].(bool)
		if !canCast {
			t.Errorf("expected parameter %d to be a []bool", i)
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be '%t', got '%t'", res, cast)
			continue
		}

	}

}

func TestBoolSlice(t *testing.T) {
	t.Parallel()

	p := parseParameter([]bool{true, false})

	slice, canCast := p.([]bool)
	if !canCast {
		t.Fatalf("expected parameter to be a []bool")
	}

	if len(slice) != 2 {
		t.Fatalf("expected 2 values, got %d", len(slice))
	}

	results := []bool{true, false}

	for i, res := range results {
		cast := slice[i]
		if cast != res {
			t.Errorf("expected first value to be '%t', got '%t'", res, cast)
			continue
		}

	}

}
