package reqdata

import (
	"math"
	"testing"
)

func TestSimpleString(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: "some-string"}

	err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error: <%s>", err)
		t.FailNow()
	}

	if !p.Parsed {
		t.Errorf("expected parameter to be parsed")
		t.FailNow()
	}

	cast, canCast := p.Value.(string)
	if !canCast {
		t.Errorf("expected parameter to be a string")
		t.FailNow()
	}

	if cast != "some-string" {
		t.Errorf("expected parameter to equal 'some-string', got '%s'", cast)
		t.FailNow()
	}
}

func TestStringSlice(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: `["str1", "str2"]`}

	err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error: <%s>", err)
		t.FailNow()
	}

	if !p.Parsed {
		t.Errorf("expected parameter to be parsed")
		t.FailNow()
	}

	slice, canCast := p.Value.([]interface{})
	if !canCast {
		t.Errorf("expected parameter to be a []interface{}")
		t.FailNow()
	}

	if len(slice) != 2 {
		t.Errorf("expected 2 values, got %d", len(slice))
		t.FailNow()
	}

	results := []string{"str1", "str2"}

	for i, res := range results {

		cast, canCast := slice[i].(string)
		if !canCast {
			t.Errorf("expected parameter %d to be a []string", i)
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be '%s', got '%s'", res, cast)
			continue
		}

	}

}

func TestJsonPrimitiveBool(t *testing.T) {
	tcases := []struct {
		Raw       string
		BoolValue bool
	}{
		{"true", true},
		{"false", false},
	}

	for i, tcase := range tcases {
		t.Run("case "+string(i), func(t *testing.T) {
			p := Parameter{Parsed: false, File: false, Value: tcase.Raw}

			err := p.Parse()
			if err != nil {
				t.Errorf("unexpected error: <%s>", err)
				t.FailNow()
			}

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(bool)
			if !canCast {
				t.Errorf("expected parameter to be a bool")
				t.FailNow()
			}

			if cast != tcase.BoolValue {
				t.Errorf("expected a value of %t, got %t", tcase.BoolValue, cast)
				t.FailNow()
			}
		})
	}

}

func TestJsonPrimitiveFloat(t *testing.T) {
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
		t.Run("case "+string(i), func(t *testing.T) {
			p := Parameter{Parsed: false, File: false, Value: tcase.Raw}

			err := p.Parse()
			if err != nil {
				t.Errorf("unexpected error: <%s>", err)
				t.FailNow()
			}

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(float64)
			if !canCast {
				t.Errorf("expected parameter to be a float64")
				t.FailNow()
			}

			if math.Abs(cast-tcase.FloatValue) > 0.00001 {
				t.Errorf("expected a value of %f, got %f", tcase.FloatValue, cast)
				t.FailNow()
			}
		})
	}

}

func TestOneSliceStringToString(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: []string{"lonely-string"}}

	if err := p.Parse(); err != nil {
		t.Errorf("unexpected error: <%s>", err)
		t.FailNow()
	}

	if !p.Parsed {
		t.Errorf("expected parameter to be parsed")
		t.FailNow()
	}

	cast, canCast := p.Value.(string)
	if !canCast {
		t.Errorf("expected parameter to be a string")
		t.FailNow()
	}

	if cast != "lonely-string" {
		t.Errorf("expected a value of '%s', got '%s'", "lonely-string", cast)
		t.FailNow()
	}
}

func TestOneSliceBoolToBool(t *testing.T) {
	tcases := []struct {
		Raw bool
	}{
		{true},
		{false},
	}

	for i, tcase := range tcases {

		t.Run("case "+string(i), func(t *testing.T) {

			p := Parameter{Parsed: false, File: false, Value: []bool{tcase.Raw}}

			if err := p.Parse(); err != nil {
				t.Errorf("unexpected error: <%s>", err)
				t.FailNow()
			}

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(bool)
			if !canCast {
				t.Errorf("expected parameter to be a bool")
				t.FailNow()
			}

			if cast != tcase.Raw {
				t.Errorf("expected a value of '%t', got '%t'", tcase.Raw, cast)
				t.FailNow()
			}

		})
	}

}

func TestOneSliceJsonBoolToBool(t *testing.T) {
	tcases := []struct {
		Raw       string
		BoolValue bool
	}{
		{"true", true},
		{"false", false},
	}

	for i, tcase := range tcases {

		t.Run("case "+string(i), func(t *testing.T) {

			p := Parameter{Parsed: false, File: false, Value: []string{tcase.Raw}}

			if err := p.Parse(); err != nil {
				t.Errorf("unexpected error: <%s>", err)
				t.FailNow()
			}

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(bool)
			if !canCast {
				t.Errorf("expected parameter to be a bool")
				t.FailNow()
			}

			if cast != tcase.BoolValue {
				t.Errorf("expected a value of '%t', got '%t'", tcase.BoolValue, cast)
				t.FailNow()
			}

		})
	}

}
