package reqdata

import (
	"math"
	"testing"
)

func TestSimpleString(t *testing.T) {
	p := Parameter{Value: "some-string"}
	p.Parse()

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

func TestSimpleFloat(t *testing.T) {
	tcases := []float64{12.3456789, -12.3456789, 0.0000001, -0.0000001}

	for i, tcase := range tcases {
		t.Run("case "+string(i), func(t *testing.T) {
			p := Parameter{Parsed: false, File: false, Value: tcase}
			p.Parse()

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(float64)
			if !canCast {
				t.Errorf("expected parameter to be a float64")
				t.FailNow()
			}

			if math.Abs(cast-tcase) > 0.00000001 {
				t.Errorf("expected parameter to equal '%f', got '%f'", tcase, cast)
				t.FailNow()
			}
		})
	}
}

func TestSimpleBool(t *testing.T) {
	tcases := []bool{true, false}

	for i, tcase := range tcases {
		t.Run("case "+string(i), func(t *testing.T) {
			p := Parameter{Parsed: false, File: false, Value: tcase}

			p.Parse()

			if !p.Parsed {
				t.Errorf("expected parameter to be parsed")
				t.FailNow()
			}

			cast, canCast := p.Value.(bool)
			if !canCast {
				t.Errorf("expected parameter to be a bool")
				t.FailNow()
			}

			if cast != tcase {
				t.Errorf("expected parameter to equal '%t', got '%t'", tcase, cast)
				t.FailNow()
			}
		})
	}
}

func TestJsonStringSlice(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: `["str1", "str2"]`}
	p.Parse()

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

func TestStringSlice(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: []string{"str1", "str2"}}
	p.Parse()

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
			p.Parse()

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
			p.Parse()

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

func TestJsonBoolSlice(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: []string{"true", "false"}}
	p.Parse()

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
	p := Parameter{Parsed: false, File: false, Value: []bool{true, false}}
	p.Parse()

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

	results := []bool{true, false}

	for i, res := range results {

		cast, canCast := slice[i].(bool)
		if !canCast {
			t.Errorf("expected parameter %d to be a bool, got %v", i, slice[i])
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be '%t', got '%t'", res, cast)
			continue
		}

	}

}
